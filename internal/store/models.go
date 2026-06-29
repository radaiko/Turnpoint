package store

// DTO structs mirror the schema; pointer fields map to nullable columns.

type Athlete struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	DOB          *string  `json:"dob"`
	Sex          string   `json:"sex"`
	BodyMassKg   *float64 `json:"bodyMassKg"`
	PrimarySport *string  `json:"primarySport"`
	Notes        string   `json:"notes"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

// AthleteSummary is the athlete-list row (FR-A4, OI-5).
type AthleteSummary struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	PrimarySport *string `json:"primarySport"`
	LastTestDate *string `json:"lastTestDate"`
	TestCount    int     `json:"testCount"`
}

type Test struct {
	ID               int64    `json:"id"`
	AthleteID        int64    `json:"athleteId"`
	TestDate         string   `json:"testDate"`
	Sport            string   `json:"sport"`
	StepDurationS    int      `json:"stepDurationS"`
	Increment        float64  `json:"increment"`
	StartIntensity   float64  `json:"startIntensity"`
	Mode             string   `json:"mode"`
	RestDurationS    *int     `json:"restDurationS"`
	BaselineLactate  *float64 `json:"baselineLactate"`
	BodyMassSnapshot *float64 `json:"bodyMassSnapshot"`
	PretestNote      string   `json:"pretestNote"`
	Remarks          string   `json:"remarks"`
	TemplateID       *int64   `json:"templateId"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
}

type Step struct {
	ID         int64    `json:"id"`
	TestID     int64    `json:"testId"`
	StepOrder  int      `json:"stepOrder"`
	Intensity  float64  `json:"intensity"`
	TimePointS *int     `json:"timePointS"`
	HeartRate  *int     `json:"heartRate"`
	Lactate    *float64 `json:"lactate"`
	RPE        *int     `json:"rpe"`
	IsBaseline bool     `json:"isBaseline"`
	Excluded   bool     `json:"excluded"`
	Aborted    bool     `json:"aborted"`
}

type ThresholdResult struct {
	ID                  int64    `json:"id"`
	TestID              int64    `json:"testId"`
	Method              string   `json:"method"`
	Intensity           *float64 `json:"intensity"`
	Lactate             *float64 `json:"lactate"`
	HeartRate           *float64 `json:"heartRate"`
	PctMax              *float64 `json:"pctMax"`
	PaceSPerKm          *float64 `json:"paceSPerKm"`
	KcalH               *float64 `json:"kcalH"`
	IsOverride          bool     `json:"isOverride"`
	FitType             string   `json:"fitType"`
	NotComputableReason *string  `json:"notComputableReason"`
	ParamsJSON          *string  `json:"paramsJson"`
}

type Zone struct {
	ID             int64    `json:"id"`
	TestID         int64    `json:"testId"`
	Model          string   `json:"model"`
	ZoneIndex      int      `json:"zoneIndex"`
	ZoneName       string   `json:"zoneName"`
	ProfileID      *int64   `json:"profileId"`
	IntensityLow   *float64 `json:"intensityLow"`
	IntensityHigh  *float64 `json:"intensityHigh"`
	HRLow          *float64 `json:"hrLow"`
	HRHigh         *float64 `json:"hrHigh"`
	LactateLow     *float64 `json:"lactateLow"`
	LactateHigh    *float64 `json:"lactateHigh"`
	PaceLowSPerKm  *float64 `json:"paceLowSPerKm"`
	PaceHighSPerKm *float64 `json:"paceHighSPerKm"`
}

type Template struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Sport          string   `json:"sport"`
	StepDurationS  int      `json:"stepDurationS"`
	Increment      float64  `json:"increment"`
	StartIntensity float64  `json:"startIntensity"`
	EndIntensity   *float64 `json:"endIntensity"`
	Mode           string   `json:"mode"`
	RestDurationS  *int     `json:"restDurationS"`
	VisibleColumns string   `json:"visibleColumns"`
	IsPredefined   bool     `json:"isPredefined"`
}

type TrainingProfile struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Sport           string `json:"sport"`
	Level           string `json:"level"`
	WeeklyFrequency *int   `json:"weeklyFrequency"`
	SpreadJSON      string `json:"spreadJson"`
	IsPredefined    bool   `json:"isPredefined"`
}

type ReportSettings struct {
	ID              int64  `json:"id"`
	TestID          *int64 `json:"testId"`
	HeaderLogo      []byte `json:"headerLogo"`
	HeaderText      string `json:"headerText"`
	FooterText      string `json:"footerText"`
	PageSize        string `json:"pageSize"`
	Orientation     string `json:"orientation"`
	BlockConfigJSON string `json:"blockConfigJson"`
	Commentary      string `json:"commentary"`
	ShowPageNumbers bool   `json:"showPageNumbers"`
}
