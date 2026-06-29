package unit

import (
	"testing"
	"time"
)

func TestPaceMMSS(t *testing.T) {
	cases := []struct {
		kmh  float64
		want string
	}{
		{16.1, "03:43"},
		{20.0, "03:00"},
		{13.1, "04:34"},
		{0, "—"},
		{-5, "—"},
	}
	for _, c := range cases {
		if got := PaceFromKmh(c.kmh).MMSS(); got != c.want {
			t.Errorf("PaceFromKmh(%v).MMSS() = %q, want %q", c.kmh, got, c.want)
		}
	}
}

func TestSportUnitsAndPace(t *testing.T) {
	if !Running.HasPace() {
		t.Error("Running should have pace")
	}
	if Cycling.HasPace() {
		t.Error("Cycling should not have pace")
	}
	if Running.Unit() != UnitKmh || Cycling.Unit() != UnitWatt {
		t.Error("sport→unit mapping wrong")
	}
	if UnitKmh.Symbol() != "km/h" || UnitWatt.Symbol() != "W" {
		t.Errorf("unit symbols wrong: %q %q", UnitKmh.Symbol(), UnitWatt.Symbol())
	}
}

func TestClockRoundTrip(t *testing.T) {
	d, err := ParseClock("22:10")
	if err != nil {
		t.Fatal(err)
	}
	if d != 22*time.Minute+10*time.Second {
		t.Errorf("ParseClock(22:10) = %v", d)
	}
	if got := FormatClock(d); got != "22:10" {
		t.Errorf("FormatClock = %q, want 22:10", got)
	}
	if _, err := ParseClock("bad"); err == nil {
		t.Error("expected error for bad clock")
	}
	if _, err := ParseClock("3:99"); err == nil {
		t.Error("expected error for seconds>59")
	}
}
