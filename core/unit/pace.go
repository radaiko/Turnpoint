package unit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Pace is the time taken to cover 1000 m — a derived display metric for distance
// sports only (FR-Z6, Appendix C). It is stored as a duration but never persisted
// as a primary value.
type Pace time.Duration

// PaceFromKmh converts a running speed in km/h to pace per 1000 m. Returns Pace(0)
// for non-positive speeds (treated as "no pace").
func PaceFromKmh(kmh float64) Pace {
	if kmh <= 0 {
		return Pace(0)
	}
	// time for 1000 m = 3600/kmh seconds.
	secs := 3600.0 / kmh
	return Pace(time.Duration(secs * float64(time.Second)))
}

// MMSS formats the pace as "mm:ss", truncating to whole seconds (a pace of
// 3:43.6 displays as "03:43", matching the reference report). Returns "—" for a
// zero/invalid pace.
func (p Pace) MMSS() string {
	if p <= 0 {
		return "—"
	}
	total := int(time.Duration(p).Seconds()) // truncate
	return fmt.Sprintf("%02d:%02d", total/60, total%60)
}

// ParseClock parses an "mm:ss" time point (FR-T5) into a duration. Minutes may
// exceed 59 (e.g. a long step). Seconds must be 0–59.
func ParseClock(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("unit: invalid clock %q, want mm:ss", s)
	}
	mm, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || mm < 0 {
		return 0, fmt.Errorf("unit: invalid minutes in %q", s)
	}
	ss, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || ss < 0 || ss > 59 {
		return 0, fmt.Errorf("unit: invalid seconds in %q", s)
	}
	return time.Duration(mm)*time.Minute + time.Duration(ss)*time.Second, nil
}

// FormatClock renders a duration as "mm:ss" (minutes may exceed 59).
func FormatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d", total/60, total%60)
}
