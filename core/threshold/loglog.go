package threshold

import "github.com/radaiko/turnpoint/core/fit"

// logLogMethod — Log-log / Beaver (1985): the breakpoint of a one-knot segmented
// regression in log(lactate) vs log(intensity), back-transformed to intensity.
// Candidate IAS default (OI-16, to confirm against lactater).
type logLogMethod struct{}

func (logLogMethod) Marker() Marker        { return LogLog }
func (logLogMethod) RequiredFit() fit.Kind { return fit.KindLogLog }

func (logLogMethod) Compute(f fit.Fit, ctx Context) Result {
	bp, ok := f.(fit.Breakpointer)
	if !ok || len(bp.Breakpoints()) == 0 {
		return notComputable(LogLog, f.Kind(), "log-log breakpoint unavailable")
	}
	v := bp.Breakpoints()[0]
	return Result{Marker: LogLog, Intensity: v, Lactate: f.Predict(v), FitKind: f.Kind(), Computable: true}
}
