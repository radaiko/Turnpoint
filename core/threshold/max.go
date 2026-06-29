package threshold

import "github.com/radaiko/turnpoint/core/fit"

// maxMethod — MAX: the peak loaded step reached in the test (not curve-derived).
// Appendix A ⇒ 20.0 km/h / 185 bpm / 7.74 mmol/L.
type maxMethod struct{}

func (maxMethod) Marker() Marker        { return MAX }
func (maxMethod) RequiredFit() fit.Kind { return fit.KindNone }

func (maxMethod) Compute(f fit.Fit, ctx Context) Result {
	var peak struct {
		intensity, lactate float64
		found              bool
	}
	for _, s := range ctx.Steps {
		if s.Intensity > peak.intensity {
			peak.intensity, peak.lactate, peak.found = s.Intensity, s.Lactate, true
		}
	}
	if !peak.found {
		return notComputable(MAX, fit.KindNone, "no loaded steps")
	}
	return Result{Marker: MAX, Intensity: peak.intensity, Lactate: peak.lactate, FitKind: fit.KindNone, Computable: true}
}
