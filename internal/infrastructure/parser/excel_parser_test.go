package parser

import (
	"os"
	"testing"
)

func TestParseSubjects_ByCareerSheets(t *testing.T) {
	file, err := os.Open("../../../test_data/excel/stripped_excel.xlsx")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	parser, err := NewExcelParser(file)
	if err != nil {
		t.Fatalf("cannot create parser: %v", err)
	}
	defer parser.Close()

	expectedOrder := []string{"IIN", "CNEL.OVIEDO", "VILLARRICA", "LCIK", "ISP"}

	expectedSubjectsByCareer := map[string][]SubjectDTO{
		"IIN": {
			subject(
				withGeneral("DCB", 2, "MI", "IIN", "Algebra Lineal"),
				withTeachers(teacher("Lic.", "Richard Adrián", "Villasanti Flores", "")),
				func(s *SubjectDTO) {
					s.Partial1Date = datePtr(2024, 9, 17)
					s.Partial1Time = hourPtr(8, 0)
					s.Partial1Room = "A50"
					s.Partial2Date = datePtr(2024, 11, 12)
					s.Partial2Time = hourPtr(8, 0)
					s.Partial2Room = "A50"
					s.Final1Date = datePtr(2024, 12, 3)
					s.Final1Time = hourPtr(8, 0)
					s.Final1Room = "A50"
					s.Final1RevDate = datePtr(2024, 12, 10)
					s.Final1RevTime = hourPtr(10, 30)
					s.Final2Date = datePtr(2024, 12, 17)
					s.Final2Time = hourPtr(8, 0)
					s.Final2Room = "A50"
					s.Final2RevDate = datePtr(2024, 12, 26)
					s.Final2RevTime = hourPtr(10, 30)
					s.CommitteePresident = "Lic. Richard Adrián Villasanti Flores"
					s.CommitteeMember1 = "Ms. Osvaldo Ramón Vega Gamarra"
					s.CommitteeMember2 = "Ms. Édgar López Pezoa"
					s.Schedule = map[WeekDay]WeekDayData{
						Wednesday: {Room: "C01", Time: slot(hourPtr(9, 15), hourPtr(12, 15))},
						Friday:    {Room: "C01", Time: slot(hourPtr(10, 0), hourPtr(12, 15))},
					}
					s.SaturdayDates = "24/03"
				},
			),
			subject(
				withGeneral("DCB", 3, "NB", "IIN", "Cálculo II"),
				withTeachers(teacher("Lic.", "Silvia Verónica", "Chamorro Hermosa", "schamorro@pol.una.py")),
				func(s *SubjectDTO) {
					s.Partial1Date = datePtr(2024, 9, 9)
					s.Partial1Time = hourPtr(19, 30)
					s.Partial1Room = "A57"
					s.Partial2Date = datePtr(2024, 11, 4)
					s.Partial2Time = hourPtr(19, 30)
					s.Partial2Room = "A59"
					s.Final1Date = datePtr(2024, 11, 25)
					s.Final1Time = hourPtr(19, 30)
					s.Final1Room = "A57"
					s.Final1RevDate = datePtr(2024, 12, 4)
					s.Final1RevTime = hourPtr(19, 0)
					s.Final2Date = datePtr(2024, 12, 23)
					s.Final2Time = hourPtr(19, 30)
					s.Final2Room = "A57"
					s.Final2RevDate = datePtr(2024, 12, 27)
					s.Final2RevTime = hourPtr(19, 0)
					s.CommitteePresident = "Lic. Silvia Verónica Chamorro Hermosa"
					s.CommitteeMember1 = "Ms. Rubén Dario Zárate Rojas"
					s.CommitteeMember2 = "Lic. Pamela Raquel Flores Acosta"
					s.Schedule = map[WeekDay]WeekDayData{
						Monday:    {Room: "A57", Time: slot(hourPtr(20, 0), hourPtr(22, 15))},
						Wednesday: {Room: "A57", Time: slot(hourPtr(20, 0), hourPtr(22, 15))},
					}
				},
			),
			subject(
				withGeneral("DG", 1, "MJ", "IIN", "Contabilidad (*)"),
				withTeachers(teacher("C.P.", "Leidy Jessica", "Ríos Argaña", "ljrios@pol.una.py")),
				func(s *SubjectDTO) {
					s.Final1Date = datePtr(2024, 12, 4)
					s.Final1Time = hourPtr(8, 0)
					s.Final1Room = "C03"
					s.Final1RevDate = datePtr(2024, 12, 13)
					s.Final1RevTime = hourPtr(10, 30)
					s.Final2Date = datePtr(2024, 12, 18)
					s.Final2Time = hourPtr(8, 0)
					s.Final2Room = "C03"
					s.Final2RevDate = datePtr(2024, 12, 27)
					s.Final2RevTime = hourPtr(10, 30)
					s.CommitteePresident = "C.P. Leidy Jessica Ríos Argaña"
					s.CommitteeMember1 = "Ms. Cynthia Elizabeth Segovia Domínguez"
					s.CommitteeMember2 = "Lic. Zulma Lucía Demattei Ortiz"
				},
			),
			subject(
				withGeneral("DEI", 10, "TQ", "IIN", "Diseño de Compiladores"),
				withTeachers(teacher("Ing.", "Sergio Andrés", "Aranda Zemán", "saranda@pol.una.py")),
				func(s *SubjectDTO) {
					s.Final1Date = datePtr(2024, 12, 13)
					s.Final1Time = hourPtr(18, 0)
					s.Final1Room = "F38"
					s.Final1RevDate = datePtr(2024, 12, 21)
					s.Final1RevTime = hourPtr(10, 30)
					s.Final2Date = datePtr(2024, 12, 28)
					s.Final2Time = hourPtr(8, 0)
					s.Final2Room = "F38"
					s.Final2RevDate = datePtr(2024, 12, 30)
					s.Final2RevTime = hourPtr(17, 0)
					s.CommitteePresident = "Ing. Sergio Andrés Aranda Zemán"
					s.CommitteeMember1 = "Ing. Deysi Natalia Leguizamón Correa"
					s.CommitteeMember2 = "Ing. Fernando Ramón Saucedo Arguello"
					s.Schedule = map[WeekDay]WeekDayData{
						Wednesday: {Room: "I06", Time: slot(hourPtr(16, 0), hourPtr(19, 0))},
						Friday:    {Room: "F39", Time: slot(hourPtr(18, 30), hourPtr(20, 45))},
					}
				},
			),
		},
		"ISP": {
			subject(
				withGeneral("DG", 2, "TQ", "ISP", "Administración de Empresas (*)"),
				withTeachers(teacher("Lic.", "Zulma Lucía", "Demattei Ortiz", "zdemattei@pol.una.py")),
				func(s *SubjectDTO) {
					s.Final1Date = datePtr(2024, 12, 13)
					s.Final1Time = hourPtr(18, 0)
					s.Final1Room = "C03"
					s.Final1RevDate = datePtr(2024, 11, 27)
					s.Final1RevTime = hourPtr(18, 0)
					s.Final2Date = datePtr(2024, 12, 23)
					s.Final2Time = hourPtr(17, 0)
					s.Final2RevDate = datePtr(2024, 12, 6)
					s.Final2RevTime = hourPtr(17, 0)
					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortiz"
					s.CommitteeMember1 = "Lic. Julio Ramón Riveros Báez"
					s.CommitteeMember2 = "Lic. Osvaldo David Sosa Cabrera"
				},
			),
		},
	}

	sheetIndex := 0
	for parser.NextSheet() {
		if sheetIndex >= len(expectedOrder) {
			t.Fatalf("more sheets than expected")
		}

		parsed, err := parser.ParseCurrentSheet()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		expectedName := expectedOrder[sheetIndex]
		if parsed.Name != expectedName {
			t.Fatalf("sheet order mismatch: got %q want %q", parsed.Name, expectedName)
		}

		expectedSubjects := expectedSubjectsByCareer[expectedName]
		if len(parsed.Subjects) != len(expectedSubjects) {
			t.Fatalf("%s subjects count mismatch: got %d want %d", expectedName, len(parsed.Subjects), len(expectedSubjects))
		}

		for i := range expectedSubjects {
			assertSubjectEqual(t, parsed.Subjects[i], expectedSubjects[i])
		}

		sheetIndex++
	}

	if sheetIndex != len(expectedOrder) {
		t.Fatalf("missing sheets: got %d want %d", sheetIndex, len(expectedOrder))
	}
}

