// Package domain holds the pure data records of a step test plus the non-blocking
// Warning type. Per the SRS, everything "non-blocking" is carried as Warning data,
// never as a Go error (DESIGN §9). domain depends only on core/unit + stdlib.
package domain

import "fmt"

// Severity distinguishes informational notes from warnings.
type Severity uint8

const (
	Info Severity = iota
	Warn
)

// WarnCode enumerates the non-blocking diagnostics surfaced to the UI.
type WarnCode uint8

const (
	WarnFewSteps            WarnCode = iota // <5 fit steps (OI-11)
	WarnNonMonotonicFit                     // interior local extremum (OI-14a)
	WarnLowR2                               // R²<0.95 (OI-14b)
	WarnIllConditioned                      // near-singular design matrix
	WarnImplausibleValue                    // out of OI-12 range
	WarnMethodNotComputable                 // a method could not be computed (FR-D1)
	WarnAbortedStep                         // aborted final step present (OI-13)
	WarnNoBodyMass                          // kcal/h disabled (FR-D5)
	WarnExtrapolated                        // marker outside fitted domain
	WarnInsufficientSteps                   // <4 fit steps, analysis blocked (FR-T6)
)

// Warning is a non-blocking diagnostic. Subject is a free-form key (e.g. a marker
// name "OBLA 4.0" or "fit:poly3") kept as a string to avoid a dependency cycle on
// the threshold package.
type Warning struct {
	Code     WarnCode `json:"code"`
	Severity Severity `json:"severity"`
	Subject  string   `json:"subject"`
	Message  string   `json:"message"`
}

// Warnf builds a Warn-severity Warning.
func Warnf(code WarnCode, subject, format string, args ...any) Warning {
	return Warning{Code: code, Severity: Warn, Subject: subject, Message: fmt.Sprintf(format, args...)}
}
