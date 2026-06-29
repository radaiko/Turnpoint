package domain

import (
	"time"

	"github.com/radaiko/turnpoint/core/unit"
)

// Mode is the test continuity (FR-T1).
type Mode uint8

const (
	Continuous Mode = iota
	Intermittent
)

// Protocol describes how a step test was run (FR-T1).
type Protocol struct {
	Sport          unit.Sport    `json:"sport"`
	StepDuration   time.Duration `json:"stepDuration"`
	Increment      float64       `json:"increment"`      // native unit (+2 km/h | +40 W)
	StartIntensity float64       `json:"startIntensity"` // native unit
	Mode           Mode          `json:"mode"`
	RestDuration   time.Duration `json:"restDuration"` // Intermittent only
}
