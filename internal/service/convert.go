package service

import (
	"time"

	"github.com/radaiko/turnpoint/core/analysis"
	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/metrics"
	"github.com/radaiko/turnpoint/core/threshold"
	"github.com/radaiko/turnpoint/core/unit"
	"github.com/radaiko/turnpoint/internal/store"
)

// ParseSport maps a stored sport string to the domain enum.
func ParseSport(s string) unit.Sport {
	switch s {
	case "cycling":
		return unit.Cycling
	default:
		return unit.Running
	}
}

// StoreToDomain builds a core domain.Test from stored test + step rows.
func StoreToDomain(t store.Test, steps []store.Step) domain.Test {
	dt := domain.Test{
		Protocol: domain.Protocol{
			Sport:          ParseSport(t.Sport),
			StepDuration:   time.Duration(t.StepDurationS) * time.Second,
			Increment:      t.Increment,
			StartIntensity: t.StartIntensity,
		},
		BodyMassKg: derefF(t.BodyMassSnapshot),
	}
	if t.Mode == "intermittent" {
		dt.Protocol.Mode = domain.Intermittent
		dt.Protocol.RestDuration = time.Duration(derefI(t.RestDurationS)) * time.Second
	}
	for _, s := range steps {
		dt.Steps = append(dt.Steps, domain.Step{
			Order:      s.StepOrder,
			Intensity:  s.Intensity,
			TimePoint:  time.Duration(derefI(s.TimePointS)) * time.Second,
			HeartRate:  derefI(s.HeartRate),
			Lactate:    derefF(s.Lactate),
			HasLactate: s.Lactate != nil,
			RPE:        s.RPE,
			Aborted:    s.Aborted,
			Excluded:   s.Excluded,
		})
	}
	return dt
}

// BuildAnalysisDTO assembles the frontend payload from a core analysis result.
func BuildAnalysisDTO(res analysis.Result, dt domain.Test) AnalysisDTO {
	sport := dt.Protocol.Sport
	lo, hi := res.DisplayFit.Domain()
	// All slice fields are initialised non-nil so they marshal to JSON [] rather
	// than null — a null would break Svelte {#each} in the consuming views.
	dto := AnalysisDTO{
		Sport:        sport.String(),
		Unit:         sport.Unit().Symbol(),
		HasPace:      sport.HasPace(),
		MaxIntensity: res.MaxIntensity,
		DomainLow:    lo,
		DomainHigh:   hi,
		RawPoints:    []XY{},
		Curve:        []XY{},
		HRPoints:     []XY{},
		TimeHR:       []XY{},
		TimeLactate:  []XY{},
		StepBars:     []StepBar{},
		Markers:      []MarkerRow{},
		Zones:        []ZoneDTO{},
		Warnings:     []WarningDTO{},
	}
	// raw points + HR/time series + step bars
	for _, s := range dt.Steps {
		if s.HasLactate && !s.IsBaseline() {
			dto.RawPoints = append(dto.RawPoints, XY{s.Intensity, s.Lactate})
			dto.HRPoints = append(dto.HRPoints, XY{s.Intensity, float64(s.HeartRate)})
		}
		ts := s.TimePoint.Seconds()
		if s.HeartRate > 0 {
			dto.TimeHR = append(dto.TimeHR, XY{ts, float64(s.HeartRate)})
		}
		if s.HasLactate {
			dto.TimeLactate = append(dto.TimeLactate, XY{ts, s.Lactate})
		}
	}
	dto.StepBars = stepBars(dt)

	// sampled display curve
	const n = 120
	for i := 0; i <= n; i++ {
		x := lo + float64(i)*(hi-lo)/n
		dto.Curve = append(dto.Curve, XY{x, res.DisplayFit.Predict(x)})
	}

	// threshold rows in canonical order
	for _, r := range res.Thresholds {
		row := MarkerRow{
			Marker:     r.Marker.String(),
			FitType:    r.FitKind.String(),
			Computable: r.Computable,
			Reason:     r.Reason,
		}
		if r.Computable {
			dm := res.Markers[r.Marker]
			row.Intensity = dm.Intensity
			row.Lactate = dm.Lactate
			row.HeartRate = dm.HeartRate
			row.PctMax = dm.PctMax
			row.Pace = paceStr(dm)
			row.KcalPerHr = dm.KcalPerHour
			row.HasKcal = dm.HasKcal
		}
		dto.Markers = append(dto.Markers, row)
	}

	dto.LT1 = anchorDTO(res.LT1)
	dto.LT2 = anchorDTO(res.LT2)
	for _, z := range res.Zones {
		dto.Zones = append(dto.Zones, ZoneDTO{
			Index: int(z.Index), Label: z.Label,
			IntensityLow: z.IntensityLow, IntensityHigh: z.IntensityHigh,
			HRLow: z.HRLow, HRHigh: z.HRHigh,
			LactateLow: z.LactateLow, LactateHigh: z.LactateHigh,
			PaceLow: z.PaceLow.MMSS(), PaceHigh: z.PaceHigh.MMSS(),
		})
	}
	for _, w := range res.Warnings {
		dto.Warnings = append(dto.Warnings, WarningDTO{Severity: sev(w.Severity), Subject: w.Subject, Message: w.Message})
	}
	return dto
}

