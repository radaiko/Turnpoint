package fit

import "github.com/radaiko/turnpoint/core/numeric"

// SegCurveFit locates the two lactate turn points (LTP1, LTP2) by running a
// two-knot segmented regression over the default polynomial interpolated onto a
// fine grid — matching lactater's segmentation of pre-smoothed data (DESIGN risk
// 4). The lactate curve itself is the underlying polynomial; only the breakpoints
// are added.
type SegCurveFit struct {
	poly  *PolyFit
	knots []float64 // [LTP1, LTP2]
}

// gridStep is the fixed interpolation spacing (km/h or W) for the segmentation —
// a parity-tuned constant (DESIGN risk 4).
const gridStep = 0.1

// SegmentedCurve builds the turn-point pseudo-fit.
func SegmentedCurve(pts []Point) (*SegCurveFit, error) {
	poly, err := Poly(pts, 3)
	if err != nil {
		return nil, err
	}
	lo, hi := poly.Domain()
	var gx, gy []float64
	for x := lo; x <= hi+1e-9; x += gridStep {
		gx = append(gx, x)
		gy = append(gy, poly.Predict(x))
	}
	f := &SegCurveFit{poly: poly}
	if seg, ok := numeric.SegmentedFit(gx, gy, 2); ok {
		f.knots = seg.Knots
	}
	return f, nil
}

// Breakpoints returns [LTP1, LTP2] (empty if the segmentation failed).
func (f *SegCurveFit) Breakpoints() []float64 { return f.knots }

func (f *SegCurveFit) Kind() Kind                         { return KindSegmented }
func (f *SegCurveFit) Predict(x float64) float64          { return f.poly.Predict(x) }
func (f *SegCurveFit) Derivative(x float64) float64       { return f.poly.Derivative(x) }
func (f *SegCurveFit) SecondDerivative(x float64) float64 { return f.poly.SecondDerivative(x) }
func (f *SegCurveFit) Domain() (float64, float64)         { return f.poly.Domain() }
func (f *SegCurveFit) Quality() Quality                   { return f.poly.Quality() }
