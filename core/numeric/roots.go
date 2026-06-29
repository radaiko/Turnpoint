// Package numeric provides deterministic root-finding and segmented-regression
// primitives shared by the fit and threshold packages. Determinism is a hard
// requirement (NFR-6): no randomness, no time, fixed grid searches.
package numeric

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// LevelSetRoot finds the smallest intensity v in [lo,hi] where f(v) == target,
// by scanning a fine grid for a sign change of f(v)-target then refining with
// bisection. Returns ok=false when target is never crossed in range.
//
// Used for OBLA (f=fitted lactate, target=concentration), Bsln+ and IAT.
func LevelSetRoot(f func(float64) float64, lo, hi, target float64) (root float64, ok bool) {
	if hi <= lo {
		return 0, false
	}
	g := func(v float64) float64 { return f(v) - target }
	const steps = 2000
	prev := lo
	gPrev := g(prev)
	if gPrev == 0 {
		return prev, true
	}
	dv := (hi - lo) / steps
	for i := 1; i <= steps; i++ {
		cur := lo + float64(i)*dv
		gCur := g(cur)
		if gCur == 0 {
			return cur, true
		}
		if (gPrev < 0) != (gCur < 0) { // sign change ⇒ bracketed root
			return bisect(g, prev, cur), true
		}
		prev, gPrev = cur, gCur
	}
	return 0, false
}

// bisect refines a bracketed root of g on [a,b] (g(a),g(b) opposite signs).
func bisect(g func(float64) float64, a, b float64) float64 {
	ga := g(a)
	for i := 0; i < 100; i++ {
		m := 0.5 * (a + b)
		gm := g(m)
		if gm == 0 || (b-a) < 1e-10 {
			return m
		}
		if (ga < 0) == (gm < 0) {
			a, ga = m, gm
		} else {
			b = m
		}
	}
	return 0.5 * (a + b)
}

// QuadraticRoots returns the real roots of a + b*x + c*x² (ascending). For a
// linear equation (c==0) it returns the single root. Empty if none/degenerate.
func QuadraticRoots(a, b, c float64) []float64 {
	if c == 0 {
		if b == 0 {
			return nil
		}
		return []float64{-a / b}
	}
	disc := b*b - 4*c*a
	if disc < 0 {
		return nil
	}
	sq := math.Sqrt(disc)
	r1 := (-b - sq) / (2 * c)
	r2 := (-b + sq) / (2 * c)
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	return []float64{r1, r2}
}

// PolyRealRoots returns the real roots of a polynomial given coefficients in
// ascending power order (coef[0] + coef[1]x + ...), via the eigenvalues of the
// companion matrix — matching R's polyroot(). Complex roots are dropped.
func PolyRealRoots(coef []float64) []float64 {
	// trim trailing (highest-order) zero coefficients
	n := len(coef)
	for n > 0 && coef[n-1] == 0 {
		n--
	}
	if n <= 1 {
		return nil // constant: no roots
	}
	deg := n - 1
	lead := coef[deg]
	// companion matrix (deg x deg)
	c := mat.NewDense(deg, deg, nil)
	for i := 0; i < deg; i++ {
		c.Set(0, i, -coef[deg-1-i]/lead)
	}
	for i := 1; i < deg; i++ {
		c.Set(i, i-1, 1)
	}
	var eig mat.Eigen
	if ok := eig.Factorize(c, mat.EigenRight); !ok {
		return nil
	}
	vals := eig.Values(nil)
	var roots []float64
	for _, v := range vals {
		if math.Abs(imag(v)) < 1e-9 {
			roots = append(roots, real(v))
		}
	}
	return roots
}
