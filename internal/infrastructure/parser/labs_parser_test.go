package parser

import (
	"testing"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/commons"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/engine"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

func labHour(h, m int) commons.Hour {
	return commons.Hour{Hour: h, Minute: m, Valid: true}
}

func labSlot(h1, m1, h2, m2 int) commons.TimeSlot {
	return commons.TimeSlot{Start: labHour(h1, m1), End: labHour(h2, m2)}
}

func TestParseDayCell(t *testing.T) {
	tests := []struct {
		name string
		cell string
		want []dayEntry
	}{
		{
			name: "single section at the end",
			cell: "18:00 - 20:00 (T1)",
			want: []dayEntry{{Section: "T1", Time: labSlot(18, 0, 20, 0)}},
		},
		{
			name: "section at the beginning",
			cell: "(T1) 18:00 - 20:00",
			want: []dayEntry{{Section: "T1", Time: labSlot(18, 0, 20, 0)}},
		},
		{
			name: "multiple sections, one per line",
			cell: "18:00 - 20:00 (T1)\n20:00 - 22:00 (T2)\n07:30 - 9:00 (T3)",
			want: []dayEntry{
				{Section: "T1", Time: labSlot(18, 0, 20, 0)},
				{Section: "T2", Time: labSlot(20, 0, 22, 0)},
				{Section: "T3", Time: labSlot(7, 30, 9, 0)},
			},
		},
		{
			name: "no section falls back to default",
			cell: "18:00 - 22:00hs",
			want: []dayEntry{{Section: defaultSection, Time: labSlot(18, 0, 22, 0)}},
		},
		{
			name: "section content is preserved as-is",
			cell: "18:00 - 20:00 (seccion_1)",
			want: []dayEntry{{Section: "seccion_1", Time: labSlot(18, 0, 20, 0)}},
		},
		{
			name: "ignore lines without a valid range",
			cell: "sin horario\n18:00 - 20:00 (T1)",
			want: []dayEntry{{Section: "T1", Time: labSlot(18, 0, 20, 0)}},
		},
		{
			name: "empty cell produces nothing",
			cell: "",
			want: nil,
		},
		{
			name: "blank lines are skipped",
			cell: "\n18:00 - 20:00 (T1)\n\n",
			want: []dayEntry{{Section: "T1", Time: labSlot(18, 0, 20, 0)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDayCell(tt.cell)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries %+v, want %d %+v", len(got), got, len(tt.want), tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("entry %d mismatch: got %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitSection(t *testing.T) {
	tests := []struct {
		line        string
		wantSection string
		wantHour    string
	}{
		{"18:00 - 20:00 (T1)", "T1", "18:00 - 20:00"},
		{"(T1) 18:00 - 20:00", "T1", "18:00 - 20:00"},
		{"18:00 - 20:00", "", "18:00 - 20:00"},
		{"18:00 - 20:00hs (Lab 1)", "Lab 1", "18:00 - 20:00hs"},
		{"(seccion_1)", "seccion_1", ""},
		{"18:00 - 20:00 (broken", "", "18:00 - 20:00 (broken"},
	}

	for _, tt := range tests {
		gotSection, gotHour := splitSection(tt.line)
		if gotSection != tt.wantSection || gotHour != tt.wantHour {
			t.Errorf("splitSection(%q) = (%q, %q), want (%q, %q)",
				tt.line, gotSection, gotHour, tt.wantSection, tt.wantHour)
		}
	}
}

func TestAccumulatorDropsForeignCareers(t *testing.T) {
	lay := &engine.Layout{Headers: []string{"asignatura", "plan", "carrera", "horaLunes", "horaMartes"}}
	acc := newLabAccumulator(commons.NormalizeCareer("IIN"))

	acc.parseRow([]string{"Algebra", "2010", "IIN", "18:00 - 20:00 (T1)", ""}, lay, 0)
	acc.parseRow([]string{"Otra Materia", "2010", "LCI", "18:00 - 20:00 (T1)", ""}, lay, 0)
	// Same subject/plan/section as the valid entry but from a foreign career: it must not
	// contaminate the valid one.
	acc.parseRow([]string{"Algebra", "2010", "LCI", "", "20:00 - 22:00 (T1)"}, lay, 0)

	labs := acc.flatten()
	if len(labs) != 1 {
		t.Fatalf("got %d labs, want 1: %+v", len(labs), labs)
	}
	if labs[0].RawName != "Algebra" || labs[0].Section != "T1" {
		t.Fatalf("unexpected lab: %+v", labs[0])
	}
	if labs[0].WeekSchedule[academic.Tuesday].Time.Start.Valid {
		t.Fatalf("foreign career row contaminated the valid entry: %+v", labs[0].WeekSchedule)
	}
}

func TestAccumulatorCarriesCareerForward(t *testing.T) {
	lay := &engine.Layout{Headers: []string{"asignatura", "plan", "carrera", "horaLunes"}}
	acc := newLabAccumulator(commons.NormalizeCareer("IIN"))

	acc.parseRow([]string{"Algebra", "2010", "LCI", "18:00 - 20:00 (T1)"}, lay, 0)
	// Empty career cell (merged) inherits the previous value, so it is discarded as foreign.
	acc.parseRow([]string{"Algebra II", "2010", "", "18:00 - 20:00 (T1)"}, lay, 0)

	if labs := acc.flatten(); len(labs) != 0 {
		t.Fatalf("got %d labs, want 0: %+v", len(labs), labs)
	}
}

func TestAccumulatorSplitsSectionsBySchedule(t *testing.T) {
	lay := &engine.Layout{Headers: []string{"asignatura", "plan", "carrera", "horaLunes"}}
	acc := newLabAccumulator(commons.NormalizeCareer("IIN"))

	acc.parseRow([]string{"Fisica", "2010", "IIN", "18:00 - 20:00 (T1)"}, lay, 0)
	acc.parseRow([]string{"Fisica", "2010", "IIN", "20:00 - 22:00 (T1)"}, lay, 0)

	labs := acc.flatten()
	if len(labs) != 2 {
		t.Fatalf("got %d labs, want 2: %+v", len(labs), labs)
	}
	if labs[0].Section != "T1" || labs[1].Section != "T1_2" {
		t.Fatalf("section names = %q, %q, want T1, T1_2", labs[0].Section, labs[1].Section)
	}
	if got := labs[0].WeekSchedule[academic.Monday].Time.Start.Hour; got != 18 {
		t.Fatalf("first section start hour = %d, want 18", got)
	}
	if got := labs[1].WeekSchedule[academic.Monday].Time.Start.Hour; got != 20 {
		t.Fatalf("second section start hour = %d, want 20", got)
	}
}

func TestAccumulatorNamesUnnamedSections(t *testing.T) {
	lay := &engine.Layout{Headers: []string{"asignatura", "plan", "carrera", "horaLunes"}}
	acc := newLabAccumulator(commons.NormalizeCareer("IIN"))

	acc.parseRow([]string{"Fisica", "2010", "IIN", "18:00 - 20:00"}, lay, 0)
	acc.parseRow([]string{"Fisica", "2010", "IIN", "20:00 - 22:00"}, lay, 0)

	labs := acc.flatten()
	if len(labs) != 2 {
		t.Fatalf("got %d labs, want 2: %+v", len(labs), labs)
	}
	if labs[0].Section != defaultSection || labs[1].Section != defaultSection+"_2" {
		t.Fatalf("section names = %q, %q, want %s, %s_2",
			labs[0].Section, labs[1].Section, defaultSection, defaultSection)
	}
}

func TestAccumulatorMergesIdenticalSchedules(t *testing.T) {
	lay := &engine.Layout{Headers: []string{"asignatura", "plan", "carrera", "horaLunes", "horaMartes"}}
	acc := newLabAccumulator(commons.NormalizeCareer("IIN"))

	acc.parseRow([]string{"Fisica", "2010", "IIN", "18:00 - 20:00 (T1)", "20:00 - 22:00 (T1)"}, lay, 0)
	acc.parseRow([]string{"Fisica", "2010", "IIN", "18:00 - 20:00 (T1)", "20:00 - 22:00 (T1)"}, lay, 0)

	if labs := acc.flatten(); len(labs) != 1 {
		t.Fatalf("got %d labs, want 1: %+v", len(labs), labs)
	}
}

func TestAccumulatorCapturesPeriodo(t *testing.T) {
	lay := &engine.Layout{Headers: []string{"asignatura", "plan", "carrera", "periodo", "horaLunes"}}
	acc := newLabAccumulator(commons.NormalizeCareer("IIN"))

	acc.parseRow([]string{"Fisica", "2010", "IIN", "2", "18:00 - 20:00 (T1)"}, lay, 0)

	labs := acc.flatten()
	if len(labs) != 1 || labs[0].Semester != 1 {
		t.Fatalf("periodo not captured: %+v", labs)
	}
}