func anchorDTO(a analysis.Anchor) AnchorDTO {
	return AnchorDTO{
		Marker: a.Marker.String(), Intensity: a.Intensity, Lactate: a.Metrics.Lactate,
		HeartRate: a.Metrics.HeartRate, Pace: paceStr(a.Metrics), PctMax: a.Metrics.PctMax, Manual: a.Manual,
	}
}

func stepBars(dt domain.Test) []StepBar {
	var bars []StepBar
	for _, s := range dt.Steps {
		if s.IsBaseline() {
			continue
		}
		start := int(s.TimePoint.Seconds()) - int(dt.Protocol.StepDuration.Seconds())
		if start < 0 {
			start = 0
		}
		bars = append(bars, StepBar{StartS: start, EndS: int(s.TimePoint.Seconds()), Intensity: s.Intensity})
	}
	return bars
}

// ThresholdSnapshotRows / ZoneSnapshotRows build persistable snapshots (FR-D5/Z4).
func ThresholdSnapshotRows(res analysis.Result) []store.ThresholdResult {
	var rows []store.ThresholdResult
	add := func(method string, dm metrics.DerivedMetrics, fitType string, computable bool, manual bool, reason string) {
		row := store.ThresholdResult{Method: method, FitType: validFit(fitType), IsOverride: manual}
		if computable {
			i, l, hr, pm := dm.Intensity, dm.Lactate, float64(dm.HeartRate), dm.PctMax
			row.Intensity, row.Lactate, row.HeartRate, row.PctMax = &i, &l, &hr, &pm
			if dm.HasPace {
				p := float64(dm.Pace) / float64(time.Second)
				row.PaceSPerKm = &p
			}
			if dm.HasKcal {
				k := dm.KcalPerHour
				row.KcalH = &k
			}
		} else if reason != "" {
			r := reason
			row.NotComputableReason = &r
		}
		rows = append(rows, row)
	}
	for _, r := range res.Thresholds {
		add(r.Marker.String(), res.Markers[r.Marker], r.FitKind.String(), r.Computable, false, r.Reason)
	}
	add("IAS", res.LT1.Metrics, fitOf(res, res.LT1), true, res.LT1.Manual, "")
	add("IANS", res.LT2.Metrics, fitOf(res, res.LT2), true, res.LT2.Manual, "")
	return rows
}

func ZoneSnapshotRows(res analysis.Result) []store.Zone {
	var rows []store.Zone
	for _, z := range res.Zones {
		il, ih, ll, lh := z.IntensityLow, z.IntensityHigh, z.LactateLow, z.LactateHigh
		hl, hh := float64(z.HRLow), float64(z.HRHigh)
		row := store.Zone{ZoneIndex: int(z.Index) + 1, ZoneName: z.Index.German(),
			IntensityLow: &il, IntensityHigh: &ih, LactateLow: &ll, LactateHigh: &lh, HRLow: &hl, HRHigh: &hh}
		if z.PaceLow > 0 {
			pl := float64(z.PaceLow) / float64(time.Second)
			ph := float64(z.PaceHigh) / float64(time.Second)
			row.PaceLowSPerKm, row.PaceHighSPerKm = &pl, &ph
		}
		rows = append(rows, row)
	}
	return rows
}

func fitOf(res analysis.Result, a analysis.Anchor) string {
	if a.Manual {
		return "none"
	}
	for _, r := range res.Thresholds {
		if r.Marker == a.Marker {
			return r.FitKind.String()
		}
	}
	return "none"
}

func validFit(s string) string {
	switch s {
	case "poly3", "poly4", "exp", "spline", "loglog", "segmented", "none":
		return s
	default:
		return "none"
	}
}

func paceStr(dm metrics.DerivedMetrics) string {
	if !dm.HasPace {
		return ""
	}
	return dm.Pace.MMSS()
}

func sev(s domain.Severity) string {
	if s == domain.Warn {
		return "warn"
	}
	return "info"
}

// AnchorMarker resolves an anchor name ("IAS"/"IANS") usage if needed elsewhere.
var _ = threshold.OBLA4

func derefF(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
func derefI(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
