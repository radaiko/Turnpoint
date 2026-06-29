// Package analysis is the public API of turnpoint-core: it runs the full compute
// pipeline (data → fits → thresholds → LT1/LT2 → zones → derived metrics) and is
// the surface the Wails app layer binds. Pure and total: identical (Input,Config)
// yields an identical Result (NFR-6).
package analysis

import (
	"errors"

	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/metrics"
	"github.com/radaiko/turnpoint/core/threshold"
	"github.com/radaiko/turnpoint/core/zone"
)

// Input is the test under analysis.
type Input struct {
	Test domain.Test `json:"test"`
}

// Override marks a manual LT1/LT2 anchor (FR-Z3/C3).
type Override struct {
	Intensity float64 `json:"intensity"`
}

// Config holds the user's analysis selections (persisted per test).
type Config struct {
	DisplayFit           fit.Kind                              `json:"displayFit"`
	IncludeBaselineInFit bool                                  `json:"includeBaselineInFit"`
	EnabledMarkers       []threshold.Marker                    `json:"enabledMarkers"`
	MethodParams         map[threshold.Marker]threshold.Params `json:"methodParams"`
	LT1Anchor            threshold.Marker                      `json:"lt1Anchor"`
	LT2Anchor            threshold.Marker                      `json:"lt2Anchor"`
	LT1Override          *Override                             `json:"lt1Override"`
	LT2Override          *Override                             `json:"lt2Override"`
	Profile              zone.TrainingProfile                  `json:"profile"`
}

// Anchor is a selected LT1/LT2 threshold with its derived metrics (FR-Z2/Z3).
type Anchor struct {
	Marker    threshold.Marker       `json:"marker"`
	Intensity float64                `json:"intensity"`
	Manual    bool                   `json:"manual"` // FR-C3 manual vs algorithmic
	Metrics   metrics.DerivedMetrics `json:"metrics"`
}

// Result is the full analysis output.
type Result struct {
	Fits         map[fit.Kind]fit.Fit                        `json:"-"`
	DisplayFit   fit.Fit                                     `json:"-"`
	Thresholds   []threshold.Result                          `json:"thresholds"`
	Markers      map[threshold.Marker]metrics.DerivedMetrics `json:"markers"`
	LT1          Anchor                                      `json:"lt1"` // IAS
	LT2          Anchor                                      `json:"lt2"` // IANS
	Zones        []zone.Zone                                 `json:"zones"`
	MaxIntensity float64                                     `json:"maxIntensity"`
	Warnings     []domain.Warning                            `json:"warnings"`
}

var (
	ErrInsufficientSteps = errors.New("analysis: need ≥4 fit steps (FR-T6)")
	ErrNoBaseline        = errors.New("analysis: method requires a baseline row")
)

// DefaultConfig returns the shipped defaults (FR-F2 display poly3; OI-16 anchors
// LT1←Log-log, LT2←OBLA 4.0; the calibrated Lauf profile).
func DefaultConfig() Config {
	return Config{
		DisplayFit:     fit.KindPoly3,
		EnabledMarkers: defaultMarkers(),
		LT1Anchor:      threshold.LogLog,
		LT2Anchor:      threshold.OBLA4,
		Profile:        zone.LaufLeistung6(),
	}
}

func defaultMarkers() []threshold.Marker {
	ms := make([]threshold.Marker, 0, len(threshold.Default()))
	for _, m := range threshold.Default() {
		ms = append(ms, m.Marker())
	}
	return ms
}

