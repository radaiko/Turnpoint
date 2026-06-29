package domain

import "sort"

// FitPoint is one (intensity, lactate) sample feeding the curve fit. Defined here
// (not in fit/) so domain stays at the bottom of the import DAG; fit.Point aliases
// this type.
type FitPoint struct {
	X float64 `json:"x"` // intensity (native unit)
	Y float64 `json:"y"` // lactate (mmol/L)
}

// Test is a single step test bound (in the app layer) to an athlete (FR-T1).
type Test struct {
	Protocol   Protocol `json:"protocol"`
	Steps      []Step   `json:"steps"`
	BodyMassKg float64  `json:"bodyMassKg"` // body_mass_snapshot (OI-3); 0 ⇒ kcal/h disabled
}

// FitPoints returns the (intensity, lactate) samples for curve fitting, sorted by
// intensity and de-duplicated by intensity. Steps without lactate or marked
// Excluded are dropped. The baseline row (intensity 0) is included only when
// includeBaseline is set (the analysis parity knob, default false).
func (t Test) FitPoints(includeBaseline bool) []FitPoint {
	pts := make([]FitPoint, 0, len(t.Steps))
	for _, s := range t.Steps {
		if !s.InFit() {
			continue
		}
		if s.IsBaseline() && !includeBaseline {
			continue
		}
		pts = append(pts, FitPoint{X: s.Intensity, Y: s.Lactate})
	}
	sort.SliceStable(pts, func(i, j int) bool { return pts[i].X < pts[j].X })
	// de-dup by intensity, keeping the first occurrence
	out := pts[:0:0]
	for i, p := range pts {
		if i > 0 && p.X == pts[i-1].X {
			continue
		}
		out = append(out, p)
	}
	return out
}

// Baseline returns the resting-row lactate (intensity 0) if such a row exists.
func (t Test) Baseline() (lactate float64, ok bool) {
	for _, s := range t.Steps {
		if s.IsBaseline() && s.HasLactate {
			return s.Lactate, true
		}
	}
	return 0, false
}

// MaxIntensity returns the peak loaded-step intensity — the MAX marker and the
// denominator for %-of-max (Appendix C).
func (t Test) MaxIntensity() float64 {
	max := 0.0
	for _, s := range t.Steps {
		if s.Intensity > max {
			max = s.Intensity
		}
	}
	return max
}

// HasAbortedStep reports whether any step is flagged aborted (OI-13).
func (t Test) HasAbortedStep() bool {
	for _, s := range t.Steps {
		if s.Aborted {
			return true
		}
	}
	return false
}
