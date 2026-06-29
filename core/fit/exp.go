package fit

import (
	"errors"
	"math"

	"github.com/radaiko/turnpoint/core/domain"
	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/optimize"
	"gonum.org/v1/gonum/stat"
)

// ExpFit is the model L(x) = a + b·e^{c·x}, matching lactater's exponential form
// (its Exp-Dmax helper depends only on c). Fitted by nonlinear least squares.
type ExpFit struct {
	a, b, c    float64
	xmin, xmax float64
	q          Quality
}

// Exponential fits a+b·e^{cx} via Nelder–Mead with a log-linear warm start
// (without which the NLS routinely diverges).
func Exponential(pts []Point) (*ExpFit, error) {
	n := len(pts)
	if n < 3 {
		return nil, ErrTooFewPoints
	}
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range pts {
		xs[i], ys[i] = p.X, p.Y
	}

	// Warm start: a0 just below the minimum lactate, then regress ln(y-a0) on x.
	minY := ys[0]
	for _, y := range ys {
		minY = math.Min(minY, y)
	}
	a0 := minY - 0.1
	z := make([]float64, n)
	for i := range ys {
		z[i] = math.Log(ys[i] - a0)
	}
	intercept, slope := stat.LinearRegression(xs, z, nil, false)
	c0 := slope
	b0 := math.Exp(intercept)

	sse := func(p []float64) float64 {
		a, b, c := p[0], p[1], p[2]
		var s float64
		for i := range xs {
			d := ys[i] - (a + b*math.Exp(c*xs[i]))
			s += d * d
		}
		return s
	}
	prob := optimize.Problem{
		Func: sse,
		Grad: func(grad, x []float64) { fd.Gradient(grad, sse, x, nil) },
	}
	res, err := optimize.Minimize(prob, []float64{a0, b0, c0}, nil, &optimize.NelderMead{})
	if err != nil {
		return nil, err
	}
	a, b, c := res.X[0], res.X[1], res.X[2]
	if math.IsNaN(a) || math.IsNaN(b) || math.IsNaN(c) || math.IsInf(a, 0) || math.IsInf(b, 0) || math.IsInf(c, 0) {
		return nil, errors.New("fit: exponential NLS did not converge")
	}
	f := &ExpFit{a: a, b: b, c: c, xmin: pts[0].X, xmax: pts[n-1].X}
	f.q = assess(pts, f.Predict, f.Derivative, f.xmin, f.xmax, 1, "fit:exp")
	if c <= 0 {
		f.q.Conditioned = false
		f.q.Warnings = append(f.q.Warnings, domain.Warnf(domain.WarnLowR2, "fit:exp",
			"exponential has no upward curvature (c=%.4f ≤ 0); model may be inappropriate", c))
	}
	return f, nil
}

// C returns the exponential rate parameter (used by Exp-Dmax's closed form).
func (f *ExpFit) C() float64 { return f.c }

func (f *ExpFit) Kind() Kind                         { return KindExp }
func (f *ExpFit) Predict(x float64) float64          { return f.a + f.b*math.Exp(f.c*x) }
func (f *ExpFit) Derivative(x float64) float64       { return f.b * f.c * math.Exp(f.c*x) }
func (f *ExpFit) SecondDerivative(x float64) float64 { return f.b * f.c * f.c * math.Exp(f.c*x) }
func (f *ExpFit) Domain() (float64, float64)         { return f.xmin, f.xmax }
func (f *ExpFit) Quality() Quality                   { return f.q }
