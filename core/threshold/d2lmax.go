package threshold

import "github.com/radaiko/turnpoint/core/fit"

// d2lmaxMethod — Maximum acceleration (Newell 2006/2007): the intensity at the
// maximum of the second derivative. On a cubic L” is linear (degenerate), so
// this method PINS a 4th-order polynomial (DESIGN risk 2, FR-D4). Not in lactater.
type d2lmaxMethod struct{}

func (d2lmaxMethod) Marker() Marker        { return D2Lmax }
func (d2lmaxMethod) RequiredFit() fit.Kind { return fit.KindPoly4 }

func (d2lmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	lo, hi := f.Domain()
	const steps = 2000
	bestV, bestAcc := lo, f.SecondDerivative(lo)
	for i := 1; i <= steps; i++ {
		v := lo + float64(i)*(hi-lo)/steps
		if a := f.SecondDerivative(v); a > bestAcc {
			bestV, bestAcc = v, a
		}
	}
	return Result{Marker: D2Lmax, Intensity: bestV, Lactate: f.Predict(bestV), FitKind: f.Kind(), Computable: true}
}
