package threshold

import (
	"math"

	"github.com/radaiko/turnpoint/core/fit"
)

// expDmaxMethod — Exponential Dmax (Newell 2007): Dmax computed on the exponential
// fit L=a+b·e^{cx}. The tangent-parallel-to-chord condition reduces to a closed
// form in c only:  v* = ln((e^{c·sf}−e^{c·si}) / (c·(sf−si))) / c.
type expDmaxMethod struct{}

func (expDmaxMethod) Marker() Marker        { return ExpDmax }
func (expDmaxMethod) RequiredFit() fit.Kind { return fit.KindExp }

func (expDmaxMethod) Compute(f fit.Fit, ctx Context) Result {
	rp, ok := f.(rateProvider)
	if !ok {
		return notComputable(ExpDmax, f.Kind(), "exponential fit unavailable")
	}
	c := rp.C()
	si, sf := f.Domain()
	if c == 0 || sf <= si {
		return notComputable(ExpDmax, f.Kind(), "degenerate exponential fit")
	}
	num := math.Exp(c*sf) - math.Exp(c*si)
	den := c * (sf - si)
	if num/den <= 0 {
		return notComputable(ExpDmax, f.Kind(), "no real Exp-Dmax solution")
	}
	v := math.Log(num/den) / c
	if v < si || v > sf || math.IsNaN(v) {
		return notComputable(ExpDmax, f.Kind(), "Exp-Dmax outside tested range")
	}
	return Result{Marker: ExpDmax, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}
