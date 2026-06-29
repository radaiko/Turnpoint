package zone

import (
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/metrics"
	"github.com/radaiko/turnpoint/core/unit"
)

// Zone is one derived training zone with its display ranges (FR-Z4).
type Zone struct {
	Index         Index     `json:"index"`
	Label         string    `json:"label"`
	IntensityLow  float64   `json:"intensityLow"`
	IntensityHigh float64   `json:"intensityHigh"`
	HRLow         int       `json:"hrLow"`
	HRHigh        int       `json:"hrHigh"`
	LactateLow    float64   `json:"lactateLow"`
	LactateHigh   float64   `json:"lactateHigh"`
	PaceLow       unit.Pace `json:"paceLow"`
	PaceHigh      unit.Pace `json:"paceHigh"`
}

// Derive produces the five zones for a profile and the two anchors. Under the v1
// SpreadPctIANS rule, intensity bounds are fractions of IANS; lactate/HR/pace at
// each bound are read off the fitted curve and HR interpolation (FR-Z4). Only
// IANS moves the bands under this rule, which makes the marker-drag fast path
// cheap (FR-C2). ias is accepted for future spread rules.
func Derive(p TrainingProfile, ias, ians float64, curve fit.Fit, hr metrics.HRCurve, sport unit.Sport) []Zone {
	_ = ias
	zones := make([]Zone, 0, len(p.Bands))
	for _, b := range p.Bands {
		lo := b.LowPct * ians
		hi := b.HighPct * ians
		z := Zone{
			Index:         b.Zone,
			Label:         label(b.Zone, p.GermanLabels),
			IntensityLow:  lo,
			IntensityHigh: hi,
			LactateLow:    curve.Predict(lo),
			LactateHigh:   curve.Predict(hi),
			HRLow:         hr.At(lo),
			HRHigh:        hr.At(hi),
		}
		if sport.HasPace() {
			z.PaceLow = unit.PaceFromKmh(lo)
			z.PaceHigh = unit.PaceFromKmh(hi)
		}
		zones = append(zones, z)
	}
	return zones
}

func label(i Index, german bool) string {
	if german {
		return i.German()
	}
	return i.English()
}
