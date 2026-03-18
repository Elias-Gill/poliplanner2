package parser

import (
	"os"
	"testing"
	"time"
)

func TestParseSubjects_ByCareerSheets(t *testing.T) {
	file, err := os.Open("../../test_data/excel/stripped_excel.xlsx")
	if err != nil {
		panic(err)
	}

	parser, err := NewExcelParser(file)
	if err != nil {
		t.Fatalf("cannot create parser: %v", err)
	}
	defer parser.Close()

	expectedOrder := []string{
		"IIN",
		"CNEL.OVIEDO",
		"VILLARRICA",
		"LCIK",
		"ISP",
	}

	// expectedSubjectsByCareer mantiene el orden por sheet
	expectedSubjectsByCareer := map[string][]SubjectDTO{
		// FIX: revisar test cases porque estan mal con el excel que uso de pruebas
		"IIN": {
			// Item 1: Algebra Lineal
			subject(
				withGeneral("DCB", 2, "MI", "IIN", "Algebra Lineal"),
				withTeachers(teacher("Lic.", "Villasanti Flores", "Richard Adrián", "")),
				func(s *SubjectDTO) {
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 17)
					s.Partial1Time = "08:00"
					s.Partial1Room = "A50"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 12)
					s.Partial2Time = "08:00"
					s.Partial2Room = "A50"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 3)
					s.Final1Time = "08:00"
					s.Final1Room = "A50"
					s.Final1RevDate = datePtr(2024, time.December, 10)
					s.Final1RevTime = "10:30:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 17)
					s.Final2Time = "08:00"
					s.Final2Room = "A50"
					s.Final2RevDate = datePtr(2024, time.December, 26)
					s.Final2RevTime = "10:30:00"
					// Comité
					s.CommitteePresident = "Lic. Richard Adrián Villasantti Flores"
					s.CommitteeMember1 = "Ms. Osvaldo Ramón Vega Gamarra"
					s.CommitteeMember2 = "Ms. Édgar López Pezoa"
					// Horarios semanales
					s.WednesdayRoom = "C01"
					s.Wednesday = slot("09:15", "12:15")
					s.FridayRoom = "C01"
					s.Friday = slot("10:00", "12:15")
					// Sábados
					s.SaturdayDates = "24/03"
				},
			),
			// Item 22: Cálculo II
			subject(
				withGeneral("DCB", 3, "NB", "IIN", "Cálculo II"),
				withTeachers(teacher("Lic.", "Chamorro Hermosa", "Silvia Verónica", "schamorro@pol.una.py")),
				func(s *SubjectDTO) {
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 9)
					s.Partial1Time = "19:30:00"
					s.Partial1Room = "A57"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 4)
					s.Partial2Time = "19:30:00"
					s.Partial2Room = "A59"
					// 1er Final
					s.Final1Date = datePtr(2024, time.November, 25)
					s.Final1Time = "19:30:00"
					s.Final1Room = "A57"
					s.Final1RevDate = datePtr(2024, time.December, 4)
					s.Final1RevTime = "19:00:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 23)
					s.Final2Time = "19:30:00"
					s.Final2Room = "A57"
					s.Final2RevDate = datePtr(2024, time.December, 27)
					s.Final2RevTime = "19:00:00"
					// Comité
					s.CommitteePresident = "Lic. Silvia Verónica Chamorro Hermosa"
					s.CommitteeMember1 = "Ms. Rubén Dario Zárate Rojas"
					s.CommitteeMember2 = "Lic. Pamela Raquel Flores Acosta"
					// Horarios semanales
					s.MondayRoom = "A57"
					s.Monday = slot("20:00", "22:15")
					s.WednesdayRoom = "A57"
					s.Wednesday = slot("20:00", "22:15")
				},
			),
			// Item 30: Contabilidad (*)
			subject(
				withGeneral("DG", 3, "MJ", "IIN", "Contabilidad (*)"),
				withTeachers(teacher("C.P.", "Ríos Argaña", "Leidy Jessica", "ljrios@pol.una.py")),
				func(s *SubjectDTO) {
					// 2do Parcial (no tiene 1er parcial)
					s.Partial2Date = datePtr(2024, time.December, 4)
					s.Partial2Time = "08:00"
					s.Partial2Room = "C03"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 18)
					s.Final1Time = "08:00"
					s.Final1Room = "C03"
					s.Final1RevDate = datePtr(2024, time.December, 13)
					s.Final1RevTime = "10:30:00"
					// 2do Final
					s.Final2RevDate = datePtr(2024, time.December, 27)
					s.Final2RevTime = "10:30:00"
					// Comité
					s.CommitteePresident = "C.P. Leidy Jessica Ríos Argaña"
					s.CommitteeMember1 = "Ms. Cynthia Elizabeth Segovia Domínguez"
					s.CommitteeMember2 = "Lic. Zulma Lucía Demattei Ortiz"
				},
			),
			// Item 31: Diseño de Compiladores
			subject(
				withGeneral("DEI", 10, "TQ", "IIN", "Diseño de Compiladores"),
				withTeachers(teacher("Ing.", "Aranda Zemán", "Sergio Andrés", "saranda@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 10
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 13)
					s.Partial1Time = "18:00:00"
					s.Partial1Room = "F16"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 8)
					s.Partial2Time = "18:00:00"
					s.Partial2Room = "F16"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 11)
					s.Final1Time = "18:00:00"
					s.Final1Room = "F38"
					s.Final1RevDate = datePtr(2024, time.December, 21)
					s.Final1RevTime = "10:30:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 28)
					s.Final2Time = "08:00:00"
					s.Final2Room = "F38"
					s.Final2RevDate = datePtr(2024, time.December, 30)
					s.Final2RevTime = "17:00:00"
					// Comité
					s.CommitteePresident = "Ing. Sergio Andrés Aranda Zemán"
					s.CommitteeMember1 = "Ing. Deysi Natalia Leguizamón Correa"
					s.CommitteeMember2 = "Ing. Fernando Ramón Saucedo Arguello"
					// Horarios semanales
					s.MondayRoom = "I06"
					s.Monday = slot("16:00", "19:00")
					s.FridayRoom = "F39"
					s.Friday = slot("18:30", "20:45")
				},
			),
		},
		"ISP": {
			// Item 1: Administración de Empresas (*)
			subject(
				withGeneral("DG", 2, "TQ", "ISP", "Administración de Empresas (*)"),
				withTeachers(teacher("Lic.", "Demattei Ortiz", "Zulma Lucía", "zdemattei@pol.una.py")),
				func(s *SubjectDTO) {
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 13)
					s.Final1Time = "18:00:00"
					s.Final1Room = "C03"
					s.Final1RevDate = datePtr(2024, time.November, 27)
					s.Final1RevTime = "18:00:00"
					// 2do Final
					s.Final2RevDate = datePtr(2024, time.December, 6)
					s.Final2RevTime = "17:00:00"
					s.Final2Date = datePtr(2024, time.December, 23)
					s.Final2Time = "17:00:00"
					// Comité
					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortiz"
					s.CommitteeMember1 = "Lic. Julio Ramón Riveros Báez"
					s.CommitteeMember2 = "Lic. Osvaldo David Sosa Cabrera"
				},
			),
			// Item 2: Álgebra
			subject(
				withGeneral("DCB", 1, "MI", "ISP", "Álgebra"),
				withTeachers(teacher("Ms.", "Benítez Penayo", "Édgar Rubén", "erbenitez@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 1
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 16)
					s.Partial1Time = "08:00"
					s.Partial1Room = "A55"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 11)
					s.Partial2Time = "08:00"
					s.Partial2Room = "A55"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 2)
					s.Final1Time = "08:00"
					s.Final1Room = "A50"
					s.Final1RevDate = datePtr(2024, time.December, 11)
					s.Final1RevTime = "10:30:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 16)
					s.Final2Time = "08:00"
					s.Final2Room = "A50"
					s.Final2RevDate = datePtr(2024, time.December, 23)
					s.Final2RevTime = "10:30:00"
					// Comité
					s.CommitteePresident = "Ms. Édgar Rubén Benítez Penayo"
					s.CommitteeMember1 = "Ms. Édgar López Pezoa"
					s.CommitteeMember2 = "Lic. María Clara Cáceres Rolón"
					// Horarios semanales
					s.MondayRoom = "A55"
					s.Monday = slot("10:00", "12:15")
					s.WednesdayRoom = "A55"
					s.Wednesday = slot("10:00", "12:15")
				},
			),
		},
		"CNEL.OVIEDO": {
			// Item 38: Programación de Aplicaciones en Redes
			subject(
				withGeneral("DEI", 7, "U", "LCIk", "Programación de Aplicaciones en Redes"),
				withTeachers(teacher("Lic.", "Guanes Romero", "María Luisa", "mguanes@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 7
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 27)
					s.Partial1Time = "13:30:00"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 15)
					s.Partial2Time = "13:30:00"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 6)
					s.Final1Time = "13:30:00"
					s.Final1RevDate = datePtr(2024, time.December, 13)
					s.Final1RevTime = "13:30:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 20)
					s.Final2Time = "13:30:00"
					s.Final2RevDate = datePtr(2024, time.December, 27)
					s.Final2RevTime = "13:30:00"
					// Comité
					s.CommitteePresident = "Lic. María Luisa Guanes Romero"
					s.CommitteeMember1 = "Lic. Flaminio Aranda Ibáñez"
					s.CommitteeMember2 = "Lic. Rodney Alberto Colmán Alvarenga"
					// Horarios semanales
					s.Friday = slot("13:30", "17:30")
				},
			),
			// Item 39: Proyecto I
			subject(
				withGeneral("DEI", 7, "U", "LCIk", "Proyecto I"),
				withTeachers(teacher("Lic.", "Guanes Romero", "María Luisa", "mguanes@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 7
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 13)
					s.Partial1Time = "17:30:00"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 8)
					s.Partial2Time = "17:30:00"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 11)
					s.Final1Time = "17:30:00"
					s.Final1RevDate = datePtr(2024, time.December, 18)
					s.Final1RevTime = "17:30:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 27)
					s.Final2Time = "17:30:00"
					s.Final2RevDate = datePtr(2024, time.December, 30)
					s.Final2RevTime = "17:30:00"
					// Comité
					s.CommitteePresident = "Lic. María Luisa Guanes Romero"
					s.CommitteeMember1 = "Lic. Rodney Alberto Colmán Alvarenga"
					s.CommitteeMember2 = "Ing. Ignacio Daniel Velázquez Guachiré"
					// Horarios semanales
					s.Friday = slot("17:30", "21:30")
				},
			),
		},
		"VILLARRICA": {
			// Item 14: Alimentos y Bebidas VI
			subject(
				withGeneral("DG", 8, "U", "LGH", "Alimentos y Bebidas VI"),
				withTeachers(teacher("Lic.", "Velázquez Romero", "Milka Paola", "milka@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 8
					s.CommitteePresident = "Lic. Milka Paola Velázquez Romero"
					// Sin fechas de exámenes
				},
			),
			// Item 27: Compiladores y Lenguajes de Bajo Nivel (*)
			subject(
				withGeneral("DEI", 6, "U", "LCIk", "Compiladores y Lenguajes de Bajo Nivel (*)"),
				withTeachers(teacher("Lic.", "Aranda Ibáñez", "Flaminio", "faranda@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 6
					s.CommitteePresident = "Lic. Flaminio Aranda Ibáñez"
					// Sin fechas de exámenes
				},
			),
			// Item 47: Electricidad de Potencia (*)
			subject(
				withGeneral("DEE", 8, "N", "LEL", "Electricidad de Potencia (*)"),
				withTeachers(teacher("Ing.", "Vera", "Rubén Darío", "rubendvera@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 8
					s.CommitteePresident = "Ing. Rubén Darío Vera"
					// Sin fechas de exámenes
				},
			),
		},
		"LCIK": {
			// Item 1: Administración I (con dos docentes)
			subject(
				withGeneral("DG", 1, "MI", "LCIk", "Administración I"),
				withTeachers(
					teacher("Dr.", "Bracho González", "Vicente Ramón", "vrbracho@pol.una.py"),
					teacher("Lic.", "Demattei Ortiz", "Zulma Lucía", "zdemattei@pol.una.py"),
				),
				func(s *SubjectDTO) {
					s.Level = 1
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 18)
					s.Partial1Time = "08:00"
					s.Partial1Room = "E01"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 13)
					s.Partial2Time = "08:00"
					s.Partial2Room = "Lab BD\nLab IA\n"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 3)
					s.Final1Time = "08:00"
					s.Final1Room = "Lab AL\nLab HPC"
					s.Final1RevDate = datePtr(2024, time.December, 13)
					s.Final1RevTime = "10:30:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 17)
					s.Final2Time = "08:00"
					s.Final2Room = "Lab HPC"
					s.Final2RevDate = datePtr(2024, time.December, 27)
					s.Final2RevTime = "10:30:00"
					// Comité
					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortíz"
					s.CommitteeMember1 = "Ms. Cynthia Elizabeth Segovia Domínguez"
					s.CommitteeMember2 = "C.P. Leidy Jessica Ríos Argaña"
					// Horarios semanales
					s.WednesdayRoom = "E01"
					s.Wednesday = slot("07:30", "09:00")
					s.FridayRoom = "E01"
					s.Friday = slot("07:30", "09:45")
				},
			),
			// Item 12: Administración IV
			subject(
				withGeneral("DG", 5, "NA", "LCIk", "Administración IV"),
				withTeachers(teacher("Ms.", "Palacios Ferreira", "María Griselda", "graciela.cuenca@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 4
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 18)
					s.Partial1Time = "19:30:00"
					s.Partial1Room = "C01"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 13)
					s.Partial2Time = "19:30:00"
					s.Partial2Room = "Lab MS"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 4)
					s.Final1Time = "19:30:00"
					s.Final1Room = "C01"
					s.Final1RevDate = datePtr(2024, time.December, 13)
					s.Final1RevTime = "19:00:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 18)
					s.Final2Time = "19:30:00"
					s.Final2Room = "C01"
					s.Final2RevDate = datePtr(2024, time.December, 27)
					s.Final2RevTime = "19:00:00"
					// Comité
					s.CommitteePresident = "Ms. María Griselda Palacios Ferreira"
					s.CommitteeMember1 = "Ms. Alcides Javier Torres Gutt"
					s.CommitteeMember2 = "Lic. Armín Jesús Molas Ovando"
					// Horarios semanales
					s.MondayRoom = "C01"
					s.Monday = slot("19:00", "20:30")
					s.TuesdayRoom = "C01"
					s.Tuesday = slot("20:45", "22:15")
					s.FridayRoom = "F11"
					s.Friday = slot("07:30", "11:30")
					s.SaturdayDates = "05/10, 23/11"
				},
			),
			// Item 32: Electiva I - Diseño de Aplicaciones Web y Mobile
			subject(
				withGeneral("DEI", 8, "NA", "LCIk", "Electiva I - Diseño de Aplicaciones Web y Mobile"),
				withTeachers(teacher("Ing.", "Ríos Villalba", "Iván Ismael", "irios@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 8
					// 1er Parcial
					s.Partial1Date = datePtr(2024, time.September, 16)
					s.Partial1Time = "19:30:00"
					s.Partial1Room = "Lab MS"
					// 2do Parcial
					s.Partial2Date = datePtr(2024, time.November, 11)
					s.Partial2Time = "19:30:00"
					s.Partial2Room = "Lab MS"
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 2)
					s.Final1Time = "19:30:00"
					s.Final1Room = "Lab AL"
					s.Final1RevDate = datePtr(2024, time.December, 12)
					s.Final1RevTime = "19:00:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 16)
					s.Final2Time = "19:30:00"
					s.Final2Room = "Lab AL"
					s.Final2RevDate = datePtr(2024, time.December, 26)
					s.Final2RevTime = "19:00:00"
					// Comité
					s.CommitteePresident = "Ing. Iván Ismael Ríos Villalba"
					s.CommitteeMember1 = "Lic. José Rodrigo Benitez Paredes"
					s.CommitteeMember2 = "Lic. Carlos David Riveros Giménez"
					// Horarios semanales
					s.MondayRoom = "Lab MS"
					s.Monday = slot("20:45", "22:15")
					s.FridayRoom = "Lab MS"
					s.Friday = slot("19:00", "20:30")
					s.SaturdayRoom = "F13"
					s.Saturday = slot("07:30", "11:30")
					s.SaturdayDates = "28/09, 02/11"
				},
			),
			// Item 33: Electiva I - Gestión de Personas (*)
			subject(
				withGeneral("DG", 8, "TQ", "LCIk", "Electiva I - Gestión de Personas (*)"),
				withTeachers(teacher("Dr.", "Bracho González", "Vicente Ramón", "vrbracho@pol.una.py")),
				func(s *SubjectDTO) {
					s.Level = 8
					// 1er Final
					s.Final1Date = datePtr(2024, time.December, 2)
					s.Final1Time = "15:00:00"
					s.Final1Room = "F13"
					s.Final1RevDate = datePtr(2024, time.December, 12)
					s.Final1RevTime = "14:00"
					// 2do Final
					s.Final2Date = datePtr(2024, time.December, 16)
					s.Final2Time = "15:00:00"
					s.Final2Room = "F15"
					s.Final2RevDate = datePtr(2024, time.December, 26)
					s.Final2RevTime = "14:00:00"
					// Comité
					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortíz"
					s.CommitteeMember1 = "Ms. Julio Néstor Sánchez Laspina"
					s.CommitteeMember2 = "Econ. Jerson Fernández Caje"
				},
			),
		},
	}

	var sheetIndex int

	for parser.NextSheet() {
		if sheetIndex >= len(expectedOrder) {
			t.Fatalf("se encontraron más sheets de lo esperado")
		}

		parsed, err := parser.ParseCurrentSheet()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		expectedName := expectedOrder[sheetIndex]

		if parsed.Name != expectedName {
			t.Fatalf(
				"sheet fuera de orden: got %q want %q (index %d)",
				parsed.Name,
				expectedName,
				sheetIndex,
			)
		}

		expectedSubjects := expectedSubjectsByCareer[expectedName]

		if len(parsed.Subjects) != len(expectedSubjects) {
			t.Fatalf(
				"cantidad incorrecta en %s: got %d want %d",
				expectedName,
				len(parsed.Subjects),
				len(expectedSubjects),
			)
		}

		for i := range expectedSubjects {
			assertSubjectEqual(
				t,
				parsed.Subjects[i],
				expectedSubjects[i],
			)
		}

		sheetIndex++
	}

	if sheetIndex != len(expectedOrder) {
		t.Fatalf(
			"faltaron sheets: got %d want %d",
			sheetIndex,
			len(expectedOrder),
		)
	}
}

