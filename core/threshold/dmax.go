package threshold

import "github.com/radaiko/turnpoint/core/fit"

// dmaxMethod — Dmax: the point where the fitted curve's slope equals the slope of
// the chord joining the first and last fitted points (Cheng 1992).
type dmaxMethod struct{}

func (dmaxMethod) Marker() Marker        { return Dmax }
func (dmaxMethod) RequiredFit() fit.Kind { return fit.KindPoly3 }

func (dmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	lo, hi := f.Domain()
	slope := (f.Predict(hi) - f.Predict(lo)) / (hi - lo)
	v, ok := rootOfDerivative(f, lo, hi, slope)
	if !ok {
		return notComputable(Dmax, f.Kind(), "no point matches the chord slope")
	}
	return Result{Marker: Dmax, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}

// modDmaxMethod — Modified Dmax (Bishop 1998): like Dmax but the chord starts at
// the first step whose lactate rose >0.4 mmol/L above the previous step (tested on
// raw consecutive steps; chord endpoints on the fitted curve). Appendix A ⇒ 14.
type modDmaxMethod struct{}

func (modDmaxMethod) Marker() Marker        { return ModDmax }
func (modDmaxMethod) RequiredFit() fit.Kind { return fit.KindPoly3 }

func (modDmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	x0, ok := firstRise(ctx.Points, 0.4)
	if !ok {
		return notComputable(ModDmax, f.Kind(), "no step rises >0.4 mmol/L above the previous")
	}
	_, hi := f.Domain()
	if x0 >= hi {
		return notComputable(ModDmax, f.Kind(), "rise point is at or beyond the last step")
	}
	slope := (f.Predict(hi) - f.Predict(x0)) / (hi - x0)
	v, found := rootOfDerivative(f, x0, hi, slope)
	if !found {
		return notComputable(ModDmax, f.Kind(), "no point matches the modified chord slope")
	}
	return Result{Marker: ModDmax, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}

// firstRise returns the intensity of the first point whose lactate exceeds the
// previous point's by more than delta.
func firstRise(pts []fit.Point, delta float64) (float64, bool) {
	for i := 1; i < len(pts); i++ {
		if pts[i].Y-pts[i-1].Y > delta {
			return pts[i].X, true
		}
	}
	return 0, false
}
