// Package zone derives the 5-zone training model from the two anchor thresholds
// (IAS/IANS) using a selectable training profile. The v1 spread rule is "% of
// IANS intensity" (FR-Z1/Z4/Z5, OI-17). Depends on core/unit, core/fit,
// core/metrics.
package zone

import "github.com/radaiko/turnpoint/core/unit"

// Index identifies a training zone (ascending intensity).
type Index uint8

const (
	REKOM Index = iota
	GA1
	GA2
	EB
	SB
)

func (i Index) German() string {
	return [...]string{"REKOM", "GA1", "GA2", "EB", "SB"}[i]
}

func (i Index) English() string {
	return [...]string{"Recovery", "Basic 1", "Basic 2", "Development", "Peak"}[i]
}

// SpreadRule is how a profile positions zone bounds around the anchors.
type SpreadRule uint8

const (
	SpreadPctIANS SpreadRule = iota // bounds = fraction × IANS intensity (v1)
	// future: SpreadIAStoIANS, SpreadPctHRmax
)

// Band is one zone's intensity range as a fraction of IANS (1.0 == IANS).
type Band struct {
	Zone    Index   `json:"zone"`
	LowPct  float64 `json:"lowPct"`
	HighPct float64 `json:"highPct"`
}

// TrainingProfile defines how zones spread around the anchors (FR-Z5, §9).
type TrainingProfile struct {
	Name         string     `json:"name"`
	Sport        unit.Sport `json:"sport"`
	Level        string     `json:"level"` // "Freizeit" | "Ambitioniert" | "Leistung"
	WeeklyFreq   int        `json:"weeklyFreq"`
	Rule         SpreadRule `json:"rule"`
	Bands        []Band     `json:"bands"`
	GermanLabels bool       `json:"germanLabels"`
	Calibrated   bool       `json:"calibrated"` // false ⇒ provisional bands (OI-17)
}

// LaufLeistung6 is the reference profile "Laufen Leistungssportler (6×/Woche)",
// calibrated to Appendix C (back-derived %-of-IANS bands).
func LaufLeistung6() TrainingProfile {
	return TrainingProfile{
		Name:         "Laufen Leistungssportler (6×/Woche)",
		Sport:        unit.Running,
		Level:        "Leistung",
		WeeklyFreq:   6,
		Rule:         SpreadPctIANS,
		GermanLabels: true,
		Calibrated:   true,
		Bands: []Band{
			{REKOM, 0.00, 0.46},
			{GA1, 0.46, 0.70},
			{GA2, 0.70, 0.88},
			{EB, 0.88, 1.02},
			{SB, 1.02, 1.25},
		},
	}
}

// Predefined returns the shipped profiles. Only LaufLeistung6 is calibrated; the
// rest carry provisional bands pending reference data (OI-17, Low confidence).
func Predefined() []TrainingProfile {
	provisional := func(name, level string, freq int, sport unit.Sport) TrainingProfile {
		p := LaufLeistung6()
		p.Name, p.Level, p.WeeklyFreq, p.Sport, p.Calibrated = name, level, freq, sport, false
		return p
	}
	return []TrainingProfile{
		LaufLeistung6(),
		provisional("Laufen Ambitioniert (4–5×/Woche)", "Ambitioniert", 5, unit.Running),
		provisional("Laufen Freizeit (3×/Woche)", "Freizeit", 3, unit.Running),
		provisional("Rad Leistungssportler (6×/Woche)", "Leistung", 6, unit.Cycling),
		provisional("Rad Ambitioniert (4–5×/Woche)", "Ambitioniert", 5, unit.Cycling),
		provisional("Rad Freizeit (3×/Woche)", "Freizeit", 3, unit.Cycling),
	}
}
