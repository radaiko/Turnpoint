// Package metrics derives the per-threshold and per-zone display metrics from a
// fitted test: interpolated heart rate, % of max intensity, pace, and kcal/h
// (FR-D3/D5, FR-Z6, Appendix C). Depends on core/domain, core/unit, gonum.
package metrics

import (
	"math"

	"github.com/radaiko/turnpoint/core/domain"
	"gonum.org/v1/gonum/interp"
)

// HRCurve interpolates heart rate as a function of intensity, piecewise-linear
// over the loaded steps (the baseline row is excluded). Appendix C: 167 bpm at
// 16.1 km/h.
type HRCurve struct {
	pl         interp.PiecewiseLinear
	xmin, xmax float64
	ok         bool
}

// NewHRCurve builds the HR-vs-intensity interpolant from a test's steps.
func NewHRCurve(steps []domain.Step) HRCurve {
	type pt struct{ x, hr float64 }
	var pts []pt
	for _, s := range steps {
		if s.Intensity <= 0 || !s.HasLactate {
			continue // exclude baseline and lactate-less rows
		}
		pts = append(pts, pt{s.Intensity, float64(s.HeartRate)})
	}
	// sort + dedup by intensity
	for i := 1; i < len(pts); i++ {
		for j := i; j > 0 && pts[j].x < pts[j-1].x; j-- {
			pts[j], pts[j-1] = pts[j-1], pts[j]
		}
	}
	xs := make([]float64, 0, len(pts))
	ys := make([]float64, 0, len(pts))
	for i, p := range pts {
		if i > 0 && p.x == pts[i-1].x {
			continue
		}
		xs = append(xs, p.x)
		ys = append(ys, p.hr)
	}
	c := HRCurve{}
	if len(xs) < 2 {
		return c
	}
	if err := c.pl.Fit(xs, ys); err != nil {
		return c
	}
	c.xmin, c.xmax, c.ok = xs[0], xs[len(xs)-1], true
	return c
}

// At returns the interpolated heart rate at an intensity, rounded to a whole bpm.
// Intensities outside the measured range are clamped.
func (c HRCurve) At(intensity float64) int {
	if !c.ok {
		return 0
	}
	x := intensity
	if x < c.xmin {
		x = c.xmin
	}
	if x > c.xmax {
		x = c.xmax
	}
	return int(math.Round(c.pl.Predict(x)))
}

// Valid reports whether the curve has enough points to interpolate.
func (c HRCurve) Valid() bool { return c.ok }
