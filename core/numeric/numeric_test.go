package numeric

import (
	"math"
	"testing"
)

func TestLevelSetRoot(t *testing.T) {
	// f(v)=v² ; solve v²=2 on [0,2] → √2
	root, ok := LevelSetRoot(func(v float64) float64 { return v * v }, 0, 2, 2)
	if !ok || math.Abs(root-math.Sqrt2) > 1e-6 {
		t.Fatalf("root = %v ok=%v, want √2", root, ok)
	}
	// target out of range
	if _, ok := LevelSetRoot(func(v float64) float64 { return v }, 0, 1, 5); ok {
		t.Error("expected no root for out-of-range target")
	}
}

func TestQuadraticRoots(t *testing.T) {
	// x²-2 = 0 → ±√2 ; here a=-2,b=0,c=1
	r := QuadraticRoots(-2, 0, 1)
	if len(r) != 2 || math.Abs(r[0]+math.Sqrt2) > 1e-9 || math.Abs(r[1]-math.Sqrt2) > 1e-9 {
		t.Fatalf("QuadraticRoots = %v", r)
	}
	// linear: 2x-4=0 → 2  (a=-4,b=2,c=0)
	if r := QuadraticRoots(-4, 2, 0); len(r) != 1 || math.Abs(r[0]-2) > 1e-9 {
		t.Fatalf("linear root = %v", r)
	}
}

func TestPolyRealRoots(t *testing.T) {
	// (x-1)(x-2)(x-3) = -6 +11x -6x² + x³
	roots := PolyRealRoots([]float64{-6, 11, -6, 1})
	want := []float64{1, 2, 3}
	if len(roots) != 3 {
		t.Fatalf("got %d roots: %v", len(roots), roots)
	}
	// sort + compare
	for i := 0; i < len(roots); i++ {
		for j := i + 1; j < len(roots); j++ {
			if roots[j] < roots[i] {
				roots[i], roots[j] = roots[j], roots[i]
			}
		}
	}
	for i := range want {
		if math.Abs(roots[i]-want[i]) > 1e-6 {
			t.Errorf("root %d = %v, want %v", i, roots[i], want[i])
		}
	}
}

func TestSegmentedFitRecoversBreakpoint(t *testing.T) {
	// piecewise: slope 0.1 below x=10, slope 1.0 above; breakpoint at 10.
	var x, y []float64
	for xi := 0.0; xi <= 20; xi++ {
		x = append(x, xi)
		if xi <= 10 {
			y = append(y, 0.1*xi)
		} else {
			y = append(y, 0.1*10+1.0*(xi-10))
		}
	}
	res, ok := SegmentedFit(x, y, 1)
	if !ok {
		t.Fatal("segmented fit failed")
	}
	if math.Abs(res.Knots[0]-10) > 0.5 {
		t.Errorf("breakpoint = %v, want ≈10", res.Knots[0])
	}
	// determinism
	res2, _ := SegmentedFit(x, y, 1)
	if res2.Knots[0] != res.Knots[0] {
		t.Error("segmented fit not deterministic")
	}
}
