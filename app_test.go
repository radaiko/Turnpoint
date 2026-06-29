package main

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"github.com/radaiko/turnpoint/internal/store"
)

// newTestApp wires the App facade to a throwaway database, bypassing the Wails
// runtime so the exact methods the frontend calls can be exercised headlessly.
func newTestApp(t *testing.T) *App {
	t.Helper()
	app := NewApp()
	db, err := store.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	app.db = db
	app.ctx = context.Background()
	t.Cleanup(func() { db.Close() })
	return app
}

func p[T any](v T) *T { return &v }

// appendixASteps returns the SRS Appendix A running test as store rows.
func appendixASteps(testID int64) []store.Step {
	type r struct {
		i, lac float64
		t, hr  int
		abort  bool
	}
	rows := []r{
		{0, 0.00, 0, 0, false},
		{6, 1.24, 180, 98, false},
		{8, 1.19, 360, 111, false},
		{10, 1.32, 540, 120, false},
		{12, 1.66, 720, 131, false},
		{14, 2.38, 900, 149, false},
		{16, 3.89, 1080, 166, false},
		{18, 6.66, 1260, 180, false},
		{20, 7.74, 1330, 185, true},
	}
	var steps []store.Step
	for i, row := range rows {
		steps = append(steps, store.Step{
			TestID: testID, StepOrder: i, Intensity: row.i,
			TimePointS: p(row.t), HeartRate: p(row.hr), Lactate: p(row.lac),
			IsBaseline: row.i == 0, Aborted: row.abort,
		})
	}
	return steps
}

// TestAppHappyPath drives the full bound-method workflow end-to-end and asserts
// the analysis reproduces the Appendix C reference (V5 happy path).
func TestAppHappyPath(t *testing.T) {
	app := newTestApp(t)

	aid, err := app.SaveAthlete(store.Athlete{Name: "Bogner Markus", Sex: "male", BodyMassKg: p(72.0)})
	if err != nil {
		t.Fatal(err)
	}
	if list, _ := app.ListAthletes(""); len(list) != 1 {
		t.Fatalf("ListAthletes = %d, want 1", len(list))
	}

	tid, err := app.SaveTest(store.Test{
		AthleteID: aid, TestDate: "2025-02-01", Sport: "running",
		StepDurationS: 180, Increment: 2, StartIntensity: 6,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveSteps(tid, appendixASteps(tid)); err != nil {
		t.Fatal(err)
	}

	dto, err := app.Analyze(tid)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// IANS (LT2 ← OBLA 4.0) reproduces 16.1 km/h / 167 bpm (Appendix C).
	if math.Abs(dto.LT2.Intensity-16.1) > 0.1 {
		t.Errorf("IANS = %.2f km/h, want 16.1", dto.LT2.Intensity)
	}
	if dto.LT2.HeartRate != 167 {
		t.Errorf("IANS HR = %d, want 167", dto.LT2.HeartRate)
	}
	if len(dto.Zones) != 5 {
		t.Errorf("zones = %d, want 5", len(dto.Zones))
	}
	if len(dto.Curve) == 0 || len(dto.RawPoints) != 8 {
		t.Errorf("chart series wrong: curve=%d raw=%d", len(dto.Curve), len(dto.RawPoints))
	}

	// snapshots persisted
	if trs, _ := app.db.Thresholds().ListByTest(app.ctx, tid); len(trs) == 0 {
		t.Error("threshold snapshot not persisted")
	}
	if zs, _ := app.db.ZonesRepo().ListByTest(app.ctx, tid); len(zs) != 5 {
		t.Errorf("zone snapshot = %d, want 5", len(zs))
	}

	// drag fast path recomputes zones for a new IANS
	dragged, err := app.RecomputeZones(tid, dto.LT1.Intensity, 17.0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(dragged.LT2.Intensity-17.0) > 1e-9 {
		t.Errorf("dragged IANS = %.2f, want 17.0", dragged.LT2.Intensity)
	}

	// CSV export round-trips through the grid
	csv, err := app.ExportCSV(tid)
	if err != nil || len(csv) == 0 {
		t.Fatalf("ExportCSV: %v", err)
	}
	rep, err := app.ImportCSV(tid, csv)
	if err != nil || rep.Imported == 0 {
		t.Fatalf("ImportCSV: %v %+v", err, rep)
	}
}
