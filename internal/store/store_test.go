package store

import (
	"context"
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrateAndSeed(t *testing.T) {
	db := openTemp(t)
	var v int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != 3 {
		t.Errorf("user_version = %d, want 3", v)
	}
	// re-migrate is a no-op
	if err := Migrate(db.DB); err != nil {
		t.Fatalf("re-migrate: %v", err)
	}
	// seed data present
	tpls, err := db.Templates().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tpls) < 2 {
		t.Errorf("expected ≥2 predefined templates, got %d", len(tpls))
	}
	profs, _ := db.Profiles().List(context.Background(), "running")
	if len(profs) != 3 {
		t.Errorf("expected 3 running profiles, got %d", len(profs))
	}
}

func TestForeignKeysOn(t *testing.T) {
	db := openTemp(t)
	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Error("foreign_keys must be ON")
	}
}

func TestAthleteTestStepRoundTripAndCascade(t *testing.T) {
	ctx := context.Background()
	db := openTemp(t)

	bm := 72.5
	aid, err := db.Athletes().Create(ctx, Athlete{Name: "Bogner Markus", Sex: "male", BodyMassKg: &bm})
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.Athletes().Get(ctx, aid)
	if err != nil || got.Name != "Bogner Markus" {
		t.Fatalf("get athlete: %+v %v", got, err)
	}

	// body mass snapshots onto the test (OI-3)
	tid, err := db.Tests().Create(ctx, Test{AthleteID: aid, TestDate: "2025-02-01", Sport: "running",
		StepDurationS: 180, Increment: 2, StartIntensity: 6})
	if err != nil {
		t.Fatal(err)
	}
	tt, _ := db.Tests().Get(ctx, tid)
	if tt.BodyMassSnapshot == nil || *tt.BodyMassSnapshot != 72.5 {
		t.Errorf("body_mass_snapshot = %v, want 72.5", tt.BodyMassSnapshot)
	}

	// editing the athlete's mass must NOT change the existing test (OI-3)
	newBM := 80.0
	_ = db.Athletes().Update(ctx, Athlete{ID: aid, Name: "Bogner Markus", Sex: "male", BodyMassKg: &newBM})
	tt2, _ := db.Tests().Get(ctx, tid)
	if *tt2.BodyMassSnapshot != 72.5 {
		t.Errorf("snapshot changed after athlete edit: %v", *tt2.BodyMassSnapshot)
	}

	// steps round-trip
	lac := 4.0
	hr := 167
	steps := []Step{
		{Intensity: 0, IsBaseline: true},
		{Intensity: 16, HeartRate: &hr, Lactate: &lac},
	}
	if err := db.Steps().ReplaceAll(ctx, tid, steps); err != nil {
		t.Fatal(err)
	}
	gotSteps, _ := db.Steps().ListByTest(ctx, tid)
	if len(gotSteps) != 2 || !gotSteps[0].IsBaseline {
		t.Fatalf("steps round-trip wrong: %+v", gotSteps)
	}

	// cascade: deleting the athlete removes the test and its steps (OI-4)
	if err := db.Athletes().Delete(ctx, aid); err != nil {
		t.Fatal(err)
	}
	if rem, _ := db.Tests().ListByAthlete(ctx, aid); len(rem) != 0 {
		t.Errorf("tests not cascade-deleted: %d remain", len(rem))
	}
	if rem, _ := db.Steps().ListByTest(ctx, tid); len(rem) != 0 {
		t.Errorf("steps not cascade-deleted: %d remain", len(rem))
	}
}

func TestThresholdAndZoneSnapshots(t *testing.T) {
	ctx := context.Background()
	db := openTemp(t)
	aid, _ := db.Athletes().Create(ctx, Athlete{Name: "A", Sex: "unspecified"})
	tid, _ := db.Tests().Create(ctx, Test{AthleteID: aid, TestDate: "2025-02-01", Sport: "running",
		StepDurationS: 180, Increment: 2, StartIntensity: 6})

	i, hr := 16.1, 167.0
	if err := db.Thresholds().ReplaceAll(ctx, tid, []ThresholdResult{
		{Method: "IANS", Intensity: &i, HeartRate: &hr, FitType: "spline"},
	}); err != nil {
		t.Fatal(err)
	}
	trs, _ := db.Thresholds().ListByTest(ctx, tid)
	if len(trs) != 1 || *trs[0].Intensity != 16.1 {
		t.Fatalf("threshold round-trip: %+v", trs)
	}

	lo, hi := 7.4, 11.3
	if err := db.ZonesRepo().ReplaceAll(ctx, tid, []Zone{
		{ZoneIndex: 2, ZoneName: "GA1", IntensityLow: &lo, IntensityHigh: &hi},
	}); err != nil {
		t.Fatal(err)
	}
	zs, _ := db.ZonesRepo().ListByTest(ctx, tid)
	if len(zs) != 1 || zs[0].ZoneName != "GA1" {
		t.Fatalf("zone round-trip: %+v", zs)
	}
}

func TestReportGlobalDefault(t *testing.T) {
	db := openTemp(t)
	rs, err := db.Reports().GetGlobal(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rs.PageSize != "A4" || rs.BlockConfigJSON == "" {
		t.Errorf("global report defaults wrong: %+v", rs)
	}
}
