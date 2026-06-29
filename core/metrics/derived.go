package metrics

import "github.com/radaiko/turnpoint/core/unit"

// DerivedMetrics are the display metrics for a threshold or zone bound (FR-D5,
// Appendix C).
type DerivedMetrics struct {
	Intensity   float64   `json:"intensity"`
	Lactate     float64   `json:"lactate"`
	PctMax      float64   `json:"pctMax"` // intensity / maxIntensity × 100
	Pace        unit.Pace `json:"pace"`   // 60/kmh → mm:ss (running only)
	HasPace     bool      `json:"hasPace"`
	HeartRate   int       `json:"heartRate"` // interpolated on HR-vs-intensity
	KcalPerHour float64   `json:"kcalPerHour"`
	HasKcal     bool      `json:"hasKcal"` // false when body mass is absent (FR-D5)
}

// PctMax returns intensity as a percentage of the max intensity (Appendix C:
// 16.1/20 = 80.5).
func PctMax(intensity, maxIntensity float64) float64 {
	if maxIntensity <= 0 {
		return 0
	}
	return intensity / maxIntensity * 100
}

// KcalPerHourRunning is the OI-15 proposed running estimate (Low confidence):
// body mass × speed × 1.036 (net running cost ≈ 1 kcal·kg⁻¹·km⁻¹).
func KcalPerHourRunning(bodyMassKg, speedKmh float64) float64 {
	return bodyMassKg * speedKmh * 1.036
}

// KcalPerHourCycling is the OI-15 proposed cycling estimate (Low confidence):
// power × 3.6 (mechanical 0.86 kcal/h per W ÷ ~24% gross efficiency).
func KcalPerHourCycling(powerW float64) float64 {
	return powerW * 3.6
}

// Derive builds the full DerivedMetrics for an intensity on a test.
func Derive(intensity, maxIntensity float64, sport unit.Sport, hr HRCurve, bodyMassKg, lactate float64) DerivedMetrics {
	dm := DerivedMetrics{
		Intensity: intensity,
		Lactate:   lactate,
		PctMax:    PctMax(intensity, maxIntensity),
		HeartRate: hr.At(intensity),
	}
	if sport.HasPace() {
		dm.Pace = unit.PaceFromKmh(intensity)
		dm.HasPace = true
	}
	if bodyMassKg > 0 {
		dm.HasKcal = true
		switch sport {
		case unit.Running:
			dm.KcalPerHour = KcalPerHourRunning(bodyMassKg, intensity)
		case unit.Cycling:
			dm.KcalPerHour = KcalPerHourCycling(intensity)
		}
	}
	return dm
}
