package threshold

import (
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
)

// dmaxMethod — Dmax: the point where the fitted curve's slope equals the slope of
// the chord joining the first and last fitted points (Cheng 1992).
type dmaxMethod struct{}

func (dmaxMethod) Marker() Marker        { return Dmax }
func (dmaxMethod) RequiredFit() fit.Kind { return fit.KindPoly3 }

func (dmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	lo, hi := f.Domain()
	v, ok := maxChordDistance(f, lo, hi)
	if !ok {
		return notComputable(Dmax, f.Kind(), "no interior maximum-distance point")
	}
	return Result{Marker: Dmax, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}

// maxChordDistance returns the intensity in (lo,hi) that maximises the vertical
// distance from the curve to the chord joining the curve's endpoints — the
// definition of Dmax (review #8). Returns ok=false if the max is at an endpoint.
func maxChordDistance(f fit.Fit, lo, hi float64) (float64, bool) {
	ylo, yhi := f.Predict(lo), f.Predict(hi)
	slope := (yhi - ylo) / (hi - lo)
	chord := func(v float64) float64 { return ylo + slope*(v-lo) }
	const steps = 2000
	bestV, bestD, bestI := lo, chord(lo)-f.Predict(lo), 0
	for i := 1; i <= steps; i++ {
		v := lo + float64(i)*(hi-lo)/steps
		if d := chord(v) - f.Predict(v); d > bestD {
			bestV, bestD, bestI = v, d, i
		}
	}
	return bestV, bestI > 0 && bestI < steps && bestD > 0
}

// modDmaxMethod — Modified Dmax (Bishop 1998): like Dmax but the chord starts at
// the first step whose lactate rose >0.4 mmol/L above the previous step (tested on
// raw consecutive steps; chord endpoints on the fitted curve). Appendix A ⇒ 14.
type modDmaxMethod struct{}

func (modDmaxMethod) Marker() Marker        { return ModDmax }
func (modDmaxMethod) RequiredFit() fit.Kind { return fit.KindPoly3 }

func (modDmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	x0, ok := firstRiseSteps(ctx.Steps, 0.4)
	if !ok {
		return notComputable(ModDmax, f.Kind(), "no step rises >0.4 mmol/L above the previous")
	}
	_, hi := f.Domain()
	if x0 >= hi {
		return notComputable(ModDmax, f.Kind(), "rise point is at or beyond the last step")
	}
	v, found := maxChordDistance(f, x0, hi)
	if !found {
		return notComputable(ModDmax, f.Kind(), "no interior maximum-distance point on the modified chord")
	}
	return Result{Marker: ModDmax, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}

// firstRiseSteps returns the intensity of the first loaded step (intensity>0, in
// fit) whose lactate exceeds the previous loaded step's by more than delta. It
// reads raw steps, NOT fit points, so the baseline-in-fit parity knob cannot
// perturb the chord anchor (review #4). Appendix A ⇒ 14 km/h.
func firstRiseSteps(steps []domain.Step, delta float64) (float64, bool) {
	type lp struct{ x, y float64 }
	var pts []lp
	for _, s := range steps {
		if s.Intensity > 0 && s.InFit() {
			pts = append(pts, lp{s.Intensity, s.Lactate})
		}
	}
	for i := 1; i < len(pts); i++ {
		for j := i; j > 0 && pts[j].x < pts[j-1].x; j-- {
			pts[j], pts[j-1] = pts[j-1], pts[j]
		}
	}
	for i := 1; i < len(pts); i++ {
		if pts[i].y-pts[i-1].y > delta {
			return pts[i].x, true
		}
	}
	return 0, false
}
