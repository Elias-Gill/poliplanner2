package parser

import (
	"testing"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/commons"
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
