// Package testutil provides shared test fixtures and floating-point tolerance
// helpers for the core parity suite (OI-1, V1/V2). Test-only; not imported by
// production code.
package testutil

import (
	"testing"

	"gonum.org/v1/gonum/floats/scalar"
)

// Tolerances per OI-1.
const (
	TolIntensityRunAbs = 0.1  // km/h
	TolIntensityRel    = 0.01 // 1%
	TolIntensityCycAbs = 2.0  // W
	TolHR              = 1.0  // bpm
	TolLactate         = 0.05 // mmol/L
	// Derived metrics computed from the engine's unrounded intensity (DESIGN §6).
	TolPctMax  = 0.3 // %
	TolPaceSec = 1.0 // seconds
)

// EqualIntensityRun compares running intensities (abs 0.1 km/h OR rel 1%).
func EqualIntensityRun(got, want float64) bool {
	return scalar.EqualWithinAbsOrRel(got, want, TolIntensityRunAbs, TolIntensityRel)
}

// EqualIntensityCyc compares cycling intensities (abs 2 W OR rel 1%).
func EqualIntensityCyc(got, want float64) bool {
	return scalar.EqualWithinAbsOrRel(got, want, TolIntensityCycAbs, TolIntensityRel)
}

// EqualHR compares heart rates within 1 bpm.
func EqualHR(got, want float64) bool { return scalar.EqualWithinAbs(got, want, TolHR) }

// EqualLactate compares lactate within 0.05 mmol/L.
func EqualLactate(got, want float64) bool { return scalar.EqualWithinAbs(got, want, TolLactate) }

// AssertIntensityRun fails the test if got is outside the running-intensity tolerance.
func AssertIntensityRun(t *testing.T, name string, got, want float64) {
	t.Helper()
	if !EqualIntensityRun(got, want) {
		t.Errorf("%s intensity = %.4f, want %.2f (±%.1f km/h or ±%.0f%%)", name, got, want, TolIntensityRunAbs, TolIntensityRel*100)
	}
}

// AssertHR fails the test if got is outside ±1 bpm of want.
func AssertHR(t *testing.T, name string, got, want float64) {
	t.Helper()
	if !EqualHR(got, want) {
		t.Errorf("%s HR = %.2f, want %.0f (±1 bpm)", name, got, want)
	}
}

// AssertLactate fails the test if got is outside ±0.05 mmol/L of want.
func AssertLactate(t *testing.T, name string, got, want float64) {
	t.Helper()
	if !EqualLactate(got, want) {
		t.Errorf("%s lactate = %.4f, want %.2f (±0.05)", name, got, want)
	}
}
