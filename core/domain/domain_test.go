package domain

import "testing"

func appendixASteps() []Step {
	// Appendix A: baseline + 8 loaded steps (8 km/h dip, 20 km/h aborted).
	return []Step{
		{Order: 0, Intensity: 0, HeartRate: 0, Lactate: 0.00, HasLactate: true},
		{Order: 1, Intensity: 6, HeartRate: 98, Lactate: 1.24, HasLactate: true},
		{Order: 2, Intensity: 8, HeartRate: 111, Lactate: 1.19, HasLactate: true},
		{Order: 3, Intensity: 10, HeartRate: 120, Lactate: 1.32, HasLactate: true},
		{Order: 4, Intensity: 12, HeartRate: 131, Lactate: 1.66, HasLactate: true},
		{Order: 5, Intensity: 14, HeartRate: 149, Lactate: 2.38, HasLactate: true},
		{Order: 6, Intensity: 16, HeartRate: 166, Lactate: 3.89, HasLactate: true},
		{Order: 7, Intensity: 18, HeartRate: 180, Lactate: 6.66, HasLactate: true},
		{Order: 8, Intensity: 20, HeartRate: 185, Lactate: 7.74, HasLactate: true, Aborted: true},
	}
}

func TestFitPointsExcludesBaselineByDefault(t *testing.T) {
	test := Test{Steps: appendixASteps()}
	pts := test.FitPoints(false)
	if len(pts) != 8 {
		t.Fatalf("got %d fit points, want 8 (baseline excluded)", len(pts))
	}
	if pts[0].X != 6 {
		t.Errorf("first fit point X = %v, want 6", pts[0].X)
	}
	// sorted ascending
	for i := 1; i < len(pts); i++ {
		if pts[i].X < pts[i-1].X {
			t.Fatalf("fit points not sorted at %d", i)
		}
	}
	if got := test.FitPoints(true); len(got) != 9 || got[0].X != 0 {
		t.Errorf("includeBaseline should add the intensity-0 row, got %d", len(got))
	}
}

func TestFitPointsDropsExcludedAndMissing(t *testing.T) {
	steps := appendixASteps()
	steps[3].Excluded = true    // 10 km/h excluded
	steps[4].HasLactate = false // 12 km/h missing lactate
	pts := Test{Steps: steps}.FitPoints(false)
	for _, p := range pts {
		if p.X == 10 || p.X == 12 {
			t.Errorf("excluded/missing step leaked into fit points: %v", p)
		}
	}
	if len(pts) != 6 {
		t.Errorf("got %d, want 6", len(pts))
	}
}

func TestBaselineAndMax(t *testing.T) {
	test := Test{Steps: appendixASteps()}
	if l, ok := test.Baseline(); !ok || l != 0.00 {
		t.Errorf("Baseline = %v,%v want 0,true", l, ok)
	}
	if m := test.MaxIntensity(); m != 20 {
		t.Errorf("MaxIntensity = %v, want 20", m)
	}
	if !test.HasAbortedStep() {
		t.Error("expected aborted step")
	}
}
