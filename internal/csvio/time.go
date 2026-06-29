// Package csvio handles CSV/TSV import and export of step data and clipboard paste
// parsing (FR-M2, OI-10/21). It produces store.Step rows for the app layer.
package csvio

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseMMSS parses "mm:ss" (or "h:mm:ss") into seconds.
func ParseMMSS(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("csvio: invalid time %q", s)
	}
	total := 0
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return 0, fmt.Errorf("csvio: invalid time %q", s)
		}
		total = total*60 + n
	}
	return total, nil
}

// FormatMMSS renders seconds as "mm:ss".
func FormatMMSS(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%02d:%02d", sec/60, sec%60)
}
