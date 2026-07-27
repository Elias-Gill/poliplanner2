package parser

import (
	"os"
	"path"
	"testing"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

func TestParseSubjects_ByCareerSheets(t *testing.T) {
	testPath := path.Join(config.Get().Paths.BaseDir, "test_data", "excel")

	file, err := os.Open(path.Join(testPath, "stripped_excel.xlsx"))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	parser, err := NewParser(file)
	if err != nil {
		t.Fatalf("cannot create parser: %v", err)
	}
	defer parser.Close()

	// Corregido: Nuevo orden secuencial de lectura de las hojas
	expectedOrder := []string{"IIN", "ISP", "LCIK", "CNEL.OVIEDO", "VILLARRICA"}

	expectedSubjectsByCareer := map[string][]SubjectDTO{
		"IIN": {
			subject(
				withGeneral("DCB", 2, 2, "MI", "Algebra Lineal"),
				withTeachers(teacher("Lic.", "Richard Adrián", "Villasanti Flores", "")),
				func(s *SubjectDTO) {
					s.Partial1Date = date(2024, 9, 17)
					s.Partial1Time = hour(8, 0)
					s.Partial1Room = "A50"
					s.Partial2Date = date(2024, 11, 12)
					s.Partial2Time = hour(8, 0)
					s.Partial2Room = "A50"
					s.Final1Date = date(2024, 12, 3)
					s.Final1Time = hour(8, 0)
					s.Final1Room = "A50"
					s.Final1RevDate = date(2024, 12, 10)
					s.Final1RevTime = hour(10, 30)
					s.Final2Date = date(2024, 12, 17)
					s.Final2Time = hour(8, 0)
					s.Final2Room = "A50"
					s.Final2RevDate = date(2024, 12, 26)
					s.Final2RevTime = hour(10, 30)
					s.CommitteePresident = "Lic. Richard Adrián Villasanti Flores"
					s.CommitteeMember1 = "Ms. Osvaldo Ramón Vega Gamarra"
					s.CommitteeMember2 = "Ms. Édgar López Pezoa"
					s.Schedule = [7]WeekDayData{
						academic.Tuesday: {Room: "C01", Time: slot(hour(9, 15), hour(12, 15))},
						academic.Friday:  {Room: "C01", Time: slot(hour(10, 0), hour(12, 15))},
					}
					s.SaturdayDates = "24/03"
				},
			),
			subject(
				withGeneral("DCB", 3, 3, "NB", "Cálculo II"),
				withTeachers(teacher("Lic.", "Silvia Verónica", "Chamorro Hermosa", "schamorro@pol.una.py")),
				func(s *SubjectDTO) {
					s.Partial1Date = date(2024, 9, 9)
					s.Partial1Time = hour(19, 30)
					s.Partial1Room = "A57"
					s.Partial2Date = date(2024, 11, 4)
					s.Partial2Time = hour(19, 30)
					s.Partial2Room = "A59"
					s.Final1Date = date(2024, 11, 25)
					s.Final1Time = hour(19, 30)
					s.Final1Room = "A57"
					s.Final1RevDate = date(2024, 12, 4)
					s.Final1RevTime = hour(19, 0)
					s.Final2Date = date(2024, 12, 23)
					s.Final2Time = hour(19, 30)
					s.Final2Room = "A57"
					s.Final2RevDate = date(2024, 12, 27)
					s.Final2RevTime = hour(19, 0)
					s.CommitteePresident = "Lic. Silvia Verónica Chamorro Hermosa"
					s.CommitteeMember1 = "Ms. Rubén Dario Zárate Rojas"
					s.CommitteeMember2 = "Lic. Pamela Raquel Flores Acosta"
					s.Schedule = [7]WeekDayData{
						academic.Monday:    {Room: "A57", Time: slot(hour(20, 0), hour(22, 15))},
						academic.Wednesday: {Room: "A57", Time: slot(hour(20, 0), hour(22, 15))},
					}
				},
			),
			subject(
				withGeneral("DG", 3, 1, "MJ", "Contabilidad (*)"),
				withTeachers(teacher("C.P.", "Leidy Jessica", "Ríos Argaña", "ljrios@pol.una.py")),
				func(s *SubjectDTO) {
					s.Final1Date = date(2024, 12, 4)
					s.Final1Time = hour(8, 0)
					s.Final1Room = "C03"
					s.Final1RevDate = date(2024, 12, 13)
					s.Final1RevTime = hour(10, 30)
					s.Final2Date = date(2024, 12, 18)
					s.Final2Time = hour(8, 0)
					s.Final2Room = "C03"
					s.Final2RevDate = date(2024, 12, 27)
					s.Final2RevTime = hour(10, 30)
					s.CommitteePresident = "C.P. Leidy Jessica Ríos Argaña"
					s.CommitteeMember1 = "Ms. Cynthia Elizabeth Segovia Domínguez"
					s.CommitteeMember2 = "Lic. Zulma Lucía Demattei Ortiz"
				},
			),
			subject(
				withGeneral("DEI", 10, 0, "TQ", "Diseño de Compiladores"),
				withTeachers(teacher("Ing.", "Sergio Andrés", "Aranda Zemán", "saranda@pol.una.py")),
				func(s *SubjectDTO) {
					s.Partial1Date = date(2024, 9, 13)
					s.Partial1Time = hour(18, 0)
					s.Partial1Room = "F16"
					s.Partial2Date = date(2024, 11, 8)
					s.Partial2Time = hour(18, 0)
					s.Partial2Room = "F16"
					s.Final1Date = date(2024, 12, 11)
					s.Final1Time = hour(18, 0)
					s.Final1Room = "F38"
					s.Final1RevDate = date(2024, 12, 21)
					s.Final1RevTime = hour(10, 30)
					s.Final2Date = date(2024, 12, 28)
					s.Final2Time = hour(8, 0)
					s.Final2Room = "F38"
					s.Final2RevDate = date(2024, 12, 30)
					s.Final2RevTime = hour(17, 0)
					s.CommitteePresident = "Ing. Sergio Andrés Aranda Zemán"
					s.CommitteeMember1 = "Ing. Deysi Natalia Leguizamón Correa"
					s.CommitteeMember2 = "Ing. Fernando Ramón Saucedo Arguello"
					s.Schedule = [7]WeekDayData{
						academic.Monday: {Room: "I06", Time: slot(hour(16, 0), hour(19, 0))},
						academic.Friday: {Room: "F39", Time: slot(hour(18, 30), hour(20, 45))},
					}
				},
			),
		},
		"ISP": {
			subject(
				withGeneral("DG", 0, 2, "TQ", "Administración de Empresas (*)"),
				withTeachers(teacher("Lic.", "Zulma Lucía", "Demattei Ortiz", "zdemattei@pol.una.py")),
				func(s *SubjectDTO) {
					s.Final1Date = date(2024, 11, 27)
					s.Final1Time = hour(18, 0)
					s.Final1Room = "C03"
					s.Final1RevDate = date(2024, 12, 6)
					s.Final1RevTime = hour(17, 0)

					s.Final2Date = date(2024, 12, 13)
					s.Final2Time = hour(18, 0)
					s.Final2Room = "C03"
					s.Final2RevDate = date(2024, 12, 23)
					s.Final2RevTime = hour(17, 0)

					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortiz"
					s.CommitteeMember1 = "Lic. Julio Ramón Riveros Báez"
					s.CommitteeMember2 = "Lic. Osvaldo David Sosa Cabrera"
				},
			),
			subject(
				withGeneral("DCB", 0, 1, "MI", "Álgebra"),
				withTeachers(teacher("Ms.", "Édgar Rubén", "Benítez Penayo", "erbenitez@pol.una.py")),
				func(s *SubjectDTO) {
					// Primer Parcial
					s.Partial1Date = date(2024, 9, 16)
					s.Partial1Time = hour(8, 0)
					s.Partial1Room = "A55"

					// Segundo Parcial
					s.Partial2Date = date(2024, 11, 11)
					s.Partial2Time = hour(8, 0)
					s.Partial2Room = "A55"

					// Primer Final y su Revisión
					s.Final1Date = date(2024, 12, 2)
					s.Final1Time = hour(8, 0)
					s.Final1Room = "A50"
					s.Final1RevDate = date(2024, 12, 11)
					s.Final1RevTime = hour(10, 30)

					// Segundo Final y su Revisión
					s.Final2Date = date(2024, 12, 16)
					s.Final2Time = hour(8, 0)
					s.Final2Room = "A50"
					s.Final2RevDate = date(2024, 12, 23)
					s.Final2RevTime = hour(10, 30)

					// Mesa Examinadora
					s.CommitteePresident = "Ms. Édgar Rubén Benítez Penayo"
					s.CommitteeMember1 = "Ms. Édgar López Pezoa"
					s.CommitteeMember2 = "Lic. María Clara Cáceres Rolón"

					// Horario de Clases (Lunes y Martes de 10:00 a 12:15 en el aula A55)
					s.Schedule = [7]WeekDayData{
						academic.Monday:    {Room: "A55", Time: slot(hour(10, 0), hour(12, 15))},
						academic.Wednesday: {Room: "A55", Time: slot(hour(10, 0), hour(12, 15))},
					}
				},
			),
		},

		"LCIK": {
			subject(
				withGeneral("DG", 1, 1, "MI", "Administración I"),
				withTeachers(
					teacher("Dr.", "Vicente Ramón", "Bracho González", "vrbracho@pol.una.py"),
					teacher("Lic.", "Zulma Lucía", "Demattei Ortiz", "zdemattei@pol.una.py"),
				),
				func(s *SubjectDTO) {
					s.Partial1Date = date(2024, 9, 18)
					s.Partial1Time = hour(8, 0)
					s.Partial1Room = "E01"

					s.Partial2Date = date(2024, 11, 13)
					s.Partial2Time = hour(8, 0)
					s.Partial2Room = "Lab BD\nLab IA\n" // Mantiene saltos de línea del excel si aplica

					s.Final1Date = date(2024, 12, 3)
					s.Final1Time = hour(8, 0)
					s.Final1Room = "Lab AL\nLab HPC"
					s.Final1RevDate = date(2024, 12, 13)
					s.Final1RevTime = hour(10, 30)

					s.Final2Date = date(2024, 12, 17)
					s.Final2Time = hour(8, 0)
					s.Final2Room = "Lab HPC"
					s.Final2RevDate = date(2024, 12, 27)
					s.Final2RevTime = hour(10, 30)

					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortíz"
					s.CommitteeMember1 = "Ms. Cynthia Elizabeth Segovia Domínguez"
					s.CommitteeMember2 = "C.P. Leidy Jessica Ríos Argaña"

					s.Schedule = [7]WeekDayData{
						academic.Wednesday: {Room: "E01", Time: slot(hour(7, 30), hour(9, 0))},
						academic.Friday:    {Room: "E01", Time: slot(hour(7, 30), hour(9, 45))},
					}
				},
			),

			// 12. Administración IV
			subject(
				withGeneral("DG", 5, 4, "NA", "Administración IV"),
				withTeachers(teacher("Ms.", "María Griselda", "Palacios Ferreira", "graciela.cuenca@pol.una.py")),
				func(s *SubjectDTO) {
					s.Partial1Date = date(2024, 9, 18)
					s.Partial1Time = hour(19, 30)
					s.Partial1Room = "C01"

					s.Partial2Date = date(2024, 11, 13)
					s.Partial2Time = hour(19, 30)
					s.Partial2Room = "Lab MS"

					s.Final1Date = date(2024, 12, 4)
					s.Final1Time = hour(19, 30)
					s.Final1Room = "C01"
					s.Final1RevDate = date(2024, 12, 13)
					s.Final1RevTime = hour(19, 0)

					s.Final2Date = date(2024, 12, 18)
					s.Final2Time = hour(19, 30)
					s.Final2Room = "C01"
					s.Final2RevDate = date(2024, 12, 27)
					s.Final2RevTime = hour(19, 0)

					s.CommitteePresident = "Ms. María Griselda Palacios Ferreira"
					s.CommitteeMember1 = "Ms. Alcides Javier Torres Gutt"
					s.CommitteeMember2 = "Lic. Armín Jesús Molas Ovando"

					s.Schedule = [7]WeekDayData{
						academic.Wednesday: {Room: "C01", Time: slot(hour(19, 0), hour(20, 30))},
						academic.Thursday:  {Room: "C01", Time: slot(hour(20, 45), hour(22, 15))},
						academic.Saturday:  {Room: "F11", Time: slot(hour(7, 30), hour(11, 30))},
					}
				},
			),

			// 32. Electiva I - Diseño de Aplicaciones Web y Mobile
			subject(
				withGeneral("DEI", 8, 8, "NA", "Electiva I - Diseño de Aplicaciones Web y Mobile"),
				withTeachers(teacher("Ing.", "Iván Ismael", "Ríos Villalba", "irios@pol.una.py")),
				func(s *SubjectDTO) {
					s.Partial1Date = date(2024, 9, 16)
					s.Partial1Time = hour(19, 30)
					s.Partial1Room = "Lab MS"

					s.Partial2Date = date(2024, 11, 11)
					s.Partial2Time = hour(19, 30)
					s.Partial2Room = "Lab MS"

					s.Final1Date = date(2024, 12, 2)
					s.Final1Time = hour(19, 30)
					s.Final1Room = "Lab AL"
					s.Final1RevDate = date(2024, 12, 12)
					s.Final1RevTime = hour(19, 0)

					s.Final2Date = date(2024, 12, 16)
					s.Final2Time = hour(19, 30)
					s.Final2Room = "Lab AL"
					s.Final2RevDate = date(2024, 12, 26)
					s.Final2RevTime = hour(19, 0)

					s.CommitteePresident = "Ing. Iván Ismael Ríos Villalba"
					s.CommitteeMember1 = "Lic. José Rodrigo Benitez Paredes"
					s.CommitteeMember2 = "Lic. Carlos David Riveros Giménez"

					s.Schedule = [7]WeekDayData{
						academic.Monday:   {Room: "Lab MS", Time: slot(hour(20, 45), hour(22, 15))},
						academic.Friday:   {Room: "Lab MS", Time: slot(hour(19, 0), hour(20, 30))},
						academic.Saturday: {Room: "F13", Time: slot(hour(7, 30), hour(11, 30))},
					}
				},
			),

			// 33. Electiva I - Gestión de Personas (*)
			subject(
				withGeneral("DG", 8, 8, "TQ", "Electiva I - Gestión de Personas (*)"),
				withTeachers(teacher("Dr.", "Vicente Ramón", "Bracho González", "vrbracho@pol.una.py")),
				func(s *SubjectDTO) {
					// No tiene fechas de parciales especificadas
					s.Final1Date = date(2024, 12, 2)
					s.Final1Time = hour(15, 0)
					s.Final1Room = "F13"
					s.Final1RevDate = date(2024, 12, 12)
					s.Final1RevTime = hour(14, 0)

					s.Final2Date = date(2024, 12, 16)
					s.Final2Time = hour(15, 0)
					s.Final2Room = "F15"
					s.Final2RevDate = date(2024, 12, 26)
					s.Final2RevTime = hour(14, 0)

					s.CommitteePresident = "Lic. Zulma Lucía Demattei Ortíz"
					s.CommitteeMember1 = "Ms. Julio Néstor Sánchez Laspina"
					s.CommitteeMember2 = "Econ. Jerson Fernández Caje"

					// No se especifican horarios semanales en las columnas de días para esta materia
					s.Schedule = [7]WeekDayData{}
				},
			),
		},

		"CNEL.OVIEDO": {
			// 38. Programación de Aplicaciones en Redes
			subject(
				withGeneral("DEI", 7, 7, "U", "Programación de Aplicaciones en Redes"),
				withTeachers(teacher("Lic.", "María Luisa", "Guanes Romero", "mguanes@pol.una.py")),
				func(s *SubjectDTO) {
					// Parciales
					s.Partial1Date = date(2024, 9, 27)
					s.Partial1Time = hour(13, 30)
					s.Partial1Room = "" // No especifica aula en el extracto

					s.Partial2Date = date(2024, 11, 15)
					s.Partial2Time = hour(13, 30)
					s.Partial2Room = ""

					// Primer Final y su Revisión
					s.Final1Date = date(2024, 12, 6)
					s.Final1Time = hour(13, 30)
					s.Final1Room = ""
					s.Final1RevDate = date(2024, 12, 13)
					s.Final1RevTime = hour(13, 30)

					// Segundo Final y su Revisión
					s.Final2Date = date(2024, 12, 20)
					s.Final2Time = hour(13, 30)
					s.Final2Room = ""
					s.Final2RevDate = date(2024, 12, 27)
					s.Final2RevTime = hour(13, 30)

					// Mesa Examinadora
					s.CommitteePresident = "Lic. María Luisa Guanes Romero"
					s.CommitteeMember1 = "Lic. Flaminio Aranda Ibáñez"
					s.CommitteeMember2 = "Lic. Rodney Alberto Colmán Alvarenga"

					// Horario de Clases (Viernes de 13:30 a 17:30)
					s.Schedule = [7]WeekDayData{
						academic.Friday: {Room: "", Time: slot(hour(13, 30), hour(17, 30))},
					}
				},
			),

			// 39. Proyecto I
			subject(
				withGeneral("DEI", 7, 7, "U", "Proyecto I"),
				withTeachers(teacher("Lic.", "María Luisa", "Guanes Romero", "mguanes@pol.una.py")),
				func(s *SubjectDTO) {
					// Parciales
					s.Partial1Date = date(2024, 9, 13)
					s.Partial1Time = hour(17, 30)
					s.Partial1Room = ""

					s.Partial2Date = date(2024, 11, 8)
					s.Partial2Time = hour(17, 30)
					s.Partial2Room = ""

					// Primer Final y su Revisión
					s.Final1Date = date(2024, 12, 11)
					s.Final1Time = hour(17, 30)
					s.Final1Room = ""
					s.Final1RevDate = date(2024, 12, 18)
					s.Final1RevTime = hour(17, 30)

					// Segundo Final y su Revisión
					s.Final2Date = date(2024, 12, 27)
					s.Final2Time = hour(17, 30)
					s.Final2Room = ""
					s.Final2RevDate = date(2024, 12, 30)
					s.Final2RevTime = hour(17, 30)

					// Mesa Examinadora
					s.CommitteePresident = "Lic. María Luisa Guanes Romero"
					s.CommitteeMember1 = "Lic. Rodney Alberto Colmán Alvarenga"
					s.CommitteeMember2 = "Ing. Ignacio Daniel Velázquez Guachiré"

					// Horario de Clases (Viernes de 17:30 a 21:30)
					s.Schedule = [7]WeekDayData{
						academic.Friday: {Room: "", Time: slot(hour(17, 30), hour(21, 30))},
					}
				},
			),
		},

		"VILLARRICA": {
			subject(
				withGeneral("DG", 8, 0, "U", "Alimentos y Bebidas VI"),
				withTeachers(teacher("Lic.", "Milka Paola", "Velázquez Romero", "milka@pol.una.py")),
				func(s *SubjectDTO) {
					s.CommitteePresident = "Lic. Milka Paola Velázquez Romero"
					s.Schedule = [7]WeekDayData{}
				},
			),
			subject(
				withGeneral("DEI", 6, 6, "U", "Compiladores y Lenguajes de Bajo Nivel (*)"),
				withTeachers(teacher("Lic.", "Flaminio", "Aranda Ibáñez", "faranda@pol.una.py")),
				func(s *SubjectDTO) {
					s.CommitteePresident = "Lic. Flaminio Aranda Ibáñez"
					s.Schedule = [7]WeekDayData{}
				},
			),
			subject(
				withGeneral("DEE", 8, 0, "U", "Electricidad de Potencia (*)"),
				withTeachers(teacher("Ing.", "Rubén Darío", "Vera", "rubendvera@pol.una.py")),
				func(s *SubjectDTO) {
					s.CommitteePresident = "Ing. Rubén Darío Vera"
					s.Schedule = [7]WeekDayData{}
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
			t.Fatalf("parse error on sheet index %d: %v", sheetIndex, err)
		}

		expectedName := expectedOrder[sheetIndex]
		if parsed.Name != expectedName {
			t.Fatalf("sheet order mismatch: got %q want %q", parsed.Name, expectedName)
		}

		expectedSubjects := expectedSubjectsByCareer[expectedName]
		if len(parsed.Subjects) != len(expectedSubjects) {
			t.Fatalf("[%s] subjects count mismatch: got %d want %d", expectedName, len(parsed.Subjects), len(expectedSubjects))
		}

		for i := range expectedSubjects {
			assertSubjectEqual(t, expectedName, parsed.Subjects[i], expectedSubjects[i])
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

func slot(start, end Hour) TimeSlot {
	return TimeSlot{Start: start, End: end}
}

func date(year, month, day int) Date {
	return Date{Year: year, Month: month, Day: day, Valid: true}
}

func hour(hour, minute int) Hour {
	return Hour{Hour: hour, Minute: minute, Valid: true}
}

func subject(opts ...func(*SubjectDTO)) SubjectDTO {
	s := SubjectDTO{}
	for _, fn := range opts {
		fn(&s)
	}
	return s
}

func withGeneral(dept string, semester int, level int, section, raw string) func(*SubjectDTO) {
	return func(s *SubjectDTO) {
		s.Department = dept
		s.Semester = semester
		s.Level = level
		s.Section = section
		s.RawSubjectName = raw
	}
}

func withTeachers(t ...TeacherDTO) func(*SubjectDTO) {
	return func(s *SubjectDTO) {
		for i, teacher := range t {
			if i >= 4 {
				break
			}
			s.Teachers[i] = teacher
		}
		s.TeacherCount = len(t)
	}
}

func assertSubjectEqual(t *testing.T, sheet string, got, want SubjectDTO) {
	ctx := "[" + sheet + " -> " + want.RawSubjectName + "]"

	if got.Department != want.Department {
		t.Errorf("%s Department mismatch: got %q, want %q", ctx, got.Department, want.Department)
	}
	if got.Semester != want.Semester {
		t.Errorf("%s Semester mismatch: got %d, want %d", ctx, got.Semester, want.Semester)
	}
	if got.Section != want.Section {
		t.Errorf("%s Section mismatch: got %q, want %q", ctx, got.Section, want.Section)
	}
	if got.RawSubjectName != want.RawSubjectName {
		t.Errorf("%s RawSubjectName mismatch: got %q, want %q", ctx, got.RawSubjectName, want.RawSubjectName)
	}

	if got.TeacherCount != want.TeacherCount {
		t.Errorf("%s TeacherCount mismatch: got %d, want %d", ctx, got.TeacherCount, want.TeacherCount)
	} else {
		for i := 0; i < want.TeacherCount; i++ {
			if got.Teachers[i] != want.Teachers[i] {
				t.Errorf("%s Teacher[%d] mismatch: got %+v, want %+v", ctx, i, got.Teachers[i], want.Teachers[i])
			}
		}
	}

	compareHours := func(field string, gotHour, wantHour Hour) {
		if gotHour.Valid != wantHour.Valid {
			t.Errorf("%s %s Valid mismatch: got %t, want %t", ctx, field, gotHour.Valid, wantHour.Valid)
			return
		}
		if gotHour.Valid && gotHour != wantHour {
			t.Errorf("%s %s mismatch: got %+v, want %+v", ctx, field, gotHour, wantHour)
		}
	}

	compareDates := func(field string, gotDate, wantDate Date) {
		if gotDate.Valid != wantDate.Valid {
			t.Errorf("%s %s Valid mismatch: got %t, want %t", ctx, field, gotDate.Valid, wantDate.Valid)
			return
		}
		if gotDate.Valid && gotDate != wantDate {
			t.Errorf("%s %s mismatch: got %+v, want %+v", ctx, field, gotDate, wantDate)
		}
	}

	compareDates("Partial1Date", got.Partial1Date, want.Partial1Date)
	compareHours("Partial1Time", got.Partial1Time, want.Partial1Time)
	compareDates("Partial2Date", got.Partial2Date, want.Partial2Date)
	compareHours("Partial2Time", got.Partial2Time, want.Partial2Time)
	compareDates("Final1Date", got.Final1Date, want.Final1Date)
	compareHours("Final1Time", got.Final1Time, want.Final1Time)
	compareDates("Final1RevDate", got.Final1RevDate, want.Final1RevDate)
	compareHours("Final1RevTime", got.Final1RevTime, want.Final1RevTime)
	compareDates("Final2Date", got.Final2Date, want.Final2Date)
	compareHours("Final2Time", got.Final2Time, want.Final2Time)
	compareDates("Final2RevDate", got.Final2RevDate, want.Final2RevDate)
	compareHours("Final2RevTime", got.Final2RevTime, want.Final2RevTime)

	if got.Partial1Room != want.Partial1Room {
		t.Errorf("%s Partial1Room mismatch: got %q, want %q", ctx, got.Partial1Room, want.Partial1Room)
	}
	if got.Partial2Room != want.Partial2Room {
		t.Errorf("%s Partial2Room mismatch: got %q, want %q", ctx, got.Partial2Room, want.Partial2Room)
	}
	if got.Final1Room != want.Final1Room {
		t.Errorf("%s Final1Room mismatch: got %q, want %q", ctx, got.Final1Room, want.Final1Room)
	}
	if got.Final2Room != want.Final2Room {
		t.Errorf("%s Final2Room mismatch: got %q, want %q", ctx, got.Final2Room, want.Final2Room)
	}

	// Mapeo de índices a días legibles de la semana
	dayNames := [7]string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"}

	for day := range 7 {
		gd := got.Schedule[day]
		wd := want.Schedule[day]
		dayLabel := dayNames[day]

		if gd.Room != wd.Room {
			t.Errorf("%s Schedule[%s] Room mismatch: got %q, want %q", ctx, dayLabel, gd.Room, wd.Room)
		}
		if gd.Time.Start != wd.Time.Start || gd.Time.End != wd.Time.End {
			t.Errorf("%s Schedule[%s] Time mismatch: got %+v, want %+v", ctx, dayLabel, gd.Time, wd.Time)
		}
	}
}
