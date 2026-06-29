package fit

import (
	"math"
	"testing"
)

// cubicPoints samples L(x) = 1 + 0.5x - 0.05x² + 0.002x³ exactly.
func cubicPoints() ([]Point, func(float64) float64) {
	f := func(x float64) float64 { return 1 + 0.5*x - 0.05*x*x + 0.002*x*x*x }
	var pts []Point
	for x := 6.0; x <= 20; x += 2 {
		pts = append(pts, Point{X: x, Y: f(x)})
	}
	return pts, f
}

func TestPolyRecoversCubic(t *testing.T) {
	pts, f := cubicPoints()
	p, err := Poly(pts, 3)
	if err != nil {
		t.Fatal(err)
	}
	for x := 6.0; x <= 20; x += 1 {
		if math.Abs(p.Predict(x)-f(x)) > 1e-6 {
			t.Errorf("Predict(%v)=%v, want %v", x, p.Predict(x), f(x))
		}
	}
	// analytic derivative vs finite difference
	const h = 1e-5
	for _, x := range []float64{8, 12, 16} {
		fd := (f(x+h) - f(x-h)) / (2 * h)
		if math.Abs(p.Derivative(x)-fd) > 1e-4 {
			t.Errorf("Derivative(%v)=%v, want ≈%v", x, p.Derivative(x), fd)
		}
	}
	if p.Quality().R2 < 0.999 {
		t.Errorf("R²=%v, want ≈1", p.Quality().R2)
	}
}

func TestPolyTooFewPoints(t *testing.T) {
	if _, err := Poly([]Point{{1, 1}, {2, 2}, {3, 3}}, 3); err != ErrTooFewPoints {
		t.Errorf("err=%v, want ErrTooFewPoints", err)
	}
}

func TestPolyCoeffsExpansion(t *testing.T) {
	pts, _ := cubicPoints()
	p, _ := Poly(pts, 3)
	c := p.Coeffs() // ascending powers in x
	// evaluate via raw coeffs and compare to Predict
	eval := func(x float64) float64 {
		s := 0.0
		for j := len(c) - 1; j >= 0; j-- {
			s = s*x + c[j]
		}
		return s
	}
	for _, x := range []float64{7, 13, 19} {
		if math.Abs(eval(x)-p.Predict(x)) > 1e-6 {
			t.Errorf("Coeffs eval(%v)=%v vs Predict=%v", x, eval(x), p.Predict(x))
		}
	}
}

func TestQualityFlagsNonMonotone(t *testing.T) {
	// data with a clear interior dip → non-monotone cubic
	pts := []Point{{0, 5}, {2, 2}, {4, 1}, {6, 2}, {8, 5}, {10, 9}}
	p, err := Poly(pts, 3)
	if err != nil {
		t.Fatal(err)
	}
	if p.Quality().Monotonic {
		t.Error("expected non-monotone flag for U-shaped data")
	}
	if len(p.Quality().Warnings) == 0 {
		t.Error("expected a warning")
	}
}

func TestExponentialFit(t *testing.T) {
	// synthetic a+b e^{cx}: 0.8 + 0.3 e^{0.18x}
	f := func(x float64) float64 { return 0.8 + 0.3*math.Exp(0.18*x) }
	var pts []Point
	for x := 6.0; x <= 20; x += 2 {
		pts = append(pts, Point{X: x, Y: f(x)})
	}
	e, err := Exponential(pts)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []float64{8, 14, 18} {
		if math.Abs(e.Predict(x)-f(x)) > 0.05 {
			t.Errorf("Exp.Predict(%v)=%v, want ≈%v", x, e.Predict(x), f(x))
		}
	}
	if e.C() <= 0 {
		t.Errorf("C()=%v, want >0", e.C())
	}
}

func TestSplineMonotone(t *testing.T) {
	pts := []Point{{6, 1.2}, {8, 1.4}, {10, 1.7}, {12, 2.2}, {14, 3.0}, {16, 4.5}}
	s, err := Spline(pts, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v := s.Predict(11); v < 1.7 || v > 2.2 {
		t.Errorf("Spline.Predict(11)=%v out of expected band", v)
	}
}

func TestLogLogBreakpoint(t *testing.T) {
	pts := []Point{{6, 1.24}, {8, 1.19}, {10, 1.32}, {12, 1.66}, {14, 2.38}, {16, 3.89}, {18, 6.66}, {20, 7.74}}
	l, err := LogLogSeg(pts)
	if err != nil {
		t.Fatal(err)
	}
	bp := l.Breakpoints()
	if len(bp) != 1 || bp[0] < 6 || bp[0] > 20 {
		t.Errorf("log-log breakpoint = %v, want in [6,20]", bp)
	}
}

func TestSegmentedCurveTurnPoints(t *testing.T) {
	pts := []Point{{6, 1.24}, {8, 1.19}, {10, 1.32}, {12, 1.66}, {14, 2.38}, {16, 3.89}, {18, 6.66}, {20, 7.74}}
	s, err := SegmentedCurve(pts)
	if err != nil {
		t.Fatal(err)
	}
	bp := s.Breakpoints()
	if len(bp) != 2 || bp[0] >= bp[1] {
		t.Errorf("turn points = %v, want two ascending knots", bp)
	}
}
