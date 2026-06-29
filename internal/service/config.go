package service

import (
	"github.com/radaiko/turnpoint/core/analysis"
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/threshold"
	"github.com/radaiko/turnpoint/core/unit"
	"github.com/radaiko/turnpoint/core/zone"
)

// MethodParamDTO carries the configurable parameters of a marker (FR-D2).
type MethodParamDTO struct {
	OBLAConc      float64 `json:"oblaConc"`
	BaselineDelta float64 `json:"baselineDelta"`
}

// AnalysisConfigDTO is the JSON-friendly analysis configuration the frontend
// edits and the app persists per test.
type AnalysisConfigDTO struct {
	DisplayFit           string                    `json:"displayFit"` // poly3 | exp | spline
	IncludeBaselineInFit bool                      `json:"includeBaselineInFit"`
	EnabledMarkers       []string                  `json:"enabledMarkers"` // marker display names
	MethodParams         map[string]MethodParamDTO `json:"methodParams"`
	LT1Anchor            string                    `json:"lt1Anchor"`
	LT2Anchor            string                    `json:"lt2Anchor"`
	LT1Override          *float64                  `json:"lt1Override"`
	LT2Override          *float64                  `json:"lt2Override"`
	ProfileName          string                    `json:"profileName"`
}

// MarkerOption describes a marker for the UI (name + the fit it depends on).
type MarkerOption struct {
	Name    string `json:"name"`
	FitType string `json:"fitType"`
}

// AllMarkerOptions lists every marker and its required fit (FR-D2/D4).
func AllMarkerOptions() []MarkerOption {
	out := []MarkerOption{}
	methods := map[threshold.Marker]threshold.ThresholdMethod{}
	for _, m := range threshold.Default() {
		methods[m.Marker()] = m
	}
	for _, mk := range threshold.AllMarkers() {
		fitName := "none"
		if m, ok := methods[mk]; ok {
			fitName = m.RequiredFit().String()
		}
		out = append(out, MarkerOption{Name: mk.String(), FitType: fitName})
	}
	return out
}

// DefaultConfigDTO returns the shipped defaults for a sport as a DTO.
func DefaultConfigDTO(sport string) AnalysisConfigDTO {
	cfg := analysis.DefaultConfig()
	prof := cfg.Profile
	if ParseSport(sport) == unit.Cycling {
		for _, p := range zone.Predefined() {
			if p.Sport == unit.Cycling && p.Level == "Leistung" {
				prof = p
				break
			}
		}
	}
	return FromConfig(cfg, prof.Name)
}

// FromConfig converts a core config (+ profile name) to the DTO.
func FromConfig(cfg analysis.Config, profileName string) AnalysisConfigDTO {
	dto := AnalysisConfigDTO{
		DisplayFit:           cfg.DisplayFit.String(),
		IncludeBaselineInFit: cfg.IncludeBaselineInFit,
		LT1Anchor:            cfg.LT1Anchor.String(),
		LT2Anchor:            cfg.LT2Anchor.String(),
		ProfileName:          profileName,
		MethodParams:         map[string]MethodParamDTO{},
		EnabledMarkers:       []string{},
	}
	for _, m := range cfg.EnabledMarkers {
		dto.EnabledMarkers = append(dto.EnabledMarkers, m.String())
	}
	for m, p := range cfg.MethodParams {
		dto.MethodParams[m.String()] = MethodParamDTO{OBLAConc: p.OBLAConc, BaselineDelta: p.BaselineDelta}
	}
	if cfg.LT1Override != nil {
		v := cfg.LT1Override.Intensity
		dto.LT1Override = &v
	}
	if cfg.LT2Override != nil {
		v := cfg.LT2Override.Intensity
		dto.LT2Override = &v
	}
	return dto
}

// ToConfig converts a DTO back to a core config, resolving the profile by name.
func ToConfig(dto AnalysisConfigDTO) analysis.Config {
	cfg := analysis.Config{
		DisplayFit:           fit.KindFromString(dto.DisplayFit),
		IncludeBaselineInFit: dto.IncludeBaselineInFit,
		MethodParams:         map[threshold.Marker]threshold.Params{},
		Profile:              profileByName(dto.ProfileName),
	}
	for _, name := range dto.EnabledMarkers {
		if m, ok := threshold.MarkerFromString(name); ok {
			cfg.EnabledMarkers = append(cfg.EnabledMarkers, m)
		}
	}
	if len(cfg.EnabledMarkers) == 0 {
		cfg.EnabledMarkers = analysis.DefaultConfig().EnabledMarkers
	}
	for name, p := range dto.MethodParams {
		if m, ok := threshold.MarkerFromString(name); ok {
			cfg.MethodParams[m] = threshold.Params{OBLAConc: p.OBLAConc, BaselineDelta: p.BaselineDelta}
		}
	}
	if m, ok := threshold.MarkerFromString(dto.LT1Anchor); ok {
		cfg.LT1Anchor = m
	} else {
		cfg.LT1Anchor = threshold.LogLog
	}
	if m, ok := threshold.MarkerFromString(dto.LT2Anchor); ok {
		cfg.LT2Anchor = m
	} else {
		cfg.LT2Anchor = threshold.OBLA4
	}
	if dto.LT1Override != nil {
		cfg.LT1Override = &analysis.Override{Intensity: *dto.LT1Override}
	}
	if dto.LT2Override != nil {
		cfg.LT2Override = &analysis.Override{Intensity: *dto.LT2Override}
	}
	return cfg
}

// ProfileOption describes a training profile for the UI.
type ProfileOption struct {
	Name       string `json:"name"`
	Sport      string `json:"sport"`
	Calibrated bool   `json:"calibrated"`
}

// ProfileOptionsForSport lists the predefined profiles for a sport (FR-Z5).
func ProfileOptionsForSport(sport string) []ProfileOption {
	want := ParseSport(sport)
	out := []ProfileOption{}
	for _, p := range zone.Predefined() {
		if p.Sport == want {
			out = append(out, ProfileOption{Name: p.Name, Sport: p.Sport.String(), Calibrated: p.Calibrated})
		}
	}
	return out
}

func profileByName(name string) zone.TrainingProfile {
	for _, p := range zone.Predefined() {
		if p.Name == name {
			return p
		}
	}
	return zone.LaufLeistung6()
}
