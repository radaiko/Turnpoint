package testutil

import "testing"

func TestLoadAppendixA(t *testing.T) {
	test := LoadAppendixA(t)
	if len(test.Steps) != 9 {
		t.Fatalf("got %d steps, want 9", len(test.Steps))
	}
	if got := test.MaxIntensity(); got != 20 {
		t.Errorf("MaxIntensity = %v, want 20", got)
	}
	if pts := test.FitPoints(false); len(pts) != 8 {
		t.Errorf("fit points = %d, want 8", len(pts))
	}
	if !test.HasAbortedStep() {
		t.Error("expected aborted final step")
	}
}

func TestLoadGoldens(t *testing.T) {
	m := LoadMarkersGolden(t)
	if g := m["OBLA 4.0"]; g.Kmh != 16.1 || g.HR != 167 {
		t.Errorf("OBLA 4.0 golden = %+v", g)
	}
	z := LoadZonesGolden(t)
	if len(z) != 5 || z[1].Zone != "GA1" || z[1].KmhHigh != 11.3 {
		t.Errorf("zone goldens wrong: %+v", z)
	}
}
