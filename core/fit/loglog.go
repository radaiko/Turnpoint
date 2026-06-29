package fit

import (
	"math"

	"github.com/radaiko/turnpoint/core/numeric"
)

// LogLogFit is the Beaver log–log breakpoint pseudo-fit: a one-knot segmented
// regression of ln(lactate) on ln(intensity); the breakpoint is back-transformed
// to an intensity. Predict reconstructs the piecewise model in linear space.
type LogLogFit struct {
	coef       []float64 // [β0, β1, γ1] in log–log space
	psi        float64   // knot in ln(intensity)
	vstar      float64   // exp(psi)
	xmin, xmax float64
	q          Quality
}

// LogLogSeg fits the log–log breakpoint. Requires strictly positive x and y.
func LogLogSeg(pts []Point) (*LogLogFit, error) {
	var lx, ly []float64
	xmin, xmax := math.Inf(1), math.Inf(-1)
	for _, p := range pts {
		if p.X <= 0 || p.Y <= 0 {
			continue // baseline / non-positive cannot be log-transformed
		}
		lx = append(lx, math.Log(p.X))
		ly = append(ly, math.Log(p.Y))
		xmin = math.Min(xmin, p.X)
		xmax = math.Max(xmax, p.X)
	}
	if len(lx) < 4 {
		return nil, ErrNonPositive
	}
	seg, ok := numeric.SegmentedFit(lx, ly, 1)
	if !ok {
		return nil, ErrTooFewPoints
	}
	f := &LogLogFit{coef: seg.Coef, psi: seg.Knots[0], vstar: math.Exp(seg.Knots[0]), xmin: xmin, xmax: xmax}
	// R² in log space.
	var ybar, ssTot, ssRes float64
	for _, v := range ly {
		ybar += v
	}
	ybar /= float64(len(ly))
	for i := range lx {
		pred := f.predictLog(lx[i])
		ssRes += (ly[i] - pred) * (ly[i] - pred)
		ssTot += (ly[i] - ybar) * (ly[i] - ybar)
	}
	if ssTot > 0 {
		f.q.R2 = 1 - ssRes/ssTot
	} else {
		f.q.R2 = 1
	}
	f.q.Monotonic = true
	f.q.Conditioned = true
	return f, nil
}

func (f *LogLogFit) predictLog(lnx float64) float64 {
	v := f.coef[0] + f.coef[1]*lnx
	if lnx > f.psi {
		v += f.coef[2] * (lnx - f.psi)
	}
	return v
}

// Breakpoints returns the single log–log breakpoint intensity.
func (f *LogLogFit) Breakpoints() []float64 { return []float64{f.vstar} }

func (f *LogLogFit) Kind() Kind { return KindLogLog }
func (f *LogLogFit) Predict(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Exp(f.predictLog(math.Log(x)))
}
func (f *LogLogFit) Derivative(x float64) float64 {
	const h = 1e-4
	return (f.Predict(x+h) - f.Predict(x-h)) / (2 * h)
}
func (f *LogLogFit) SecondDerivative(x float64) float64 {
	const h = 1e-3
	return (f.Predict(x+h) - 2*f.Predict(x) + f.Predict(x-h)) / (h * h)
}
func (f *LogLogFit) Domain() (float64, float64) { return f.xmin, f.xmax }
func (f *LogLogFit) Quality() Quality           { return f.q }
