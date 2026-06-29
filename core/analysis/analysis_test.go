package analysis

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/internal/testutil"
	"github.com/radaiko/turnpoint/core/threshold"
)

// V2 binding parity: the full pipeline on Appendix A reproduces the Appendix C
// marker table and zone table within the OI-1 tolerance.
func TestAnalyzeReproducesAppendixC(t *testing.T) {
	in := Input{Test: testutil.LoadAppendixA(t)}
	res, err := Analyze(in, DefaultConfig())
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	gold := testutil.LoadMarkersGolden(t)

	check := func(m threshold.Marker, goldName string) {
		dm, ok := res.Markers[m]
		g := gold[goldName]
		if !ok {
			t.Errorf("%s missing from markers", goldName)
			return
		}
		testutil.AssertIntensityRun(t, goldName, dm.Intensity, g.Kmh)
		testutil.AssertHR(t, goldName, float64(dm.HeartRate), g.HR)
	}
	check(threshold.OBLA2, "OBLA 2.0")
	check(threshold.OBLA4, "OBLA 4.0")
	check(threshold.OBLA6, "OBLA 6.0")
	check(threshold.MAX, "MAX")

	// IANS ← OBLA 4.0 must equal 16.1 km/h / 167 bpm field-for-field (binding).
	testutil.AssertIntensityRun(t, "IANS", res.LT2.Intensity, 16.1)
	testutil.AssertHR(t, "IANS", float64(res.LT2.Metrics.HeartRate), 167)
	if res.LT2.Marker != threshold.OBLA4 {
		t.Errorf("LT2 anchor = %v, want OBLA4", res.LT2.Marker)
	}
	if res.LT2.Manual {
		t.Error("LT2 should be algorithmic, not manual")
	}

	// Zones reproduce Appendix C intensity bounds.
	zg := testutil.LoadZonesGolden(t)
	if len(res.Zones) != 5 {
		t.Fatalf("got %d zones, want 5", len(res.Zones))
	}
	for i, z := range res.Zones {
		if !testutil.EqualIntensityRun(z.IntensityLow, zg[i].KmhLow) || !testutil.EqualIntensityRun(z.IntensityHigh, zg[i].KmhHigh) {
			t.Errorf("zone %s = %.2f–%.2f, want %.1f–%.1f", z.Label, z.IntensityLow, z.IntensityHigh, zg[i].KmhLow, zg[i].KmhHigh)
		}
	}
}

// NFR-6: identical input ⇒ byte-identical JSON.
func TestAnalyzeDeterministic(t *testing.T) {
	in := Input{Test: testutil.LoadAppendixA(t)}
	a, _ := Analyze(in, DefaultConfig())
	b, _ := Analyze(in, DefaultConfig())
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Error("Analyze is not deterministic")
	}
}

// V4 edge: <4 fit steps blocks analysis.
func TestInsufficientSteps(t *testing.T) {
	test := domain.Test{Steps: []domain.Step{
		{Intensity: 6, Lactate: 1.2, HasLactate: true, HeartRate: 100},
		{Intensity: 8, Lactate: 1.5, HasLactate: true, HeartRate: 110},
		{Intensity: 10, Lactate: 2.0, HasLactate: true, HeartRate: 120},
	}}
	if _, err := Analyze(Input{Test: test}, DefaultConfig()); err != ErrInsufficientSteps {
		t.Errorf("err = %v, want ErrInsufficientSteps", err)
	}
}

// V4 edge: exactly 4 steps runs but warns.
func TestFewStepsWarns(t *testing.T) {
	test := domain.Test{Steps: []domain.Step{
		{Intensity: 6, Lactate: 1.2, HasLactate: true, HeartRate: 100},
		{Intensity: 8, Lactate: 1.5, HasLactate: true, HeartRate: 110},
		{Intensity: 10, Lactate: 2.5, HasLactate: true, HeartRate: 130},
		{Intensity: 12, Lactate: 4.5, HasLactate: true, HeartRate: 150},
	}}
	res, err := Analyze(Input{Test: test}, DefaultConfig())
	if err != nil {
		t.Fatalf("4 steps should analyse: %v", err)
	}
	if !hasWarn(res.Warnings, domain.WarnFewSteps) {
		t.Error("expected WarnFewSteps")
	}
}

