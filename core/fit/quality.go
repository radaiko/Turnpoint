package fit

import (
	"github.com/radaiko/turnpoint/core/domain"
)

// condLimit is the conditioning threshold above which a fit is flagged
// ill-conditioned (FR-F3).
const condLimit = 1e8

// assess computes R², monotonicity, conditioning and the resulting non-blocking
// warnings for a fit over its domain (FR-F3 / OI-14).
func assess(pts []Point, predict, deriv func(float64) float64, lo, hi, cond float64, subject string) Quality {
	q := Quality{Monotonic: true, Conditioned: true}

	// R² against the fit points.
	var ybar float64
	for _, p := range pts {
		ybar += p.Y
	}
	ybar /= float64(len(pts))
	var ssRes, ssTot float64
	for _, p := range pts {
		d := p.Y - predict(p.X)
		ssRes += d * d
		dt := p.Y - ybar
		ssTot += dt * dt
	}
	if ssTot > 0 {
		q.R2 = 1 - ssRes/ssTot
	} else {
		q.R2 = 1
	}

	// Monotonicity: scan the derivative for an interior sign change.
	const samples = 200
	var prevSign int
	for i := 0; i <= samples; i++ {
		x := lo + float64(i)*(hi-lo)/samples
		d := deriv(x)
		sign := 0
		if d > 0 {
			sign = 1
		} else if d < 0 {
			sign = -1
		}
		if sign != 0 && prevSign != 0 && sign != prevSign && i > 0 && i < samples {
			q.Monotonic = false
			x := x
			q.LocalExtremum = &x
			break
		}
		if sign != 0 {
			prevSign = sign
		}
	}

	if !q.Monotonic {
		q.Warnings = append(q.Warnings, domain.Warnf(domain.WarnNonMonotonicFit, subject,
			"fit is non-monotonic (interior extremum near %.1f); threshold placement may be unreliable", *q.LocalExtremum))
	}
	if q.R2 < 0.95 {
		q.Conditioned = false
		q.Warnings = append(q.Warnings, domain.Warnf(domain.WarnLowR2, subject,
			"fit R²=%.3f is below 0.95", q.R2))
	}
	if cond >= condLimit {
		q.Conditioned = false
		q.Warnings = append(q.Warnings, domain.Warnf(domain.WarnIllConditioned, subject,
			"design matrix is ill-conditioned (cond=%.2g)", cond))
	}
	return q
}
