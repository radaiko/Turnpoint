// Package threshold computes the lactate threshold methods of SRS §6. Each method
// declares the canonical fit it requires (FR-D4); the analysis pipeline builds
// that fit once and dispatches, never silently mixing fit/method pairs.
package threshold

import (
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/numeric"
)

// Marker identifies a threshold marker.
type Marker uint8

const (
	OBLA2 Marker = iota
	OBLA3
	OBLA4
	OBLA6
	Bsln05
	Bsln10
	Bsln15
	LogLog
	Dmax
	ModDmax
	ExpDmax
	LTP1
	LTP2
	IAT
	LTratio
	D2Lmax
	MAX
)

func (m Marker) String() string {
	switch m {
	case OBLA2:
		return "OBLA 2.0"
	case OBLA3:
		return "OBLA 3.0"
	case OBLA4:
		return "OBLA 4.0"
	case OBLA6:
		return "OBLA 6.0"
	case Bsln05:
		return "Bsln+0.5"
	case Bsln10:
		return "Bsln+1.0"
	case Bsln15:
		return "Bsln+1.5"
	case LogLog:
		return "Log-log"
	case Dmax:
		return "Dmax"
	case ModDmax:
		return "ModDmax"
	case ExpDmax:
		return "Exp-Dmax"
	case LTP1:
		return "LTP1"
	case LTP2:
		return "LTP2"
	case IAT:
		return "IAT"
	case LTratio:
		return "LTratio"
	case D2Lmax:
		return "D2Lmax"
	case MAX:
		return "MAX"
	default:
		return "?"
	}
}

// Params are the user-configurable parameters for parameterised methods (FR-D2).
type Params struct {
	OBLAConc      float64 `json:"oblaConc"`
	BaselineDelta float64 `json:"baselineDelta"`
}

// Context carries everything a method needs beyond its required fit.
type Context struct {
	Points          []fit.Point   // the fit input points (sorted, deduped)
	Steps           []domain.Step // raw steps (for MAX and ModDmax)
	BaselineLactate float64       // resolved: resting if >0 else min measured (DESIGN risk 5)
	HasBaseline     bool
	Params          Params
}

// Result is one row of the threshold table (FR-D5), snapshot-friendly.
type Result struct {
	Marker     Marker           `json:"marker"`
	Intensity  float64          `json:"intensity"` // native unit; meaningless if !Computable
	Lactate    float64          `json:"lactate"`
	FitKind    fit.Kind         `json:"fitKind"` // FR-D4: the fit this row depends on
	Computable bool             `json:"computable"`
	Reason     string           `json:"reason,omitempty"`
	Warnings   []domain.Warning `json:"warnings,omitempty"`
}

// ThresholdMethod computes exactly one marker.
type ThresholdMethod interface {
	Marker() Marker
	RequiredFit() fit.Kind
	Compute(f fit.Fit, ctx Context) Result
}

// Default returns every §6 method with its shipped parameters.
func Default() []ThresholdMethod {
	return []ThresholdMethod{
		oblaMethod{OBLA2, 2.0}, oblaMethod{OBLA3, 3.0}, oblaMethod{OBLA4, 4.0}, oblaMethod{OBLA6, 6.0},
		bslnMethod{Bsln05, 0.5}, bslnMethod{Bsln10, 1.0}, bslnMethod{Bsln15, 1.5},
		logLogMethod{}, dmaxMethod{}, modDmaxMethod{}, expDmaxMethod{},
		ltpMethod{LTP1, 0}, ltpMethod{LTP2, 1},
		iatMethod{}, ltratioMethod{}, d2lmaxMethod{}, maxMethod{},
	}
}

// For returns the subset of Default methods matching the enabled markers (FR-D2).
func For(markers ...Marker) []ThresholdMethod {
	want := make(map[Marker]bool, len(markers))
	for _, m := range markers {
		want[m] = true
	}
	var out []ThresholdMethod
	for _, m := range Default() {
		if want[m.Marker()] {
			out = append(out, m)
		}
	}
	return out
}

// ── shared helpers ──────────────────────────────────────────────────────────

// rateProvider is implemented by ExpFit (the exponential rate parameter c).
type rateProvider interface{ C() float64 }

// smallestRoot finds the smallest intensity in the fit's domain where the curve
// equals target, returning ok=false when never crossed.
func smallestRoot(f fit.Fit, target float64) (float64, bool) {
	lo, hi := f.Domain()
	return numeric.LevelSetRoot(f.Predict, lo, hi, target)
}

// smallestRootAbove finds the smallest intensity in [lo,hi] where the curve
// equals target.
func smallestRootAbove(f fit.Fit, lo, hi, target float64) (float64, bool) {
	return numeric.LevelSetRoot(f.Predict, lo, hi, target)
}

func notComputable(m Marker, k fit.Kind, reason string) Result {
	return Result{Marker: m, FitKind: k, Computable: false, Reason: reason,
		Warnings: []domain.Warning{domain.Warnf(domain.WarnMethodNotComputable, m.String(), "%s", reason)}}
}
