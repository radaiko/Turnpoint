package main

import "testing"

func TestSemverLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"0.0.1", "0.0.2", true},
		{"0.0.2", "0.0.2", false},
		{"0.0.3", "0.0.2", false},
		{"0.9.0", "1.0.0", true},
		{"1.2.0", "1.10.0", true}, // numeric, not lexicographic
		{"0.0.2", "0.0.2-rc1", false},
		{"0.0.2-rc1", "0.0.2", false}, // pre-release suffix dropped → equal
		{"1.0.0", "0.9.9", false},
	}
	for _, c := range cases {
		if got := semverLess(c.a, c.b); got != c.want {
			t.Errorf("semverLess(%q,%q)=%v want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestParseVer(t *testing.T) {
	if got := parseVer("2.5.7"); got != [3]int{2, 5, 7} {
		t.Errorf("parseVer(2.5.7) = %v", got)
	}
	if got := parseVer("3"); got != [3]int{3, 0, 0} {
		t.Errorf("parseVer(3) = %v", got)
	}
}
