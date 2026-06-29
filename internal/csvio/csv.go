package csvio

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/radaiko/turnpoint/internal/store"
)

// Options configures CSV parsing/formatting. A zero Delimiter/Decimal triggers
// auto-detection on read (OI-10/21).
type Options struct {
	Delimiter rune // ',' ';' or '\t' (0 ⇒ auto)
	Decimal   rune // '.' or ',' (0 ⇒ auto)
}

// RowError records a rejected row (non-fatal; FR-T6).
type RowError struct {
	Line int    `json:"line"`
	Msg  string `json:"msg"`
}

// ImportReport summarises an import (FR-T6).
type ImportReport struct {
	Imported int        `json:"imported"`
	Skipped  int        `json:"skipped"`
	Errors   []RowError `json:"errors"`
}

// DetectDialect sniffs the delimiter and decimal separator from a sample (OI-10).
func DetectDialect(sample []byte) Options {
	s := string(sample)
	tabs := strings.Count(s, "\t")
	semis := strings.Count(s, ";")
	commas := strings.Count(s, ",")
	switch {
	case tabs > 0 && tabs >= semis:
		return Options{Delimiter: '\t', Decimal: '.'}
	case semis > 0 && semis >= commas:
		// semicolons usually pair with decimal comma (European)
		return Options{Delimiter: ';', Decimal: ','}
	default:
		return Options{Delimiter: ',', Decimal: '.'}
	}
}

var headerAliases = map[string]string{
	"intensity": "intensity", "speed": "intensity", "kmh": "intensity", "km/h": "intensity",
	"watt": "intensity", "watts": "intensity", "power": "intensity", "w": "intensity",
	"time": "time", "t": "time", "timepoint": "time",
	"hr": "hr", "bpm": "hr", "heartrate": "hr", "heart rate": "hr", "puls": "hr",
	"lactate": "lactate", "lac": "lactate", "laktat": "lactate", "lactat": "lactate",
	"rpe": "rpe", "borg": "rpe",
}

// ParseSteps reads steps from r. It maps a tolerant header (case-insensitive,
// bracketed units stripped, aliases) or falls back to canonical column order
// (intensity, time, hr, lactate, rpe) when the first row is numeric. Bad rows are
// reported, not fatal.
func ParseSteps(r io.Reader, opt Options) ([]store.Step, ImportReport, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, ImportReport{}, err
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // strip UTF-8 BOM
	if opt.Delimiter == 0 {
		opt = DetectDialect(data)
	}
	if opt.Decimal == 0 {
		opt.Decimal = '.'
	}

	cr := csv.NewReader(bytes.NewReader(data))
	cr.Comma = opt.Delimiter
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	records, err := cr.ReadAll()
	if err != nil {
		return nil, ImportReport{}, err
	}
	if len(records) == 0 {
		return nil, ImportReport{}, nil
	}

	cols := map[string]int{}
	start := 0
	if isHeader(records[0]) {
		for i, h := range records[0] {
			if key, ok := headerAliases[normalizeHeader(h)]; ok {
				cols[key] = i
			}
		}
		start = 1
	} else {
		cols = map[string]int{"intensity": 0, "time": 1, "hr": 2, "lactate": 3, "rpe": 4}
	}

	var steps []store.Step
	var rep ImportReport
	for li := start; li < len(records); li++ {
		rec := records[li]
		s, err := parseRow(rec, cols, opt.Decimal, li+1)
		if err != nil {
			rep.Skipped++
			rep.Errors = append(rep.Errors, RowError{Line: li + 1, Msg: err.Error()})
			continue
		}
		s.StepOrder = len(steps)
		s.IsBaseline = s.Intensity == 0
		steps = append(steps, s)
		rep.Imported++
	}
	return steps, rep, nil
}

func parseRow(rec []string, cols map[string]int, dec rune, line int) (store.Step, error) {
	var s store.Step
	get := func(key string) (string, bool) {
		if i, ok := cols[key]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i]), true
		}
		return "", false
	}
	iv, ok := get("intensity")
	if !ok || iv == "" {
		return s, fmt.Errorf("missing intensity")
	}
	f, err := parseDecimal(iv, dec)
	if err != nil {
		return s, fmt.Errorf("bad intensity %q", iv)
	}
	s.Intensity = f
	if v, ok := get("time"); ok && v != "" {
		if sec, err := ParseMMSS(v); err == nil {
			s.TimePointS = &sec
		}
	}
	if v, ok := get("hr"); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.HeartRate = &n
		}
	}
	if v, ok := get("lactate"); ok && v != "" {
		if f, err := parseDecimal(v, dec); err == nil {
			s.Lactate = &f
		}
	}
	if v, ok := get("rpe"); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.RPE = &n
		}
	}
	return s, nil
}

// WriteSteps emits canonical-order CSV (intensity, time, hr, lactate, rpe).
func WriteSteps(w io.Writer, steps []store.Step, sport string, opt Options) error {
	if opt.Delimiter == 0 {
		opt.Delimiter = ','
	}
	if opt.Decimal == 0 {
		opt.Decimal = '.'
	}
	cw := csv.NewWriter(w)
	cw.Comma = opt.Delimiter
	unit := "km/h"
	if sport == "cycling" {
		unit = "W"
	}
	_ = cw.Write([]string{"intensity [" + unit + "]", "time", "hr", "lactate [mmol/L]", "rpe"})
	for _, s := range steps {
		rec := []string{
			formatDecimal(s.Intensity, opt.Decimal, 1),
			ptrTime(s.TimePointS),
			ptrInt(s.HeartRate),
			ptrFloat(s.Lactate, opt.Decimal),
			ptrInt(s.RPE),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func isHeader(rec []string) bool {
	for _, f := range rec {
		if _, err := strconv.ParseFloat(strings.TrimSpace(strings.Replace(f, ",", ".", 1)), 64); err == nil {
			return false // a numeric field ⇒ not a header row
		}
		if strings.Contains(f, ":") {
			return false // looks like a time value
		}
	}
	return true
}

func normalizeHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	if i := strings.IndexAny(h, "[("); i >= 0 {
		h = strings.TrimSpace(h[:i]) // drop bracketed unit
	}
	return h
}

func parseDecimal(s string, dec rune) (float64, error) {
	if dec == ',' {
		s = strings.Replace(s, ",", ".", 1)
	}
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func formatDecimal(f float64, dec rune, places int) string {
	s := strconv.FormatFloat(f, 'f', places, 64)
	if dec == ',' {
		s = strings.Replace(s, ".", ",", 1)
	}
	return s
}

func ptrTime(p *int) string {
	if p == nil {
		return ""
	}
	return FormatMMSS(*p)
}
func ptrInt(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}
func ptrFloat(p *float64, dec rune) string {
	if p == nil {
		return ""
	}
	return formatDecimal(*p, dec, 2)
}
