package zone

import (
	"testing"

	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/metrics"
	"github.com/radaiko/turnpoint/core/unit"
)

func appendixAPoints() []fit.Point {
	return []fit.Point{{6, 1.24}, {8, 1.19}, {10, 1.32}, {12, 1.66}, {14, 2.38}, {16, 3.89}, {18, 6.66}, {20, 7.74}}
}

func appendixASteps() []domain.Step {
	hr := []int{98, 111, 120, 131, 149, 166, 180, 185}
	pts := appendixAPoints()
	steps := []domain.Step{{Intensity: 0, HeartRate: 0, Lactate: 0, HasLactate: true}}
	for i, p := range pts {
		steps = append(steps, domain.Step{Intensity: p.X, HeartRate: hr[i], Lactate: p.Y, HasLactate: true})
	}
	return steps
}

func TestLaufLeistung6ReproducesAppendixCZones(t *testing.T) {
	const ians = 16.1
	f, err := fit.Poly(appendixAPoints(), 3)
	if err != nil {
		t.Fatal(err)
	}
	hr := metrics.NewHRCurve(appendixASteps())
	zones := Derive(LaufLeistung6(), 10.5, ians, f, hr, unit.Running)
	if len(zones) != 5 {
		t.Fatalf("got %d zones, want 5", len(zones))
	}
	want := []struct {
		idx    Index
		lo, hi float64
		label  string
	}{
		{REKOM, 0.0, 7.4, "REKOM"},
		{GA1, 7.4, 11.3, "GA1"},
		{GA2, 11.3, 14.2, "GA2"},
		{EB, 14.2, 16.4, "EB"},
		{SB, 16.4, 20.1, "SB"},
	}
	for i, w := range want {
		z := zones[i]
		if z.Label != w.label {
			t.Errorf("zone %d label = %q, want %q", i, z.Label, w.label)
		}
		if abs(z.IntensityLow-w.lo) > 0.1 || abs(z.IntensityHigh-w.hi) > 0.1 {
			t.Errorf("%s intensity = %.2f–%.2f, want %.1f–%.1f", w.label, z.IntensityLow, z.IntensityHigh, w.lo, w.hi)
		}
	}
	// GA1 upper pace ≈ 05:19 at 11.27 km/h (Appendix C).
	if got := zones[1].PaceHigh.MMSS(); got != "05:19" {
		t.Errorf("GA1 pace high = %s, want 05:19", got)
	}
}

func TestPredefinedProfiles(t *testing.T) {
	ps := Predefined()
	if len(ps) != 6 {
		t.Fatalf("got %d profiles, want 6", len(ps))
	}
	if !ps[0].Calibrated {
		t.Error("first profile should be the calibrated reference")
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
