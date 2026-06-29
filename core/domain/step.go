package domain

import "time"

// Step is one stage of a step test (FR-T2). Intensity is in the sport's native
// unit; intensity 0 marks the resting/baseline row (FR-T3).
type Step struct {
	Order      int           `json:"order"`
	Intensity  float64       `json:"intensity"`
	TimePoint  time.Duration `json:"timePoint"`     // mm:ss from start
	HeartRate  int           `json:"heartRate"`     // bpm
	Lactate    float64       `json:"lactate"`       // mmol/L
	HasLactate bool          `json:"hasLactate"`    // false ⇒ excluded from the fit (empty cell)
	RPE        *int          `json:"rpe,omitempty"` // optional Borg 6..20
	Aborted    bool          `json:"aborted"`       // aborted final step; still in fit by default (OI-13)
	Excluded   bool          `json:"excluded"`      // user per-step fit exclusion (OI-13)
}

// IsBaseline reports whether this is the resting/baseline row (intensity 0).
func (s Step) IsBaseline() bool { return s.Intensity == 0 }

// InFit reports whether the step participates in curve fitting: it must have a
// lactate value and not be user-excluded.
func (s Step) InFit() bool { return s.HasLactate && !s.Excluded }
