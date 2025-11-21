package parser

import (
	"testing"
	"time"

	"github.com/elias-gill/poliplanner2/internal/excelparser/dto"
)

func TestParseSheet(t *testing.T) {
    testExcelPath := "../../testdata/excelparser/testExcel.xlsx"
    parser, err := NewExcelParser("./layouts")
    if err != nil {
        t.Fatalf("Failed to create parser: %v", err)
    }
    err = parser.ParseExcel(testExcelPath)
    if err != nil {
        t.Fatalf("Failed to parse Excel file: %v", err)
    }

    start := time.Now()
    foundSheet := false
	var iin *ParsingResult
    for parser.NextValidSheet() {
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
	if first.SubjectName != "Algebra Lineal" {
		t.Errorf("First subject SubjectName = %s, want %s", first.SubjectName, "Algebra Lineal")
	}
	if first.Semester != 2 {
		t.Errorf("First subject Semester = %d, want %d", first.Semester, 2)
	}
	if first.Section != "MI" {
		t.Errorf("First subject Section = %s, want %s", first.Section, "MI")
	}
	if first.TeacherTitle != "Lic." {
		t.Errorf("First subject TeacherTitle = %s, want %s", first.TeacherTitle, "Lic.")
	}
	if first.TeacherLastName != "Villasanti Flores" {
		t.Errorf("First subject TeacherLastName = %s, want %s", first.TeacherLastName, "Villasanti Flores")
	}
	if first.TeacherName != "Richard Adrián" {
		t.Errorf("First subject TeacherName = %s, want %s", first.TeacherName, "Richard Adrián")
	}
	if first.TeacherEmail != "" {
		t.Errorf("First subject TeacherEmail = %s, want empty", first.TeacherEmail)
	}

	// Última entrada
	last := result.Subjects[len(result.Subjects)-1]
	validateLastSubject(t, last)
}

func validateLastSubject(t *testing.T, last dto.SubjectDTO) {
	if last.Department != "DG" {
		t.Errorf("Last subject Department = %s, want %s", last.Department, "DG")
	}
	if last.SubjectName != "Técnicas de Organización y Métodos" {
		t.Errorf("Last subject SubjectName = %s, want %s", last.SubjectName, "Técnicas de Organización y Métodos")
	}
	if last.Semester != 5 {
		t.Errorf("Last subject Semester = %d, want %d", last.Semester, 5)
	}
	if last.Section != "NA" {
		t.Errorf("Last subject Section = %s, want %s", last.Section, "NA")
	}
	if last.TeacherTitle != "Ms." {
		t.Errorf("Last subject TeacherTitle = %s, want %s", last.TeacherTitle, "Ms.")
	}
	if last.TeacherLastName != "Ramírez Barboza" {
		t.Errorf("Last subject TeacherLastName = %s, want %s", last.TeacherLastName, "Ramírez Barboza")
	}
	if last.TeacherName != "Estela Mary" {
		t.Errorf("Last subject TeacherName = %s, want %s", last.TeacherName, "Estela Mary")
	}
	if last.TeacherEmail != "emramirez@pol.una.py" {
		t.Errorf("Last subject TeacherEmail = %s, want %s", last.TeacherEmail, "emramirez@pol.una.py")
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

	// Validar horarios y aulas
	if last.TuesdayRoom != "E01" {
		t.Errorf("Last subject TuesdayRoom = %s, want %s", last.TuesdayRoom, "E01")
	}
	if last.Tuesday != "20:45 - 22:15" {
		t.Errorf("Last subject Tuesday = %s, want %s", last.Tuesday, "20:45 - 22:15")
	}

	if last.ThursdayRoom != "E01" {
		t.Errorf("Last subject ThursdayRoom = %s, want %s", last.ThursdayRoom, "E01")
	}
	if last.Thursday != "19:00 - 20:30" {
		t.Errorf("Last subject Thursday = %s, want %s", last.Thursday, "19:00 - 20:30")
	}

	if last.SaturdayRoom != "E01" {
		t.Errorf("Last subject SaturdayRoom = %s, want %s", last.SaturdayRoom, "E01")
	}
	if last.Saturday != "07:30 - 11:30" {
		t.Errorf("Last subject Saturday = %s, want %s", last.Saturday, "07:30 - 11:30")
	}

	if last.SaturdayDates != "05/10, 23/11" {
		t.Errorf("Last subject SaturdayDates = %s, want %s", last.SaturdayDates, "05/10, 23/11")
	}

	if last.WednesdayRoom != "" {
		t.Errorf("Last subject WednesdayRoom = %s, want empty", last.WednesdayRoom)
	}
}