// Analyze runs the full pipeline.
func Analyze(in Input, cfg Config) (Result, error) {
	pts := in.Test.FitPoints(cfg.IncludeBaselineInFit)
	if len(pts) < 4 {
		return Result{}, ErrInsufficientSteps
	}
	var warnings []domain.Warning
	if len(pts) < 5 {
		warnings = append(warnings, domain.Warnf(domain.WarnFewSteps, "analysis",
			"only %d fit steps; results are less reliable", len(pts)))
	}
	if in.Test.HasAbortedStep() {
		warnings = append(warnings, domain.Warning{Code: domain.WarnAbortedStep, Severity: domain.Info,
			Subject: "analysis", Message: "the final step was aborted; it is included in the fit"})
	}

	methods := enabledMethods(cfg)
	fits, fitWarnings := buildFits(pts, cfg, methods)
	warnings = append(warnings, fitWarnings...)

	displayFit := fits[cfg.DisplayFit]
	if displayFit == nil {
		return Result{}, errors.New("analysis: display fit could not be built")
	}

	ctx := buildContext(in.Test, pts, cfg)
	hr := metrics.NewHRCurve(in.Test.Steps)
	if !hr.Valid() {
		warnings = append(warnings, domain.Warnf(domain.WarnImplausibleValue, "hr",
			"heart-rate data is unavailable; interpolated HR will read 0"))
	}
	maxI := in.Test.MaxIntensity()
	flo, fhi := displayFit.Domain()

	res := Result{
		Fits:         fits,
		DisplayFit:   displayFit,
		Markers:      map[threshold.Marker]metrics.DerivedMetrics{},
		MaxIntensity: maxI,
	}
	byMarker := map[threshold.Marker]threshold.Result{}
	for _, m := range methods {
		f := fits[m.RequiredFit()]
		var r threshold.Result
		if m.RequiredFit() != fit.KindNone && f == nil {
			r = threshold.Result{Marker: m.Marker(), FitKind: m.RequiredFit(), Computable: false,
				Reason: "required fit unavailable",
				Warnings: []domain.Warning{domain.Warnf(domain.WarnMethodNotComputable, m.Marker().String(),
					"required %s fit could not be built", m.RequiredFit())}}
		} else {
			mctx := ctx
			if p, ok := cfg.MethodParams[m.Marker()]; ok { // FR-D2 per-method params
				mctx.Params = p
			}
			r = m.Compute(f, mctx)
			if f != nil {
				r.Warnings = append(r.Warnings, f.Quality().Warnings...)
			}
		}
		res.Thresholds = append(res.Thresholds, r)
		byMarker[m.Marker()] = r
		if r.Computable {
			res.Markers[m.Marker()] = metrics.Derive(r.Intensity, maxI, in.Test.Protocol.Sport, hr, in.Test.BodyMassKg, r.Lactate)
		}
	}

	res.LT1 = selectAnchor(cfg.LT1Anchor, cfg.LT1Override, byMarker, displayFit, hr, in.Test, maxI)
	res.LT2 = selectAnchor(cfg.LT2Anchor, cfg.LT2Override, byMarker, displayFit, hr, in.Test, maxI)
	warnings = append(warnings, anchorWarnings(res.LT1, "IAS", flo, fhi)...)
	warnings = append(warnings, anchorWarnings(res.LT2, "IANS", flo, fhi)...)

	// Only derive zones when the IANS scaling anchor is usable; a zero anchor
	// would otherwise collapse all five bands to [0,0] (review #2).
	if res.LT2.Intensity > 0 {
		res.Zones = zone.Derive(cfg.Profile, res.LT1.Intensity, res.LT2.Intensity, displayFit, hr, in.Test.Protocol.Sport)
	} else {
		warnings = append(warnings, domain.Warnf(domain.WarnMethodNotComputable, "zones",
			"IANS anchor (%s) is not computable; training zones cannot be derived", res.LT2.Marker))
	}

	if in.Test.BodyMassKg == 0 {
		warnings = append(warnings, domain.Warning{Code: domain.WarnNoBodyMass, Severity: domain.Info,
			Subject: "analysis", Message: "no body mass: energy expenditure (kcal/h) is disabled"})
	}
	res.Warnings = warnings
	return res, nil
}

// Validate runs the FR-T6 pre-analysis checks without computing thresholds.
func Validate(in Input) []domain.Warning {
	var w []domain.Warning
	pts := in.Test.FitPoints(false)
	if len(pts) < 4 {
		w = append(w, domain.Warnf(domain.WarnInsufficientSteps, "analysis",
			"only %d fit steps; need ≥4 to analyse", len(pts)))
	} else if len(pts) < 5 {
		w = append(w, domain.Warnf(domain.WarnFewSteps, "analysis",
			"only %d fit steps; results are less reliable", len(pts)))
	}
	w = append(w, rangeWarnings(in.Test)...)
	return w
}
