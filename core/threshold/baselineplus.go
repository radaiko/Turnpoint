package threshold

import "github.com/radaiko/turnpoint/core/fit"

// bslnMethod — Baseline-plus: intensity where the curve reaches baseline lactate
// plus a fixed delta (Berg 1990, Zoladz 1995). Baseline is resolved upstream as
// the resting value if >0, else the minimum measured lactate (DESIGN risk 5).
type bslnMethod struct {
	marker Marker
	delta  float64
}

func (m bslnMethod) Marker() Marker { return m.marker }

// RequiredFit = spline: like OBLA, baseline-plus is a fixed-concentration crossing
// and is reproduced faithfully on the monotone interpolating curve (see oblaMethod).
func (m bslnMethod) RequiredFit() fit.Kind { return fit.KindSpline }

func (m bslnMethod) Compute(f fit.Fit, ctx Context) Result {
	delta := m.delta
	if ctx.Params.BaselineDelta > 0 {
		delta = ctx.Params.BaselineDelta
	}
	target := ctx.BaselineLactate + delta
	v, ok := smallestRoot(f, target)
	if !ok {
		return notComputable(m.marker, f.Kind(), "baseline+delta not reached on the fitted curve")
	}
	return Result{Marker: m.marker, Intensity: v, Lactate: target, FitKind: f.Kind(), Computable: true}
}
