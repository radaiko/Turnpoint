// Package fit fits continuous lactate-vs-intensity models and exposes their
// derivatives and quality diagnostics (FR-F1/F2/F3). It depends only on stdlib,
// gonum, and core/domain + core/numeric.
package fit

import (
	"errors"

	"github.com/radaiko/turnpoint/core/domain"
)

// Point is one (intensity, lactate) sample. Aliased to domain.FitPoint so domain
// stays at the bottom of the import DAG.
type Point = domain.FitPoint

// Kind identifies a fit strategy. Poly3/Poly4/Exp/Spline/LogLog/Segmented are
// constructible; None labels the MAX marker (no fit).
type Kind uint8

const (
	KindPoly3 Kind = iota
	KindPoly4
	KindExp
	KindSpline
	KindLogLog    // breakpoint pseudo-fit: segmented in log–log space
	KindSegmented // breakpoint pseudo-fit: segmented on the interpolated curve (LTP)
	KindNone      // MAX: no fit
)

// KindFromString resolves a fit kind from its string form, defaulting to Poly3.
func KindFromString(s string) Kind {
	switch s {
	case "poly4":
		return KindPoly4
	case "exp":
		return KindExp
	case "spline":
		return KindSpline
	case "loglog":
		return KindLogLog
	case "segmented":
		return KindSegmented
	case "none":
		return KindNone
	default:
		return KindPoly3
	}
}

func (k Kind) String() string {
	switch k {
	case KindPoly3:
		return "poly3"
	case KindPoly4:
		return "poly4"
	case KindExp:
		return "exp"
	case KindSpline:
		return "spline"
	case KindLogLog:
		return "loglog"
	case KindSegmented:
		return "segmented"
	default:
		return "none"
	}
}

// Fit is an immutable continuous model lactate(intensity) with derivatives and
// quality diagnostics.
type Fit interface {
	Kind() Kind
	Predict(x float64) float64          // fitted lactate at x
	Derivative(x float64) float64       // dL/dx
	SecondDerivative(x float64) float64 // d²L/dx²
	Domain() (xmin, xmax float64)       // fitted input range
	Quality() Quality
}

// Breakpointer is implemented by segmented pseudo-fits (LogLog, Segmented) that
// carry one or more breakpoint intensities.
type Breakpointer interface {
	Breakpoints() []float64
}

// Quality holds the FR-F3 / OI-14 diagnostics for a fit.
type Quality struct {
	R2            float64          `json:"r2"`
	Monotonic     bool             `json:"monotonic"`
	Conditioned   bool             `json:"conditioned"`
	LocalExtremum *float64         `json:"localExtremum,omitempty"`
	Warnings      []domain.Warning `json:"warnings,omitempty"`
}

var (
	ErrTooFewPoints = errors.New("fit: need ≥ order+1 distinct points")
	ErrSingular     = errors.New("fit: design matrix ill-conditioned")
	ErrNonPositive  = errors.New("fit: log-log needs positive x and y")
)

// New builds a fit of the given constructible kind from points.
func New(k Kind, pts []Point) (Fit, error) {
	switch k {
	case KindPoly3:
		return Poly(pts, 3)
	case KindPoly4:
		return Poly(pts, 4)
	case KindExp:
		return Exponential(pts)
	case KindSpline:
		return Spline(pts, 0)
	case KindLogLog:
		return LogLogSeg(pts)
	case KindSegmented:
		return SegmentedCurve(pts)
	default:
		return nil, errors.New("fit: kind not constructible")
	}
}
