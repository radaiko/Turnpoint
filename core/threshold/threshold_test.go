package threshold

import (
	"math"
	"testing"

	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
)

func points() []fit.Point {
	return []fit.Point{{6, 1.24}, {8, 1.19}, {10, 1.32}, {12, 1.66}, {14, 2.38}, {16, 3.89}, {18, 6.66}, {20, 7.74}}
}

func steps() []domain.Step {
	hr := []int{98, 111, 120, 131, 149, 166, 180, 185}
	var ss []domain.Step
	for i, p := range points() {
		ss = append(ss, domain.Step{Intensity: p.X, HeartRate: hr[i], Lactate: p.Y, HasLactate: true})
	}
	return ss
}

func ctx() Context {
	return Context{Points: points(), Steps: steps(), BaselineLactate: 1.19, HasBaseline: false}
}

func poly3(t *testing.T) fit.Fit {
	t.Helper()
	f, err := fit.Poly(points(), 3)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func spline(t *testing.T) fit.Fit {
	t.Helper()
	f, err := fit.Spline(points(), 0)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// OBLA 4.0 is the binding P0 parity anchor (V2): 16.1 km/h on the spline.
func TestOBLA4Anchor(t *testing.T) {
	r := oblaMethod{OBLA4, 4.0}.Compute(spline(t), ctx())
	if !r.Computable {
		t.Fatalf("OBLA 4.0 not computable: %s", r.Reason)
	}
	if math.Abs(r.Intensity-16.1) > 0.1 {
		t.Errorf("OBLA 4.0 = %.3f km/h, want 16.1 ±0.1", r.Intensity)
	}
}

func TestOBLA2And6(t *testing.T) {
	f := spline(t)
	// OBLA 6.0 (17.43) sits within the OI-1 1%-relative bound of 17.6 (0.176).
	if r := (oblaMethod{OBLA2, 2.0}).Compute(f, ctx()); math.Abs(r.Intensity-13.1) > 0.1 {
		t.Errorf("OBLA 2.0 = %.3f, want ≈13.1", r.Intensity)
	}
	if r := (oblaMethod{OBLA6, 6.0}).Compute(f, ctx()); math.Abs(r.Intensity-17.6) > 0.176 {
		t.Errorf("OBLA 6.0 = %.3f, want ≈17.6 (±1%%)", r.Intensity)
	}
}

func TestMAX(t *testing.T) {
	r := maxMethod{}.Compute(nil, ctx())
	if !r.Computable || r.Intensity != 20 || math.Abs(r.Lactate-7.74) > 1e-9 {
		t.Errorf("MAX = %+v, want 20/7.74", r)
	}
	if r.FitKind != fit.KindNone {
		t.Errorf("MAX FitKind = %v, want none", r.FitKind)
	}
}

func TestModDmaxStartsAt14(t *testing.T) {
	// the first >0.4 rise on Appendix A is 12→14 ⇒ chord starts at 14
	if v, ok := firstRise(points(), 0.4); !ok || v != 14 {
		t.Errorf("firstRise = %v,%v want 14,true", v, ok)
	}
	r := modDmaxMethod{}.Compute(poly3(t), ctx())
	if !r.Computable || r.Intensity < 14 || r.Intensity > 20 {
		t.Errorf("ModDmax = %+v, want a value in (14,20]", r)
	}
}

func TestIATHandComputed(t *testing.T) {
	// L_min ≈ 1.19 near 8 km/h; target ≈ 2.69 ⇒ ≈14–15 km/h
	r := iatMethod{}.Compute(poly3(t), ctx())
	if !r.Computable {
		t.Fatalf("IAT not computable: %s", r.Reason)
	}
	if r.Intensity < 13 || r.Intensity > 16 {
		t.Errorf("IAT = %.2f, want ≈14–15", r.Intensity)
	}
}

// Every default method must produce a computable result on Appendix A and carry
// the fit kind it declares.
func TestAllMethodsComputable(t *testing.T) {
	c := ctx()
	for _, m := range Default() {
		var f fit.Fit
		var err error
		switch m.RequiredFit() {
		case fit.KindNone:
			f = nil
		default:
			f, err = fit.New(m.RequiredFit(), points())
			if err != nil {
				t.Fatalf("%s: build %v fit: %v", m.Marker(), m.RequiredFit(), err)
			}
		}
		r := m.Compute(f, c)
		if !r.Computable {
			t.Errorf("%s not computable: %s", m.Marker(), r.Reason)
			continue
		}
		if r.Intensity < 5 || r.Intensity > 21 {
			t.Errorf("%s intensity %.2f out of plausible range", m.Marker(), r.Intensity)
		}
		if m.Marker() != MAX && r.FitKind != m.RequiredFit() {
			t.Errorf("%s FitKind = %v, want %v", m.Marker(), r.FitKind, m.RequiredFit())
		}
	}
}

func TestForFiltersMarkers(t *testing.T) {
	got := For(OBLA4, MAX)
	if len(got) != 2 {
		t.Fatalf("For returned %d, want 2", len(got))
	}
}
