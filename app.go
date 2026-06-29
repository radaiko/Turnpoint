package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/radaiko/turnpoint/core/analysis"
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/threshold"
	"github.com/radaiko/turnpoint/internal/backup"
	"github.com/radaiko/turnpoint/internal/csvio"
	"github.com/radaiko/turnpoint/internal/service"
	"github.com/radaiko/turnpoint/internal/store"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

func (a *App) ListTemplates() ([]store.Template, error) { return a.db.Templates().List(a.ctx) }
func (a *App) ListProfiles(sport string) ([]store.TrainingProfile, error) {
	return a.db.Profiles().List(a.ctx, sport)
}

// SaveTemplate creates (id==0) or updates a user template (FR-T8). Predefined
// templates are read-only and rejected on update.
func (a *App) SaveTemplate(t store.Template) (int64, error) {
	if t.ID == 0 {
		return a.db.Templates().Create(a.ctx, t)
	}
	return t.ID, a.db.Templates().Update(a.ctx, t)
}

// DeleteTemplate removes a user template (predefined are protected).
func (a *App) DeleteTemplate(id int64) error { return a.db.Templates().Delete(a.ctx, id) }

// ── Analysis (FR-D/Z/C) ─────────────────────────────────────────────────────

// Analyze runs the full pipeline for a test using its persisted (or default)
// configuration, persists the snapshot rows, caches the result for the drag fast
// path, and returns the frontend payload.
func (a *App) Analyze(testID int64) (service.AnalysisDTO, error) {
	dto, err := a.GetAnalysisConfig(testID)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	return a.analyzeInternal(testID, dto)
}

// AnalyzeWith runs the pipeline with an explicit configuration and persists it as
// this test's analysis config (FR-D2/F2/Z2/Z5).
func (a *App) AnalyzeWith(testID int64, cfg service.AnalysisConfigDTO) (service.AnalysisDTO, error) {
	if j, err := json.Marshal(cfg); err == nil {
		_ = a.db.Configs().Upsert(a.ctx, testID, string(j))
	}
	return a.analyzeInternal(testID, cfg)
}

// GetAnalysisConfig returns the persisted config for a test, or the sport default.
func (a *App) GetAnalysisConfig(testID int64) (service.AnalysisConfigDTO, error) {
	test, err := a.db.Tests().Get(a.ctx, testID)
	if err != nil {
		return service.AnalysisConfigDTO{}, err
	}
	if j, ok, _ := a.db.Configs().Get(a.ctx, testID); ok {
		var dto service.AnalysisConfigDTO
		if err := json.Unmarshal([]byte(j), &dto); err == nil {
			return dto, nil
		}
	}
	return service.DefaultConfigDTO(test.Sport), nil
}

// ResetAnalysisConfig discards a test's saved config and reverts to the default.
func (a *App) ResetAnalysisConfig(testID int64) (service.AnalysisDTO, error) {
	test, err := a.db.Tests().Get(a.ctx, testID)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	return a.AnalyzeWith(testID, service.DefaultConfigDTO(test.Sport))
}

// GetMarkerOptions lists all markers and their required fit (FR-D2/D4).
func (a *App) GetMarkerOptions() []service.MarkerOption { return service.AllMarkerOptions() }

// GetProfileOptions lists predefined training profiles for a sport (FR-Z5).
func (a *App) GetProfileOptions(sport string) []service.ProfileOption {
	return service.ProfileOptionsForSport(sport)
}

func (a *App) analyzeInternal(testID int64, dto service.AnalysisConfigDTO) (service.AnalysisDTO, error) {
	test, err := a.db.Tests().Get(a.ctx, testID)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	steps, err := a.db.Steps().ListByTest(a.ctx, testID)
	if err != nil {
		return service.AnalysisDTO{}, err
	}
	dt := service.StoreToDomain(test, steps)
	cfg := service.ToConfig(dto)
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

// BackupDatabase opens a save dialog and writes a consistent backup (FR-M3).
// Returns the chosen path ("" if cancelled).
func (a *App) BackupDatabase() (string, error) {
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Back up database",
		DefaultFilename: "turnpoint-backup.db",
		Filters:         []wruntime.FileFilter{{DisplayName: "Turnpoint database", Pattern: "*.db"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := backup.Backup(a.db.DB, path); err != nil {
		return "", err
	}
	return path, nil
}

// RestoreDatabase opens a file dialog, validates and restores the chosen backup,
// replacing the live database (FR-M3). Returns the chosen path ("" if cancelled).
func (a *App) RestoreDatabase() (string, error) {
	src, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:   "Restore database from backup",
		Filters: []wruntime.FileFilter{{DisplayName: "Turnpoint database", Pattern: "*.db"}},
	})
	if err != nil || src == "" {
		return "", err
	}
	dbPath := a.db.Path()
	a.db.Close()
	rErr := backup.Restore(dbPath, src)
	reopened, err := store.Open(dbPath) // reopens restored DB, or the original on failure
	if err != nil {
		return "", err
	}
	a.db = reopened
	a.mu.Lock()
	a.cache = map[int64]cachedAnalysis{}
	a.mu.Unlock()
	return src, rErr
}

var _ = threshold.OBLA4 // keep threshold import referenced
