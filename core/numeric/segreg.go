package numeric

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// SegmentedResult holds a fitted continuous piecewise-linear regression.
type SegmentedResult struct {
	Knots []float64 // breakpoints ψ (ascending), len == k
	Coef  []float64 // [β0, β1, γ1, …, γk] for basis {1, x, (x-ψ1)+, …}
	RSS   float64
}

// SegmentedFit fits a continuous piecewise-linear model with k interior
// breakpoints by a deterministic grid search over knot positions, minimising the
// residual sum of squares (NFR-6). k=1 yields the log-log breakpoint; k=2 yields
// the two lactate turn points (LTP1, LTP2).
//
// Returns ok=false when the data has too few points to identify k+1 segments.
func SegmentedFit(x, y []float64, k int) (SegmentedResult, bool) {
	n := len(x)
	if n < 2*(k+1) || len(y) != n || k < 1 {
		return SegmentedResult{}, false
	}
	lo, hi := x[0], x[0]
	for _, xi := range x {
		lo = math.Min(lo, xi)
		hi = math.Max(hi, xi)
	}
	const grid = 80
	cands := make([]float64, 0, grid-1)
	for i := 1; i < grid; i++ {
		cands = append(cands, lo+float64(i)*(hi-lo)/grid)
	}

	best := SegmentedResult{RSS: math.Inf(1)}
	found := false
	var search func(start int, knots []float64)
	search = func(start int, knots []float64) {
		if len(knots) == k {
			if !segmentsPopulated(x, knots) {
				return
			}
			coef, rss, ok := fitHinges(x, y, knots)
			if ok && rss < best.RSS {
				ks := append([]float64(nil), knots...)
				best = SegmentedResult{Knots: ks, Coef: coef, RSS: rss}
				found = true
			}
			return
		}
		for i := start; i < len(cands); i++ {
			search(i+1, append(knots, cands[i]))
		}
	}
	search(0, nil)
	return best, found
}

// segmentsPopulated requires ≥2 points before the first knot and after the last,
// and ≥1 point in every interior segment, so the basis is identifiable.
func segmentsPopulated(x, knots []float64) bool {
	counts := make([]int, len(knots)+1)
	for _, xi := range x {
		seg := 0
		for seg < len(knots) && xi >= knots[seg] {
			seg++
		}
		counts[seg]++
	}
	if counts[0] < 2 || counts[len(counts)-1] < 2 {
		return false
	}
	for i := 1; i < len(counts)-1; i++ {
		if counts[i] < 1 {
			return false
		}
	}
	return true
}

// fitHinges solves the least-squares hinge basis for fixed knots.
func fitHinges(x, y, knots []float64) (coef []float64, rss float64, ok bool) {
	n := len(x)
	p := 2 + len(knots)
	xm := mat.NewDense(n, p, nil)
	for i := range x {
		xm.Set(i, 0, 1)
		xm.Set(i, 1, x[i])
		for j, k := range knots {
			h := x[i] - k
			if h < 0 {
				h = 0
			}
			xm.Set(i, 2+j, h)
		}
	}
	yv := mat.NewVecDense(n, append([]float64(nil), y...))
	var qr mat.QR
	qr.Factorize(xm)
	var c mat.VecDense
	if err := qr.SolveVecTo(&c, false, yv); err != nil {
		return nil, math.Inf(1), false
	}
	coef = make([]float64, p)
	for j := range coef {
		coef[j] = c.AtVec(j)
	}
	for i := range x {
		pred := coef[0] + coef[1]*x[i]
		for j, k := range knots {
			if x[i] > k {
				pred += coef[2+j] * (x[i] - k)
			}
		}
		d := y[i] - pred
		rss += d * d
	}
	return coef, rss, true
}
