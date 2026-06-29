package analysis

import (
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/fit"
	"github.com/radaiko/turnpoint/core/metrics"
	"github.com/radaiko/turnpoint/core/threshold"
)

// enabledMethods resolves the methods to run from the config (all default methods
// when none are explicitly enabled).
func enabledMethods(cfg Config) []threshold.ThresholdMethod {
	if len(cfg.EnabledMarkers) == 0 {
		return threshold.Default()
	}
	return threshold.For(cfg.EnabledMarkers...)
}

// buildFits constructs each distinct fit kind required by the display choice and
// the enabled methods exactly once, collecting any build failures as warnings.
func buildFits(pts []fit.Point, cfg Config, methods []threshold.ThresholdMethod) (map[fit.Kind]fit.Fit, []domain.Warning) {
	need := map[fit.Kind]bool{cfg.DisplayFit: true}
	for _, m := range methods {
		if k := m.RequiredFit(); k != fit.KindNone {
			need[k] = true
		}
	}
	fits := map[fit.Kind]fit.Fit{}
	var warns []domain.Warning
	// deterministic order over the kinds
	for k := fit.KindPoly3; k <= fit.KindSegmented; k++ {
		if !need[k] {
			continue
		}
		f, err := fit.New(k, pts)
		if err != nil {
			warns = append(warns, domain.Warnf(domain.WarnMethodNotComputable, "fit:"+k.String(),
				"could not build %s fit: %v", k, err))
			continue
		}
		fits[k] = f
	}
	return fits, warns
}

// buildContext resolves the baseline lactate (resting if >0, else the minimum
// measured lactate — DESIGN risk 5) and assembles the threshold context.
func buildContext(test domain.Test, pts []fit.Point, cfg Config) threshold.Context {
	base, hasBase := test.Baseline()
	if !hasBase || base <= 0 {
		base = minLactate(pts)
		hasBase = false
	}
	params := threshold.Params{}
	return threshold.Context{
		Points:          pts,
		Steps:           test.Steps,
		BaselineLactate: base,
		HasBaseline:     hasBase,
		Params:          params,
	}
}

func minLactate(pts []fit.Point) float64 {
	if len(pts) == 0 {
		return 0
	}
	m := pts[0].Y
	for _, p := range pts {
		if p.Y < m {
			m = p.Y
		}
	}
	return m
}

// selectAnchor resolves an LT1/LT2 anchor: a manual override, else the computed
// result for the configured marker (FR-Z2/Z3).
func selectAnchor(marker threshold.Marker, override *Override, byMarker map[threshold.Marker]threshold.Result,
	displayFit fit.Fit, hr metrics.HRCurve, test domain.Test, maxI float64) Anchor {

	if override != nil {
		lact := displayFit.Predict(override.Intensity)
		return Anchor{
			Marker:    marker,
			Intensity: override.Intensity,
			Manual:    true,
			Metrics:   metrics.Derive(override.Intensity, maxI, test.Protocol.Sport, hr, test.BodyMassKg, lact),
		}
	}
	r, ok := byMarker[marker]
	if !ok || !r.Computable {
		return Anchor{Marker: marker}
	}
	return Anchor{
		Marker:    marker,
		Intensity: r.Intensity,
		Manual:    false,
		Metrics:   metrics.Derive(r.Intensity, maxI, test.Protocol.Sport, hr, test.BodyMassKg, r.Lactate),
	}
}

// anchorWarnings flags a non-computable or extrapolated LT1/LT2 anchor (review #2/#5).
func anchorWarnings(a Anchor, label string, flo, fhi float64) []domain.Warning {
	if a.Intensity <= 0 {
		return []domain.Warning{domain.Warnf(domain.WarnMethodNotComputable, label,
			"%s anchor (%s) is not computable", label, a.Marker)}
	}
	if a.Intensity < flo || a.Intensity > fhi {
		return []domain.Warning{domain.Warnf(domain.WarnExtrapolated, label,
			"%s at %.1f is outside the tested range [%.1f, %.1f]; value is extrapolated", label, a.Intensity, flo, fhi)}
	}
	return nil
}

// rangeWarnings flags out-of-range values (OI-12 proposed bounds).
func rangeWarnings(test domain.Test) []domain.Warning {
	var w []domain.Warning
	for _, s := range test.Steps {
		if s.HasLactate && (s.Lactate < 0 || s.Lactate > 30) {
			w = append(w, domain.Warnf(domain.WarnImplausibleValue, "lactate",
				"lactate %.2f mmol/L at intensity %.1f is out of range 0–30", s.Lactate, s.Intensity))
		}
		if s.HeartRate < 0 || s.HeartRate > 250 {
			w = append(w, domain.Warnf(domain.WarnImplausibleValue, "hr",
				"HR %d bpm at intensity %.1f is out of range 0–250", s.HeartRate, s.Intensity))
		}
	}
	return w
}
