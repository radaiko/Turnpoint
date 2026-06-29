package threshold

import "github.com/radaiko/turnpoint/core/fit"

// ltpMethod — Lactate Turn Points (Hofmann & Tschakert; Pokan): the two
// breakpoints of a two-knot segmented regression over the interpolated curve.
// LTP1 is a candidate IAS, LTP2 a candidate IANS (OI-16). idx selects ψ₁ or ψ₂.
type ltpMethod struct {
	marker Marker
	idx    int
}

func (m ltpMethod) Marker() Marker        { return m.marker }
func (m ltpMethod) RequiredFit() fit.Kind { return fit.KindSegmented }

func (m ltpMethod) Compute(f fit.Fit, ctx Context) Result {
	bp, ok := f.(fit.Breakpointer)
	if !ok || len(bp.Breakpoints()) <= m.idx {
		return notComputable(m.marker, f.Kind(), "turn point unavailable")
	}
	v := bp.Breakpoints()[m.idx]
	return Result{Marker: m.marker, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}
