package unit

// Sport is the v1 set of supported step-test sports (FR-T4).
type Sport uint8

const (
	SportUnknown Sport = iota
	Running            // intensity in km/h
	Cycling            // intensity in W
)

// Unit returns the native intensity unit for the sport.
func (s Sport) Unit() Unit {
	switch s {
	case Running:
		return UnitKmh
	case Cycling:
		return UnitWatt
	default:
		return UnitKmh
	}
}

// HasPace reports whether pace (time per 1000 m) is a meaningful derived metric
// for the sport. True only for distance sports (Running); cycling has no pace.
func (s Sport) HasPace() bool {
	return s == Running
}

func (s Sport) String() string {
	switch s {
	case Running:
		return "Running"
	case Cycling:
		return "Cycling"
	default:
		return "Unknown"
	}
}
