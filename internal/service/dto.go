// Package service converts between the store DTOs, the core domain/analysis types,
// and the JSON-friendly DTOs the Wails frontend consumes.
package service

// XY is a chart point.
type XY struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MarkerRow is one row of the threshold results table (FR-D5).
type MarkerRow struct {
	Marker     string  `json:"marker"`
	Intensity  float64 `json:"intensity"`
	Lactate    float64 `json:"lactate"`
	HeartRate  int     `json:"heartRate"`
	PctMax     float64 `json:"pctMax"`
	Pace       string  `json:"pace"`
	KcalPerHr  float64 `json:"kcalPerHr"`
	HasKcal    bool    `json:"hasKcal"`
	FitType    string  `json:"fitType"`
	Computable bool    `json:"computable"`
	Reason     string  `json:"reason,omitempty"`
}

// AnchorDTO is an LT1/LT2 anchor for the frontend.
type AnchorDTO struct {
	Marker    string  `json:"marker"`
	Intensity float64 `json:"intensity"`
	Lactate   float64 `json:"lactate"`
	HeartRate int     `json:"heartRate"`
	Pace      string  `json:"pace"`
	PctMax    float64 `json:"pctMax"`
	Manual    bool    `json:"manual"`
}

// ZoneDTO is one training zone with display ranges.
type ZoneDTO struct {
	Index         int     `json:"index"`
	Label         string  `json:"label"`
	IntensityLow  float64 `json:"intensityLow"`
	IntensityHigh float64 `json:"intensityHigh"`
	HRLow         int     `json:"hrLow"`
	HRHigh        int     `json:"hrHigh"`
	LactateLow    float64 `json:"lactateLow"`
	LactateHigh   float64 `json:"lactateHigh"`
	PaceLow       string  `json:"paceLow"`
	PaceHigh      string  `json:"paceHigh"`
}

// WarningDTO surfaces a non-blocking diagnostic to the UI.
type WarningDTO struct {
	Severity string `json:"severity"`
	Subject  string `json:"subject"`
	Message  string `json:"message"`
}

// StepBar is one intensity step bar for the temporal chart (FR-C4).
type StepBar struct {
	StartS    int     `json:"startS"`
	EndS      int     `json:"endS"`
	Intensity float64 `json:"intensity"`
}

// AnalysisDTO is the full analysis payload for the frontend.
type AnalysisDTO struct {
	Sport        string       `json:"sport"`
	Unit         string       `json:"unit"`
	HasPace      bool         `json:"hasPace"`
	RawPoints    []XY         `json:"rawPoints"`
	Curve        []XY         `json:"curve"`    // sampled display fit
	HRPoints     []XY         `json:"hrPoints"` // (intensity, hr)
	TimeHR       []XY         `json:"timeHR"`   // (time s, hr) for temporal chart
	TimeLactate  []XY         `json:"timeLactate"`
	StepBars     []StepBar    `json:"stepBars"`
	Markers      []MarkerRow  `json:"markers"`
	LT1          AnchorDTO    `json:"lt1"`
	LT2          AnchorDTO    `json:"lt2"`
	Zones        []ZoneDTO    `json:"zones"`
	MaxIntensity float64      `json:"maxIntensity"`
	DomainLow    float64      `json:"domainLow"`
	DomainHigh   float64      `json:"domainHigh"`
	Warnings     []WarningDTO `json:"warnings"`
}
