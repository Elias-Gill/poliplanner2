package excel

import (
	"reflect"
	"testing"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

func TestNormalizeSubjectName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Fisica 2", "fisica II"},
		{"calculo 7", "calculo VII"},
		{"Algebra 10", "algebra X"},
		{"programacion 1", "programacion I"},
		{"estadistica 20", "estadistica XX"},

		// Spaces and capital letters
		{"  Fisica   3  ", "fisica III"},
		{"CALCULO 5", "calculo V"},
		{"CáLCULO 5", "calculo V"},

		// No modifications needed
		{"fisica", "fisica"},
		{"fisica II", "fisica II"},
		{"fisica 0", "fisica 0"},
		{"fisica 21", "fisica 21"},

		// Dash delimiters truncation
		{"Electiva 1 - Machine Learning", "electiva I"},
		{"electIVa 2 - quien sabe", "electiva II"},

		// Parenthesis removal
		{"calculo V (variable vectorial)", "calculo V"},
		{"calculo V (*)", "calculo V"},
		{"calculo V (**)", "calculo V"},

		// Additional cases
		{"Álgebra Línea 1", "algebra linea I"},
		{"Óptica Élite 4", "optica elite IV"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeSubjectName(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeSubjectName(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestBuildCareerFromDTO(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected academic.Career
	}{
		{
			name:  "Standard lowercase with spaces",
			input: "  inf  ",
			expected: academic.Career{
				Code: "INF",
			},
		},
		{
			name:  "Clean uppercase",
			input: "ECA",
			expected: academic.Career{
				Code: "ECA",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCareerFromDTO(tc.input)
			if got != tc.expected {
				t.Errorf("buildCareerFromDTO(%q) = %+v; want %+v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestBuildSubject(t *testing.T) {
	input := parser.SubjectDTO{
		RawSubjectName: "  Química Orgánica 2 - Semestre Común ",
		Department:     "DEPT-QMC",
	}

	expected := academic.Subject{
		Name: "quimica organica II",
		Department: academic.Department{
			Code: "DEPT-QMC",
		},
	}

	got := buildSubject(input)
	if got != expected {
		t.Errorf("buildSubject() = %+v; want %+v", got, expected)
	}
}

func TestBuildCurriculum(t *testing.T) {
	input := parser.SubjectDTO{
		Level:    4,
		Semester: 5,
	}

	expected := academic.Curriculum{
		Level:    4,
		Semester: 5,
	}

	got := buildCurriculum(input)
	// FIX: test case for emphasis
	if got != expected {
		t.Errorf("buildCurriculum() = %+v; want %+v", got, expected)
	}
}

func TestBuildOfferingFromDTO_ExamsMapping(t *testing.T) {
	// Modificado: Ahora pasamos structs por valor asignando de forma explícita Valid: true
	input := parser.SubjectDTO{
		RawSubjectName: "Matematica I",
		CourseType:     academic.ExamOnly,
		Section:        "A",
		Partial1Date:   parser.Date{Year: 2026, Month: 5, Day: 10, Valid: true},
		Partial1Time:   parser.Hour{Hour: 14, Minute: 30, Valid: true},
		Partial1Room:   "Aula 1",
		Final1Date:     parser.Date{Year: 2026, Month: 7, Day: 15, Valid: true},
		Final1Time:     parser.Hour{Hour: 8, Minute: 0, Valid: true},
		Final1RevDate:  parser.Date{Year: 2026, Month: 7, Day: 18, Valid: true},
		Final1RevTime:  parser.Hour{Hour: 10, Minute: 0, Valid: true},
		Final1Room:     "Aula Magna",
		// Los campos omitidos (Partial2, Final2) se inicializan en cero por defecto con Valid: false
	}

	got := buildOfferingFromDTO(input)

	if got.Name != "Matematica I" || got.Section != "A" || got.Type != academic.ExamOnly {
		t.Errorf("Basic fields mapping failed: %+v", got)
	}

	if len(got.Exams) != 2 {
		t.Fatalf("Expected exactly 2 mapped exams (Partial1 and Final1), got %d", len(got.Exams))
	}

	p1 := got.Exams[0]
	if p1.Type != academic.ExamPartial || p1.Instance != 1 || p1.Room != "Aula 1" {
		t.Errorf("Partial1 basic fields mismatch: %+v", p1)
	}
	expectedP1Date := time.Date(2026, time.May, 10, 14, 30, 0, 0, timezone.ParaguayTZ)
	if !reflect.DeepEqual(p1.Date(), &expectedP1Date) {
		t.Errorf("Partial1 date mismatch: %v; want %v", p1.Date(), expectedP1Date)
	}

	f1 := got.Exams[1]
	if f1.Type != academic.ExamFinal || f1.Instance != 1 || f1.Room != "Aula Magna" {
		t.Errorf("Final1 basic fields mismatch: %+v", f1)
	}
	expectedF1Date := time.Date(2026, time.July, 15, 8, 0, 0, 0, timezone.ParaguayTZ)
	expectedF1Rev := time.Date(2026, time.July, 18, 10, 0, 0, 0, timezone.ParaguayTZ)

	if !reflect.DeepEqual(f1.Date(), &expectedF1Date) {
		t.Errorf("Final1 date mismatch: %v; want %v", f1.Date(), expectedF1Date)
	}
	if !reflect.DeepEqual(f1.Revision(), &expectedF1Rev) {
		t.Errorf("Final1 revision date mismatch: %v; want %v", f1.Revision(), expectedF1Rev)
	}
}
