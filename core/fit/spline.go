package fit

import (
	"sort"

	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/interp"
)

// SplineFit is a monotone cubic (Fritsch–Butland) interpolating spline — the
// user-selectable "smoothing" display fit. It is monotone by construction for
// monotone data (FR-F3); no threshold method pins it, so parity is unaffected.
// (The full Eilers–Marx penalised P-spline is deferred — DESIGN risk 12.)
type SplineFit struct {
	fb         interp.FritschButland
	xmin, xmax float64
	q          Quality
}

// Spline builds a Fritsch–Butland interpolant. lambda is reserved for the future
// penalised spline and currently ignored.
func Spline(pts []Point, lambda float64) (*SplineFit, error) {
	n := len(pts)
	if n < 3 {
		return nil, ErrTooFewPoints
	}
	sorted := append([]Point(nil), pts...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].X < sorted[j].X })
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range sorted {
		xs[i], ys[i] = p.X, p.Y
	}
	f := &SplineFit{xmin: xs[0], xmax: xs[n-1]}
	if err := f.fb.Fit(xs, ys); err != nil {
		return nil, err
	}
	f.q = assess(sorted, f.Predict, f.Derivative, f.xmin, f.xmax, 1, "fit:spline")
	return f, nil
}

func (f *SplineFit) clamp(x float64) float64 {
	if x < f.xmin {
		return f.xmin
	}
	if x > f.xmax {
		return f.xmax
	}
	return x
}

func (f *SplineFit) Kind() Kind                { return KindSpline }
func (f *SplineFit) Predict(x float64) float64 { return f.fb.Predict(f.clamp(x)) }

// Derivative/SecondDerivative use finite differences on the interpolant (robust
// across gonum versions; the spline is a display fit, not a parity anchor).
func (f *SplineFit) Derivative(x float64) float64 {
	return fd.Derivative(f.Predict, f.clamp(x), nil)
}
func (f *SplineFit) SecondDerivative(x float64) float64 {
	return fd.Derivative(f.Predict, f.clamp(x), &fd.Settings{Formula: fd.Central2nd})
}
func (f *SplineFit) Domain() (float64, float64) { return f.xmin, f.xmax }
func (f *SplineFit) Quality() Quality           { return f.q }
