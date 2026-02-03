package parser

import (
	"testing"
	"time"
)

func TestParseSheet(t *testing.T) {
	testExcelPath := "../../testdata/excelparser/testExcel.xlsx"
	parser, err := NewExcelParser("./layouts", testExcelPath)
	if err != nil {
		t.Fatalf("Failed to initialize parser: %v", err)
	}

	start := time.Now()
	foundSheet := false
	var iin *ParsingResult
	for parser.NextSheet() {
		localStart := time.Now()
		result, err := parser.ParseCurrentSheet()
		localEnd := time.Now()
		t.Logf("Parsing %s, duration: %dms", result.Career, localEnd.Sub(localStart).Milliseconds())
		if err != nil {
			t.Fatalf("Failed to parse sheet: %v", err)
		}
		if result.Career == "IIN" {
			foundSheet = true
			iin = result
		}
	}
	end := time.Now()
	t.Logf("Total parsing duration: %dms", end.Sub(start).Milliseconds())

	if !foundSheet {
		t.Fatal("Cannot find 'IIN' sheet inside testExcel.xlsx")
	}
	validateParsingResult(t, iin)
}

func TestNameNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Fisica 2", "fisica II"},
		{"calculo 7", "calculo VII"},
		{"Algebra 10", "algebra X"},
		{"programacion 1", "programacion I"},
		{"estadistica 20", "estadistica XX"},

		// Espacios y mayúsculas
		{"  Fisica   3  ", "fisica III"},
		{"CALCULO 5", "calculo V"},
		{"CáLCULO 5", "calculo V"},

		// No se debe normalizar
		{"fisica", "fisica"},
		{"fisica II", "fisica II"},
		{"fisica 0", "fisica 0"},
		{"fisica 21", "fisica 21"},

		{"Electiva 1 - Machine Learning", "electiva I"},
		{"electIVa 2 - quien sabe", "electiva II"},
	}

	for _, testCase := range tests {
		t.Run(testCase.input, func(t *testing.T) {
			dto := SubjectDTO{}
			dto.SetSubjectName(testCase.input)

			if dto.TentativeRealSubjectName != testCase.expected {
				t.Errorf(
					"SetSubjectName(%q) => %q, expected %q",
					testCase.input,
					dto.TentativeRealSubjectName,
					testCase.expected,
				)
			}
		})
	}
}

func validateParsingResult(t *testing.T, result *ParsingResult) {
	if result == nil {
		t.Fatal("Parsing result should not be nil")
	}
	if result.Subjects == nil {
		t.Fatal("Subjects list should not be nil")
	}
	if len(result.Subjects) == 0 {
		t.Fatal("Subjects list should not be empty")
	}
	if len(result.Subjects) != 243 {
		t.Errorf("Subjects parsed are = %d, want %d", 243, len(result.Subjects))
	}

	// Primera entrada
	first := result.Subjects[0]
	if first.Department != "DCB" {
		t.Errorf("First subject Department = %s, want %s", first.Department, "DCB")
	}
	if first.RawSubjectName != "Algebra Lineal" {
		t.Errorf("First subject SubjectName = %s, want %s", first.RawSubjectName, "Algebra Lineal")
	}
	if first.Semester != 2 {
		t.Errorf("First subject Semester = %d, want %d", first.Semester, 2)
	}
	if first.Section != "MI" {
		t.Errorf("First subject Section = %s, want %s", first.Section, "MI")
	}

	if len(first.Teachers) != 1 {
		t.Errorf("Last subject teachers length = %s, want %d", first.Teachers[0].LastName, 1)
	}
	if first.Teachers[0].LastName != "Villasanti Flores" {
		t.Errorf("First subject TeacherLastName = %s, want %s", first.Teachers[0].LastName, "Villasanti Flores")
	}
	if first.Teachers[0].FirstName != "Richard Adrián" {
		t.Errorf("First subject TeacherName = %s, want %s", first.Teachers[0].FirstName, "Richard Adrián")
	}
	if first.Teachers[0].Email != "" {
		t.Errorf("First subject TeacherEmail = %s, want empty", first.Teachers[0].Email)
	}

	// Última entrada
	last := result.Subjects[len(result.Subjects)-1]
	validateLastSubject(t, last)
}

func validateLastSubject(t *testing.T, last SubjectDTO) {
	if last.Department != "DG" {
		t.Errorf("Last subject Department = %s, want %s", last.Department, "DG")
	}
	if last.RawSubjectName != "Técnicas de Organización y Métodos" {
		t.Errorf("Last subject SubjectName = %s, want %s", last.RawSubjectName, "Técnicas de Organización y Métodos")
	}
	if last.Semester != 5 {
		t.Errorf("Last subject Semester = %d, want %d", last.Semester, 5)
	}
	if last.Section != "NA" {
		t.Errorf("Last subject Section = %s, want %s", last.Section, "NA")
	}

	if len(last.Teachers) != 1 {
		t.Errorf("Last subject teachers length = %s, want %d", last.Teachers[0].LastName, 1)
	}
	if last.Teachers[0].LastName != "Ramírez Barboza" {
		t.Errorf("Last subject TeacherLastName = %s, want %s", last.Teachers[0].LastName, "Ramírez Barboza")
	}
	if last.Teachers[0].FirstName != "Estela Mary" {
		t.Errorf("Last subject TeacherName = %s, want %s", last.Teachers[0].FirstName, "Estela Mary")
	}
	if last.Teachers[0].Email != "emramirez@pol.una.py" {
		t.Errorf("Last subject TeacherEmail = %s, want %s", last.Teachers[0].Email, "emramirez@pol.una.py")
	}

	// Validar fecha del primer parcial (17/09/24)
	expectedDate := time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC)
	if last.Partial1Date == nil || !last.Partial1Date.Equal(expectedDate) {
		t.Errorf("Last subject Partial1Date = %v, want %v", last.Partial1Date, expectedDate)
	}

	// Validar comité
	if last.CommitteePresident != "Ms. Estela Mary Ramírez Barboza" {
		t.Errorf("Last subject CommitteePresident = %s, want %s", last.CommitteePresident, "Ms. Estela Mary Ramírez Barboza")
	}
	if last.CommitteeMember1 != "Lic. Zulma Lucía Demattei Ortíz" {
		t.Errorf("Last subject CommitteeMember1 = %s, want %s", last.CommitteeMember1, "Lic. Zulma Lucía Demattei Ortíz")
	}
	if last.CommitteeMember2 != "Lic. Osvaldo David Sosa Cabrera" {
		t.Errorf("Last subject CommitteeMember2 = %s, want %s", last.CommitteeMember2, "Lic. Osvaldo David Sosa Cabrera")
	}

	assertTimeSlot(t, last.Tuesday, "20:45", "22:15")
	assertTimeSlot(t, last.Thursday, "19:00", "20:30")
	assertTimeSlot(t, last.Saturday, "07:30", "11:30")

	if last.SaturdayDates != "05/10, 23/11" {
		t.Errorf("Last subject SaturdayDates = %s, want %s", last.SaturdayDates, "05/10, 23/11")
	}

	if last.WednesdayRoom != "" {
		t.Errorf("Last subject WednesdayRoom = %s, want empty", last.WednesdayRoom)
	}
}

func assertTimeSlot(t *testing.T, slot TimeSlot, wantStart, wantEnd string) {
	if slot.Start != wantStart || slot.End != wantEnd {
		t.Errorf("got %s-%s, want %s-%s", slot.Start, slot.End, wantStart, wantEnd)
	}
}
