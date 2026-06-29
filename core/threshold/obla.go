package threshold

import "github.com/radaiko/turnpoint/core/fit"

// oblaMethod — Onset of Blood Lactate Accumulation: intensity where the fitted
// curve reaches a fixed concentration (FR §6; Sjödin & Jacobs 1981, Heck 1985).
//
// Canonical fit = monotone spline. Empirically (cmd/probe) the WinLactat report's
// fixed thresholds (13.1/16.1/17.6) are reproduced by a monotone interpolating
// spline through the measured points — NOT by a least-squares cubic, which smooths
// the points and yields ~15.8 for 4.0. This is the binding parity anchor (V2):
// OBLA 4.0 ⇒ 16.1 km/h / 167 bpm, with IANS ← OBLA 4.0.
type oblaMethod struct {
	marker Marker
	conc   float64
}

func (m oblaMethod) Marker() Marker        { return m.marker }
func (m oblaMethod) RequiredFit() fit.Kind { return fit.KindSpline }

func (m oblaMethod) Compute(f fit.Fit, ctx Context) Result {
	conc := m.conc
	if ctx.Params.OBLAConc > 0 {
		conc = ctx.Params.OBLAConc
	}
	v, ok := smallestRoot(f, conc)
	if !ok {
		return notComputable(m.marker, f.Kind(), "concentration not reached on the fitted curve")
	}
	return Result{Marker: m.marker, Intensity: v, Lactate: conc, FitKind: f.Kind(), Computable: true}
}
