package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/radaiko/turnpoint/core/analysis"
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/threshold"
	"github.com/radaiko/turnpoint/core/unit"
	"github.com/radaiko/turnpoint/core/zone"
	"github.com/radaiko/turnpoint/internal/backup"
	"github.com/radaiko/turnpoint/internal/csvio"
	"github.com/radaiko/turnpoint/internal/service"
	"github.com/radaiko/turnpoint/internal/store"
)

// App is the Wails-bound facade: its exported methods are callable from the
// Svelte frontend. It orchestrates the core analysis engine and the SQLite store.
type App struct {
	ctx   context.Context
	db    *store.DB
	mu    sync.Mutex
	cache map[int64]cachedAnalysis // last full analysis per test (for the drag fast path)
}

type cachedAnalysis struct {
	res analysis.Result
	cfg analysis.Config
	dt  domain.Test
}

// NewApp constructs the app.
func NewApp() *App { return &App{cache: map[int64]cachedAnalysis{}} }

// startup opens (and migrates) the database under the OS app-data directory.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	dir = filepath.Join(dir, "Turnpoint")
	_ = os.MkdirAll(dir, 0o755)
	db, err := store.Open(filepath.Join(dir, "turnpoint.db"))
	if err != nil {
		fmt.Println("turnpoint: failed to open database:", err)
		return
	}
	a.db = db
}

// shutdown closes the database.
func (a *App) shutdown(context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

// ── Athletes (FR-A) ─────────────────────────────────────────────────────────

func (a *App) ListAthletes(search string) ([]store.AthleteSummary, error) {
	return a.db.Athletes().List(a.ctx, search)
}
func (a *App) GetAthlete(id int64) (store.Athlete, error) { return a.db.Athletes().Get(a.ctx, id) }

// SaveAthlete creates (id==0) or updates an athlete, returning its id.
func (a *App) SaveAthlete(in store.Athlete) (int64, error) {
	if in.ID == 0 {
		return a.db.Athletes().Create(a.ctx, in)
	}
	return in.ID, a.db.Athletes().Update(a.ctx, in)
}
func (a *App) DeleteAthlete(id int64) error { return a.db.Athletes().Delete(a.ctx, id) }

// ── Tests & steps (FR-T) ────────────────────────────────────────────────────

func (a *App) ListTests(athleteID int64) ([]store.Test, error) {
	return a.db.Tests().ListByAthlete(a.ctx, athleteID)
}
func (a *App) GetTest(id int64) (store.Test, error) { return a.db.Tests().Get(a.ctx, id) }

func (a *App) SaveTest(in store.Test) (int64, error) {
	if in.ID == 0 {
		return a.db.Tests().Create(a.ctx, in)
	}
	return in.ID, a.db.Tests().Update(a.ctx, in)
}
func (a *App) DeleteTest(id int64) error { return a.db.Tests().Delete(a.ctx, id) }

func (a *App) GetSteps(testID int64) ([]store.Step, error) {
	return a.db.Steps().ListByTest(a.ctx, testID)
}
func (a *App) SaveSteps(testID int64, steps []store.Step) error {
	return a.db.Steps().ReplaceAll(a.ctx, testID, steps)
}

// ── Catalog (FR-T7/T8, FR-Z5) ───────────────────────────────────────────────

func (a *App) ListTemplates() ([]store.Template, error)             { return a.db.Templates().List(a.ctx) }
func (a *App) ListProfiles(sport string) ([]store.TrainingProfile, error) {
	return a.db.Profiles().List(a.ctx, sport)
}

// ── Analysis (FR-D/Z/C) ─────────────────────────────────────────────────────

// Analyze runs the full pipeline for a test, persists the snapshot rows, caches
// the result for the drag fast path, and returns the frontend payload.
func (a *App) Analyze(testID int64) (service.AnalysisDTO, error) {
	test, err := a.db.Tests().Get(a.ctx, testID)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	steps, err := a.db.Steps().ListByTest(a.ctx, testID)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	dt := service.StoreToDomain(test, steps)
	cfg := configForSport(test.Sport)
	res, err := analysis.Analyze(analysis.Input{Test: dt}, cfg)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	_ = a.db.Thresholds().ReplaceAll(a.ctx, testID, service.ThresholdSnapshotRows(res))
	_ = a.db.ZonesRepo().ReplaceAll(a.ctx, testID, service.ZoneSnapshotRows(res))

	a.mu.Lock()
	a.cache[testID] = cachedAnalysis{res: res, cfg: cfg, dt: dt}
	a.mu.Unlock()
	return service.BuildAnalysisDTO(res, dt), nil
}

// RecomputeZones is the marker-drag fast path (FR-C2). It reuses the cached fits.
func (a *App) RecomputeZones(testID int64, lt1, lt2 float64) (service.AnalysisDTO, error) {
	a.mu.Lock()
	c, ok := a.cache[testID]
	a.mu.Unlock()
	if !ok {
		return a.Analyze(testID)
	}
	res, err := analysis.RecomputeZones(c.res, analysis.Input{Test: c.dt}, c.cfg, lt1, lt2)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	a.mu.Lock()
	a.cache[testID] = cachedAnalysis{res: res, cfg: c.cfg, dt: c.dt}
	a.mu.Unlock()
	_ = a.db.ZonesRepo().ReplaceAll(a.ctx, testID, service.ZoneSnapshotRows(res))
	return service.BuildAnalysisDTO(res, c.dt), nil
}

// ── Import / export / backup (FR-M, FR-R5) ──────────────────────────────────

func (a *App) ImportCSV(testID int64, text string) (csvio.ImportReport, error) {
	steps, rep, err := csvio.ParseSteps(strings.NewReader(text), csvio.Options{})
	if err != nil {
		return rep, err
	}
	if err := a.db.Steps().ReplaceAll(a.ctx, testID, steps); err != nil {
		return rep, err
	}
	return rep, nil
}

// ParsePaste parses clipboard text into steps without persisting (grid paste, OI-10).
func (a *App) ParsePaste(text string) ([]store.Step, error) {
	steps, _, err := csvio.ParseSteps(strings.NewReader(text), csvio.Options{})
	return steps, err
}

func (a *App) ExportCSV(testID int64) (string, error) {
	test, err := a.db.Tests().Get(a.ctx, testID)
	if err != nil {
		return "", err
	}
	steps, err := a.db.Steps().ListByTest(a.ctx, testID)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := csvio.WriteSteps(&sb, steps, test.Sport, csvio.Options{}); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (a *App) Backup(destPath string) error { return backup.Backup(a.db.DB, destPath) }

// ── helpers ─────────────────────────────────────────────────────────────────

// configForSport returns the default analysis config, swapping in a cycling
// profile when the test is a bike test.
func configForSport(sport string) analysis.Config {
	cfg := analysis.DefaultConfig()
	if service.ParseSport(sport) == unit.Cycling {
		for _, p := range zone.Predefined() {
			if p.Sport == unit.Cycling && p.Level == "Leistung" {
				cfg.Profile = p
				break
			}
		}
	}
	return cfg
}

var _ = threshold.OBLA4 // keep threshold import referenced for future config DTOs
