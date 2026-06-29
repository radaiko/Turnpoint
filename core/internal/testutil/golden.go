package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/radaiko/turnpoint/core/domain"
	"github.com/radaiko/turnpoint/core/unit"
)

// testdataDir resolves core/testdata relative to this source file, so the helpers
// work no matter which package's test invokes them.
func testdataDir() string {
	_, self, _, _ := runtime.Caller(0) // core/internal/testutil/golden.go
	return filepath.Join(filepath.Dir(self), "..", "..", "testdata")
}

type fixtureStep struct {
	Intensity  float64 `json:"intensity"`
	TimeSec    int     `json:"timeSec"`
	HR         int     `json:"hr"`
	Lactate    float64 `json:"lactate"`
	HasLactate bool    `json:"hasLactate"`
	RPE        *int    `json:"rpe"`
	Aborted    bool    `json:"aborted"`
	Excluded   bool    `json:"excluded"`
}

type fixtureTest struct {
	Sport           string        `json:"sport"`
	StepDurationSec int           `json:"stepDurationSec"`
	Increment       float64       `json:"increment"`
	StartIntensity  float64       `json:"startIntensity"`
	Mode            string        `json:"mode"`
	BodyMassKg      float64       `json:"bodyMassKg"`
	Steps           []fixtureStep `json:"steps"`
}

// LoadAppendixA returns the canonical running step test (SRS Appendix A) as a
// domain.Test — the primary P0 validation fixture (V2).
func LoadAppendixA(t *testing.T) domain.Test {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testdataDir(), "datasets", "appendix_a.json"))
	if err != nil {
		t.Fatalf("load appendix_a: %v", err)
	}
	var f fixtureTest
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("decode appendix_a: %v", err)
	}
	sport := unit.Running
	if f.Sport == "cycling" {
		sport = unit.Cycling
	}
	mode := domain.Continuous
	if f.Mode == "intermittent" {
		mode = domain.Intermittent
	}
	test := domain.Test{
		Protocol: domain.Protocol{
			Sport:          sport,
			StepDuration:   time.Duration(f.StepDurationSec) * time.Second,
			Increment:      f.Increment,
			StartIntensity: f.StartIntensity,
			Mode:           mode,
		},
		BodyMassKg: f.BodyMassKg,
	}
	for i, s := range f.Steps {
		test.Steps = append(test.Steps, domain.Step{
			Order:      i,
			Intensity:  s.Intensity,
			TimePoint:  time.Duration(s.TimeSec) * time.Second,
			HeartRate:  s.HR,
			Lactate:    s.Lactate,
			HasLactate: s.HasLactate,
			RPE:        s.RPE,
			Aborted:    s.Aborted,
			Excluded:   s.Excluded,
		})
	}
	return test
}

// GoldenMarker is one row of the frozen WinLactat marker table (Appendix C).
type GoldenMarker struct {
	Marker  string  `json:"marker"`
	Kmh     float64 `json:"kmh"`
	Lactate float64 `json:"lactate"`
	HR      float64 `json:"hr"`
	PctMax  float64 `json:"pctMax"`
	Pace    string  `json:"pace"`
}

// LoadMarkersGolden returns the Appendix C marker table keyed by marker name.
func LoadMarkersGolden(t *testing.T) map[string]GoldenMarker {
	t.Helper()
	var doc struct {
		Markers []GoldenMarker `json:"markers"`
	}
	loadJSON(t, filepath.Join("golden", "winlactat", "appendix_c_markers.json"), &doc)
	m := make(map[string]GoldenMarker, len(doc.Markers))
	for _, row := range doc.Markers {
		m[row.Marker] = row
	}
	return m
}

// GoldenZone is one row of the frozen WinLactat zone table (Appendix C / §7).
type GoldenZone struct {
	Zone        string  `json:"zone"`
	LowPct      float64 `json:"lowPct"`
	HighPct     float64 `json:"highPct"`
	KmhLow      float64 `json:"kmhLow"`
	KmhHigh     float64 `json:"kmhHigh"`
	LactateLow  float64 `json:"lactateLow"`
	LactateHigh float64 `json:"lactateHigh"`
	HRLow       int     `json:"hrLow"`
	HRHigh      int     `json:"hrHigh"`
}

// LoadZonesGolden returns the Appendix C reference zones in ascending order.
func LoadZonesGolden(t *testing.T) []GoldenZone {
	t.Helper()
	var doc struct {
		Zones []GoldenZone `json:"zones"`
	}
	loadJSON(t, filepath.Join("golden", "winlactat", "appendix_c_zones.json"), &doc)
	return doc.Zones
}

func loadJSON(t *testing.T, rel string, dst any) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testdataDir(), rel))
	if err != nil {
		t.Fatalf("load %s: %v", rel, err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("decode %s: %v", rel, err)
	}
}
