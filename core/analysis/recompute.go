package analysis

import (
	"errors"

	"github.com/radaiko/turnpoint/core/metrics"
	"github.com/radaiko/turnpoint/core/zone"
)

// RecomputeZones is the marker-drag fast path (FR-C2/F4, NFR-3 <100 ms). It reuses
// the already-built fits from a previous Analyze result and recomputes ONLY the
// two anchors (as manual overrides at lt1/lt2), the zones, and their metrics — no
// curve refit.
func RecomputeZones(prev Result, in Input, cfg Config, lt1, lt2 float64) (Result, error) {
	if prev.DisplayFit == nil {
		return Result{}, errors.New("analysis: RecomputeZones needs a prior Analyze result")
	}
	hr := metrics.NewHRCurve(in.Test.Steps)
	sport := in.Test.Protocol.Sport
	maxI := prev.MaxIntensity

	out := prev
	out.LT1 = Anchor{
		Marker:    cfg.LT1Anchor,
		Intensity: lt1,
		Manual:    true,
		Metrics:   metrics.Derive(lt1, maxI, sport, hr, in.Test.BodyMassKg, prev.DisplayFit.Predict(lt1)),
	}
	out.LT2 = Anchor{
		Marker:    cfg.LT2Anchor,
		Intensity: lt2,
		Manual:    true,
		Metrics:   metrics.Derive(lt2, maxI, sport, hr, in.Test.BodyMassKg, prev.DisplayFit.Predict(lt2)),
	}
	out.Zones = zone.Derive(cfg.Profile, lt1, lt2, prev.DisplayFit, hr, sport)
	return out, nil
}
