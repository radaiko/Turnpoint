package fit

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

// PolyFit is a least-squares polynomial fit. To keep the Vandermonde system
// well-conditioned, the fit is performed on centred/scaled inputs t=(x-mean)/sd;
// coefficients are stored in t-space and derivatives chain-rule the 1/sd scale.
type PolyFit struct {
	coef       []float64 // ascending powers in t-space
	mean, sd   float64
	xmin, xmax float64
	order      int
	q          Quality
}

// Poly fits a polynomial of the given order by QR least squares.
func Poly(pts []Point, order int) (*PolyFit, error) {
	n := len(pts)
	if n < order+1 {
		return nil, ErrTooFewPoints
	}
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range pts {
		xs[i], ys[i] = p.X, p.Y
	}
	mean, sd := stat.MeanStdDev(xs, nil)
	if sd == 0 {
		return nil, ErrSingular
	}
	p := order + 1
	x := mat.NewDense(n, p, nil)
	for i := range xs {
		t := (xs[i] - mean) / sd
		pw := 1.0
		for j := 0; j < p; j++ {
			x.Set(i, j, pw)
			pw *= t
		}
	}
	y := mat.NewVecDense(n, append([]float64(nil), ys...))
	var qr mat.QR
	qr.Factorize(x)
	var c mat.VecDense
	if err := qr.SolveVecTo(&c, false, y); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSingular, err)
	}
	coef := make([]float64, p)
	for j := range coef {
		coef[j] = c.AtVec(j)
	}
	f := &PolyFit{coef: coef, mean: mean, sd: sd, xmin: pts[0].X, xmax: pts[n-1].X, order: order}
	cond := mat.Cond(x, 2)
	f.q = assess(pts, f.Predict, f.Derivative, f.xmin, f.xmax, cond, "fit:"+f.Kind().String())
	return f, nil
}

func (f *PolyFit) Kind() Kind {
	if f.order == 4 {
		return KindPoly4
	}
	return KindPoly3
}

// Predict evaluates the polynomial at x via Horner in t-space.
func (f *PolyFit) Predict(x float64) float64 {
	t := (x - f.mean) / f.sd
	s := 0.0
	for j := len(f.coef) - 1; j >= 0; j-- {
		s = s*t + f.coef[j]
	}
	return s
}

// Derivative returns dL/dx = (dL/dt)·(1/sd).
func (f *PolyFit) Derivative(x float64) float64 {
	t := (x - f.mean) / f.sd
	s := 0.0
	for j := len(f.coef) - 1; j >= 1; j-- {
		s = s*t + float64(j)*f.coef[j]
	}
	return s / f.sd
}

// SecondDerivative returns d²L/dx² = (d²L/dt²)·(1/sd²).
func (f *PolyFit) SecondDerivative(x float64) float64 {
	t := (x - f.mean) / f.sd
	s := 0.0
	for j := len(f.coef) - 1; j >= 2; j-- {
		s = s*t + float64(j*(j-1))*f.coef[j]
	}
	return s / (f.sd * f.sd)
}

// Coeffs returns the polynomial coefficients in the ORIGINAL x basis (ascending
// powers), expanding the centred/scaled form. Used for companion-matrix roots.
func (f *PolyFit) Coeffs() []float64 {
	// L(x) = Σ a_j ((x-mean)/sd)^j. Expand to powers of x.
	p := len(f.coef)
	out := make([]float64, p)
	// binomial expansion of ((x-mean)/sd)^j = (1/sd^j) Σ_k C(j,k) x^k (-mean)^{j-k}
	for j := 0; j < p; j++ {
		aj := f.coef[j] / pow(f.sd, j)
		for k := 0; k <= j; k++ {
			out[k] += aj * binom(j, k) * pow(-f.mean, j-k)
		}
	}
	return out
}

func (f *PolyFit) Domain() (float64, float64) { return f.xmin, f.xmax }
func (f *PolyFit) Quality() Quality           { return f.q }

func pow(b float64, e int) float64 {
	r := 1.0
	for i := 0; i < e; i++ {
		r *= b
	}
	return r
}

func binom(n, k int) float64 {
	if k < 0 || k > n {
		return 0
	}
	r := 1.0
	for i := 0; i < k; i++ {
		r = r * float64(n-i) / float64(i+1)
	}
	return r
}