func teacher(title, first, last, email string) TeacherDTO {
	return TeacherDTO{
		Title:     title,
		FirstName: first,
		LastName:  last,
		Email:     email,
	}
}

func slot(start, end string) TimeSlot {
	return TimeSlot{Start: start, End: end}
}

func datePtr(year int, month time.Month, day int) *time.Time {
	d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &d
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
		s.SubjectName = raw
	}
}

func withTeachers(t ...TeacherDTO) func(*SubjectDTO) {
	return func(s *SubjectDTO) {
		s.Teachers = t
	}
}

func withCommittee(p, m1, m2 string) func(*SubjectDTO) {
	return func(s *SubjectDTO) {
		s.CommitteePresident = p
		s.CommitteeMember1 = m1
		s.CommitteeMember2 = m2
	}
}

func assertSubjectEqual(t *testing.T, got, want SubjectDTO) {
	t.Helper()

	if got.Department != want.Department {
		t.Fatalf("Department: got %q want %q", got.Department, want.Department)
	}
	if got.Semester != want.Semester {
		t.Fatalf("Semester: got %d want %d", got.Semester, want.Semester)
	}
	if got.Section != want.Section {
		t.Fatalf("Section: got %q want %q", got.Section, want.Section)
	}
	if got.SubjectName != want.SubjectName {
		t.Fatalf("RawSubjectName: got %q want %q", got.SubjectName, want.SubjectName)
	}

	if len(got.Teachers) != len(want.Teachers) {
		t.Fatalf("Teachers len: got %d want %d", len(got.Teachers), len(want.Teachers))
	}

	for i := range got.Teachers {
		g := got.Teachers[i]
		w := want.Teachers[i]

		if g != w {
			t.Fatalf("Teacher[%d]: got %+v want %+v", i, g, w)
		}
	}

	if !datesEqual(got.Partial1Date, want.Partial1Date) {
		t.Fatalf("Partial1Date mismatch")
	}
	if got.Partial1Time != want.Partial1Time {
		t.Fatalf("Partial1Time mismatch")
	}
	if got.Partial1Room != want.Partial1Room {
		t.Fatalf("Partial1Room mismatch")
	}

	if !datesEqual(got.Final1Date, want.Final1Date) {
		t.Fatalf("Final1Date mismatch")
	}
	if got.Final1Time != want.Final1Time {
		t.Fatalf("Final1Time mismatch")
	}
	if got.Final1Room != want.Final1Room {
		t.Fatalf("Final1Room mismatch")
	}

	assertSlot(t, got.Monday, want.Monday)
	assertSlot(t, got.Tuesday, want.Tuesday)
	assertSlot(t, got.Wednesday, want.Wednesday)
	assertSlot(t, got.Thursday, want.Thursday)
	assertSlot(t, got.Friday, want.Friday)
	assertSlot(t, got.Saturday, want.Saturday)

	if got.SaturdayDates != want.SaturdayDates {
		t.Fatalf("SaturdayDates mismatch")
	}

	if got.CommitteePresident != want.CommitteePresident {
		t.Fatalf("CommitteePresident mismatch")
	}
	if got.CommitteeMember1 != want.CommitteeMember1 {
		t.Fatalf("CommitteeMember1 mismatch")
	}
	if got.CommitteeMember2 != want.CommitteeMember2 {
		t.Fatalf("CommitteeMember2 mismatch")
	}
}

func assertSlot(t *testing.T, got, want TimeSlot) {
	t.Helper()
	if got != want {
		t.Fatalf("TimeSlot: got %+v want %+v", got, want)
	}
}

func datesEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