// Helpers

func teacher(title, first, last, email string) TeacherDTO {
	return TeacherDTO{Title: title, FirstName: first, LastName: last, Email: email}
}

func slot(start, end *Hour) TimeSlot {
	return TimeSlot{Start: start, End: end}
}

func datePtr(year, month, day int) *Date {
	return &Date{Year: year, Month: month, Day: day}
}

func hourPtr(hour, minute int) *Hour {
	return &Hour{Hour: hour, Minute: minute}
}

func subject(opts ...func(*SubjectDTO)) SubjectDTO {
	s := SubjectDTO{}
	for _, fn := range opts {
		fn(&s)
	}
	return s
}

func withGeneral(dept string, semester int, section, career, raw string) func(*SubjectDTO) {
	return func(s *SubjectDTO) {
		s.Department = dept
		s.Semester = semester
		s.Section = section
		s.RawSubjectName = raw
	}
}

func withTeachers(t ...TeacherDTO) func(*SubjectDTO) {
	return func(s *SubjectDTO) { s.Teachers = t }
}

func assertSubjectEqual(t *testing.T, got, want SubjectDTO) {
	if got.Department != want.Department {
		t.Errorf("Department mismatch: got %q, want %q", got.Department, want.Department)
	}
	if got.Semester != want.Semester {
		t.Errorf("Semester mismatch: got %d, want %d", got.Semester, want.Semester)
	}
	if got.Section != want.Section {
		t.Errorf("Section mismatch: got %q, want %q", got.Section, want.Section)
	}
	if got.RawSubjectName != want.RawSubjectName {
		t.Errorf("RawSubjectName mismatch: got %q, want %q", got.RawSubjectName, want.RawSubjectName)
	}

	// Comparar profesores
	if len(got.Teachers) != len(want.Teachers) {
		t.Errorf("Teachers count mismatch: got %d, want %d", len(got.Teachers), len(want.Teachers))
	} else {
		for i := range got.Teachers {
			if got.Teachers[i] != want.Teachers[i] {
				t.Errorf("Teacher[%d] mismatch: got %+v, want %+v", i, got.Teachers[i], want.Teachers[i])
			}
		}
	}

	// Comparar fechas de parciales y finales
	compareDates := func(field string, gotDate, wantDate *Hour) {
		if gotDate == nil && wantDate == nil {
			return
		}
		if gotDate == nil || wantDate == nil || *gotDate != *wantDate {
			t.Errorf("%s mismatch: got %+v, want %+v", field, gotDate, wantDate)
		}
	}

	compareDates("Partial1Time", got.Partial1Time, want.Partial1Time)
	compareDates("Partial2Time", got.Partial2Time, want.Partial2Time)
	compareDates("Final1Time", got.Final1Time, want.Final1Time)
	compareDates("Final1RevTime", got.Final1RevTime, want.Final1RevTime)
	compareDates("Final2Time", got.Final2Time, want.Final2Time)
	compareDates("Final2RevTime", got.Final2RevTime, want.Final2RevTime)

	// Comparar Schedule
	if len(got.Schedule) != len(want.Schedule) {
		t.Errorf("Schedule length mismatch: got %d, want %d", len(got.Schedule), len(want.Schedule))
	} else {
		for day, gd := range got.Schedule {
			wd, ok := want.Schedule[day]
			if !ok {
				t.Errorf("Schedule missing expected day %v", day)
				continue
			}
			if gd.Room != wd.Room {
				t.Errorf("Schedule[%v] Room mismatch: got %q, want %q", day, gd.Room, wd.Room)
			}
			if gd.Time != wd.Time {
				t.Errorf("Schedule[%v] Time mismatch: got %+v, want %+v", day, gd.Time, wd.Time)
			}
		}
	}
}
