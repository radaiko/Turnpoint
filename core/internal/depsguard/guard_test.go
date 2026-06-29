// Package depsguard enforces the core-purity firewall (SRS §10, V6): no package
// under core/ may import the app layer (internal/), Wails, or the SQLite driver.
package depsguard

import (
	"os/exec"
	"strings"
	"testing"
)

// forbidden import-path substrings a pure-compute core package must never pull in.
var forbidden = []string{
	"github.com/radaiko/turnpoint/internal", // app-only layer
	"github.com/wailsapp/wails",             // desktop shell
	"modernc.org/sqlite",                    // DB driver
	"database/sql",                          // DB access
}

func TestCoreHasNoForbiddenDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}",
		"github.com/radaiko/turnpoint/core/...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed: %v\n%s", err, out)
	}
	for _, dep := range strings.Fields(string(out)) {
		for _, bad := range forbidden {
			if strings.Contains(dep, bad) {
				t.Errorf("core/ transitively imports forbidden dependency %q (matched %q)", dep, bad)
			}
		}
	}
}
