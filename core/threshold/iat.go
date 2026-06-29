package threshold

import "github.com/radaiko/turnpoint/core/fit"

// iatMethod — Individual Anaerobic Threshold (Dickhuth 1991): the lactate-minimum
// equivalent plus 1.5 mmol/L. Find the curve's interior lactate minimum (the low-
// intensity dip), then the intensity above it where the curve reaches L_min+1.5.
// Not in lactater — validated against a hand-computed fixture (DESIGN risk 6).
type iatMethod struct{}

func (iatMethod) Marker() Marker        { return IAT }
func (iatMethod) RequiredFit() fit.Kind { return fit.KindPoly3 }

func (iatMethod) Compute(f fit.Fit, ctx Context) Result {
	lo, hi := f.Domain()
	vmin, lmin := curveMinimum(f, lo, hi)
	target := lmin + 1.5
	// search for the target strictly above the minimum
	v, ok := smallestRootAbove(f, vmin, hi, target)
	if !ok {
		return notComputable(IAT, f.Kind(), "lactate minimum + 1.5 not reached above the minimum")
	}
	return Result{Marker: IAT, Intensity: v, Lactate: target, FitKind: f.Kind(), Computable: true}
}

// curveMinimum scans for the interior lactate minimum; if the curve is monotone
// it returns the low endpoint.
func curveMinimum(f fit.Fit, lo, hi float64) (vmin, lmin float64) {
	vmin, lmin = lo, f.Predict(lo)
	const steps = 1000
	for i := 1; i <= steps; i++ {
		x := lo + float64(i)*(hi-lo)/steps
		if y := f.Predict(x); y < lmin {
			vmin, lmin = x, y
		}
	}
	return vmin, lmin
}
