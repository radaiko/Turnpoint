// Package unit models the measurement units of a step test. Intensity is always
// stored in the sport's native numeric unit (km/h or W); pace is a derived
// display-only metric and is never a primary stored value (OI-9, FR-T4).
package unit

// Unit is the native intensity unit of a sport.
type Unit uint8

const (
	UnitKmh  Unit = iota // running: kilometres per hour
	UnitWatt             // cycling: watts
)

// Symbol returns the unit label shown in the entry grid header (FR-T5).
func (u Unit) Symbol() string {
	switch u {
	case UnitKmh:
		return "km/h"
	case UnitWatt:
		return "W"
	default:
		return ""
	}
}
