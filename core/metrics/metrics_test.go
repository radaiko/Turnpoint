package metrics

import (
	"testing"

	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/unit"
)

func appendixASteps() []domain.Step {
	return []domain.Step{
		{Intensity: 0, HeartRate: 0, Lactate: 0, HasLactate: true},
		{Intensity: 6, HeartRate: 98, Lactate: 1.24, HasLactate: true},
		{Intensity: 8, HeartRate: 111, Lactate: 1.19, HasLactate: true},
		{Intensity: 10, HeartRate: 120, Lactate: 1.32, HasLactate: true},
		{Intensity: 12, HeartRate: 131, Lactate: 1.66, HasLactate: true},
		{Intensity: 14, HeartRate: 149, Lactate: 2.38, HasLactate: true},
		{Intensity: 16, HeartRate: 166, Lactate: 3.89, HasLactate: true},
		{Intensity: 18, HeartRate: 180, Lactate: 6.66, HasLactate: true},
		{Intensity: 20, HeartRate: 185, Lactate: 7.74, HasLactate: true},
	}
}

func TestHRCurveInterpolation(t *testing.T) {
	c := NewHRCurve(appendixASteps())
	if !c.Valid() {
		t.Fatal("curve invalid")
	}
	if got := c.At(16.1); got != 167 { // 166 + 0.05*(180-166) = 166.7 → 167
		t.Errorf("HR@16.1 = %d, want 167", got)
	}
	if got := c.At(16.0); got != 166 {
		t.Errorf("HR@16 = %d, want 166", got)
	}
	// clamp beyond range
	if got := c.At(25); got != 185 {
		t.Errorf("HR@25 (clamped) = %d, want 185", got)
	}
}

func TestPctMaxAndPace(t *testing.T) {
	if got := PctMax(16.1, 20); got < 80.49 || got > 80.51 {
		t.Errorf("PctMax = %v, want 80.5", got)
	}
	dm := Derive(16.1, 20, unit.Running, NewHRCurve(appendixASteps()), 0, 4.0)
	if !dm.HasPace || dm.Pace.MMSS() != "03:43" {
		t.Errorf("pace = %q, want 03:43", dm.Pace.MMSS())
	}
	if dm.HasKcal {
		t.Error("kcal should be disabled when body mass is 0")
	}
}

func TestKcalWithBodyMass(t *testing.T) {
	dm := Derive(16.1, 20, unit.Running, NewHRCurve(appendixASteps()), 70, 4.0)
	if !dm.HasKcal || dm.KcalPerHour <= 0 {
		t.Errorf("expected kcal with body mass, got %+v", dm)
	}
	cyc := Derive(300, 400, unit.Cycling, HRCurve{}, 70, 4.0)
	if cyc.HasPace {
		t.Error("cycling should have no pace")
	}
	if cyc.KcalPerHour != 300*3.6 {
		t.Errorf("cycling kcal = %v, want %v", cyc.KcalPerHour, 300*3.6)
	}
}
