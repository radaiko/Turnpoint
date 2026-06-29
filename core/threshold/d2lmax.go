package threshold

import (
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
)

// d2lmaxMethod — Maximum acceleration (Newell 2006/2007): the intensity at the
// maximum of the second derivative. On a cubic L” is linear (degenerate), so
// this method PINS a 4th-order polynomial (DESIGN risk 2, FR-D4). Not in lactater.
type d2lmaxMethod struct{}

func (d2lmaxMethod) Marker() Marker        { return D2Lmax }
func (d2lmaxMethod) RequiredFit() fit.Kind { return fit.KindPoly4 }

func (d2lmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	lo, hi := f.Domain()
	const steps = 2000
	bestV, bestAcc, bestI := lo, f.SecondDerivative(lo), 0
	for i := 1; i <= steps; i++ {
		v := lo + float64(i)*(hi-lo)/steps
		if a := f.SecondDerivative(v); a > bestAcc {
			bestV, bestAcc, bestI = v, a, i
		}
	}
	r := Result{Marker: D2Lmax, Intensity: bestV, Lactate: f.Predict(bestV), FitKind: f.Kind(), Computable: true}
	// A maximum on the domain boundary is degenerate (the L'' max is not interior
	// for this fit) — flag it rather than reporting a boundary value as a result (review #9).
	if bestI == 0 || bestI == steps {
		r.Warnings = append(r.Warnings, domain.Warnf(domain.WarnMethodNotComputable, D2Lmax.String(),
			"D2Lmax maximum lies on the domain boundary; the fit has no interior acceleration peak"))
	}
	return r
}
