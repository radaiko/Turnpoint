package csvio

import (
	"bytes"
	"strings"
	"testing"
)

const appendixACSV = `intensity [km/h],time,hr,lactate [mmol/L],rpe
0,00:00,0,0.00,6
6,03:00,98,1.24,6
8,06:00,111,1.19,6
20,22:10,185,7.74,6
`

func TestParseAppendixA(t *testing.T) {
	steps, rep, err := ParseSteps(strings.NewReader(appendixACSV), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Imported != 4 || rep.Skipped != 0 {
		t.Fatalf("report = %+v", rep)
	}
	if steps[0].Intensity != 0 || !steps[0].IsBaseline {
		t.Errorf("baseline row wrong: %+v", steps[0])
	}
	if steps[3].Intensity != 20 || steps[3].TimePointS == nil || *steps[3].TimePointS != 1330 {
		t.Errorf("last row wrong: %+v", steps[3])
	}
	if steps[1].Lactate == nil || *steps[1].Lactate != 1.24 {
		t.Errorf("lactate parse wrong: %+v", steps[1])
	}
}

func TestDetectDialectAndDecimalComma(t *testing.T) {
	// European: semicolon delimiter, decimal comma
	euro := "intensity;time;hr;lactate;rpe\n6;03:00;98;1,24;6\n8;06:00;111;1,19;6\n"
	steps, rep, err := ParseSteps(strings.NewReader(euro), Options{})
	if err != nil || rep.Imported != 2 {
		t.Fatalf("euro parse: %+v %v", rep, err)
	}
	if *steps[0].Lactate != 1.24 {
		t.Errorf("decimal-comma parse wrong: %v", *steps[0].Lactate)
	}
}

func TestParseTabSeparatedPaste(t *testing.T) {
	tsv := "6\t03:00\t98\t1.24\t6\n8\t06:00\t111\t1.19\t6\n" // headerless paste
	steps, rep, err := ParseSteps(strings.NewReader(tsv), Options{})
	if err != nil || rep.Imported != 2 {
		t.Fatalf("tsv parse: %+v %v", rep, err)
	}
	if steps[0].Intensity != 6 || *steps[0].HeartRate != 98 {
		t.Errorf("headerless TSV wrong: %+v", steps[0])
	}
}

func TestRoundTrip(t *testing.T) {
	steps, _, _ := ParseSteps(strings.NewReader(appendixACSV), Options{})
	var buf bytes.Buffer
	if err := WriteSteps(&buf, steps, "running", Options{}); err != nil {
		t.Fatal(err)
	}
	steps2, _, err := ParseSteps(&buf, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(steps2) != len(steps) {
		t.Fatalf("round-trip lost rows: %d → %d", len(steps), len(steps2))
	}
	if steps2[3].Intensity != 20 {
		t.Errorf("round-trip value wrong: %+v", steps2[3])
	}
}