// V4 edge: pinned-fit methods are independent of the display fit (FR-D4).
func TestPinnedFitsIgnoreDisplayChoice(t *testing.T) {
	in := Input{Test: testutil.LoadAppendixA(t)}
	cfg := DefaultConfig()
	a, _ := Analyze(in, cfg)
	cfg.DisplayFit = fit.KindSpline
	b, _ := Analyze(in, cfg)
	// OBLA 4.0 pins the spline either way → identical.
	if math.Abs(a.Markers[threshold.OBLA4].Intensity-b.Markers[threshold.OBLA4].Intensity) > 1e-9 {
		t.Error("OBLA 4.0 changed with display fit, but it pins its own fit")
	}
}

// FR-C2 fast path equals a full analyze for the same anchors.
func TestRecomputeZonesMatchesFull(t *testing.T) {
	in := Input{Test: testutil.LoadAppendixA(t)}
	cfg := DefaultConfig()
	full, _ := Analyze(in, cfg)
	fast, err := RecomputeZones(full, in, cfg, full.LT1.Intensity, full.LT2.Intensity)
	if err != nil {
		t.Fatal(err)
	}
	for i := range full.Zones {
		if math.Abs(full.Zones[i].IntensityLow-fast.Zones[i].IntensityLow) > 1e-9 ||
			math.Abs(full.Zones[i].IntensityHigh-fast.Zones[i].IntensityHigh) > 1e-9 {
			t.Errorf("zone %d differs between full and fast path", i)
		}
	}
}

// review #3: per-method params (custom OBLA concentration) change the result.
func TestMethodParamsApplied(t *testing.T) {
	in := Input{Test: testutil.LoadAppendixA(t)}
	cfg := DefaultConfig()
	a, _ := Analyze(in, cfg)
	cfg.MethodParams = map[threshold.Marker]threshold.Params{threshold.OBLA4: {OBLAConc: 2.0}}
	b, _ := Analyze(in, cfg)
	// OBLA4 with a 2.0 override should now match the 4.0-default's 2.0 marker position.
	if math.Abs(a.Markers[threshold.OBLA4].Intensity-b.Markers[threshold.OBLA4].Intensity) < 0.5 {
		t.Errorf("custom OBLAConc had no effect: %.2f vs %.2f", a.Markers[threshold.OBLA4].Intensity, b.Markers[threshold.OBLA4].Intensity)
	}
}

// review #2: a non-computable IANS anchor warns and yields no (zero-width) zones.
func TestNonComputableAnchorSkipsZones(t *testing.T) {
	// an easy test that never reaches 4 mmol/L → OBLA4 (default IANS) not computable
	test := domain.Test{Steps: []domain.Step{
		{Intensity: 6, Lactate: 1.0, HasLactate: true, HeartRate: 100},
		{Intensity: 8, Lactate: 1.1, HasLactate: true, HeartRate: 110},
		{Intensity: 10, Lactate: 1.2, HasLactate: true, HeartRate: 120},
		{Intensity: 12, Lactate: 1.3, HasLactate: true, HeartRate: 130},
		{Intensity: 14, Lactate: 1.5, HasLactate: true, HeartRate: 140},
	}}
	res, err := Analyze(Input{Test: test}, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Zones) != 0 {
		t.Errorf("expected no zones for non-computable IANS, got %d", len(res.Zones))
	}
	if !hasWarn(res.Warnings, domain.WarnMethodNotComputable) {
		t.Error("expected a not-computable warning for the IANS anchor/zones")
	}
}

func hasWarn(ws []domain.Warning, code domain.WarnCode) bool {
	for _, w := range ws {
		if w.Code == code {
			return true
		}
	}
	return false
}
