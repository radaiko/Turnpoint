package threshold

import "github.com/radaiko/turnpoint/core/fit"

// ltratioMethod — Minimum Lactate-Intensity Ratio (Jamnick 2018): the intensity
// minimising L(v)/v on the fitted curve (typically at low intensity).
type ltratioMethod struct{}

func (ltratioMethod) Marker() Marker        { return LTratio }
func (ltratioMethod) RequiredFit() fit.Kind { return fit.KindPoly3 }

func (ltratioMethod) Compute(f fit.Fit, ctx Context) Result {
	lo, hi := f.Domain()
	if lo <= 0 {
		lo = hi / 1000 // avoid division by zero at v=0
	}
	const steps = 2000
	bestV, bestRatio := lo, f.Predict(lo)/lo
	for i := 1; i <= steps; i++ {
		v := lo + float64(i)*(hi-lo)/steps
		if v <= 0 {
			continue
		}
		if r := f.Predict(v) / v; r < bestRatio {
			bestV, bestRatio = v, r
		}
	}
	return Result{Marker: LTratio, Intensity: bestV, Lactate: f.Predict(bestV), FitKind: f.Kind(), Computable: true}
}
