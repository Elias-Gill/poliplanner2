package excelimport

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/metadata"
)

// ============================================================
// Entry point
// ============================================================

func buildOfferingFromDTO(
	career string,
	periodID period.PeriodID,
	data parser.SubjectDTO,
	planLoader *metadata.AcademicPlanLoader,
) Offering {
	enrichDTO(&data, planLoader)

	offer := Offering{
		CourseName: data.RawSubjectName,
		Period:     periodID,
		Teachers:   buildTeachers(data.Teachers),
		Subject:    buildSubject(career, data),
		Section:    data.Section,
		CourseType: mapCourseType(data.CourseType),

		Exams: buildExams(data),

		CommitteePresident: data.CommitteePresident,
		CommitteeMember1:   data.CommitteeMember1,
		CommitteeMember2:   data.CommitteeMember2,
	}

	applySchedule(&offer, data.Schedule)

	return offer
}

// ============================================================
// Enrichment (with known metadata)
// ============================================================

func enrichDTO(data *parser.SubjectDTO, loader *metadata.AcademicPlanLoader) {
	if loader == nil {
		return
	}

	if data.Semester != 0 {
		return
	}

	if m, err := loader.FindSubject(data.RawSubjectName); err == nil {
		data.Semester = m.Semester
	}
}

// ============================================================
// Builders
// ============================================================

func buildSubject(career string, data parser.SubjectDTO) subject {
	return subject{
		Career:     strings.ToUpper(strings.TrimSpace(career)),
		Name:       normalizeSubjectName(data.RawSubjectName),
		Department: data.Department,
		Level:      data.Level,
		Semester:   data.Semester,
	}
}

func buildTeachers(src []parser.TeacherDTO) []teacher {
	teachers := make([]teacher, 0, len(src))

	for _, t := range src {
		teachers = append(teachers, newTeacher(
			t.FirstName,
			t.LastName,
			t.Email,
		))
	}

	return teachers
}

func mapCourseType(t parser.CourseType) courseOffering.CourseType {
	if t == parser.ExamOnlyCourse {
		return courseOffering.ExamOnly
	}
	return courseOffering.Normal
}

// ============================================================
// Exams
// ============================================================

type examInput struct {
	t       courseOffering.ExamType
	i       courseOffering.ExamInstance
	date    *parser.Date
	hour    *parser.Hour
	revDate *parser.Date
	revHour *parser.Hour
	room    string
}

func buildExams(data parser.SubjectDTO) []Exam {
	inputs := []examInput{
		{courseOffering.ExamPartial, 1, data.Partial1Date, data.Partial1Time, nil, nil, data.Partial1Room},
		{courseOffering.ExamPartial, 2, data.Partial2Date, data.Partial2Time, nil, nil, data.Partial2Room},
		{courseOffering.ExamFinal, 1, data.Final1Date, data.Final1Time, data.Final1RevDate, data.Final1RevTime, data.Final1Room},
		{courseOffering.ExamFinal, 2, data.Final2Date, data.Final2Time, data.Final2RevDate, data.Final2RevTime, data.Final2Room},
	}

	exams := make([]Exam, 0, 4)

	for _, in := range inputs {
		date := combineDateHour(in.date, in.hour)
		if date == nil {
			continue
		}

		exams = append(exams, Exam{
			Type:     in.t,
			Instance: in.i,
			Date:     date,
			Room:     in.room,
			Revision: combineDateHour(in.revDate, in.revHour),
		})
	}

	return exams
}

// ============================================================
// Schedule
// ============================================================

func applySchedule(
	c *Offering,
	s map[parser.WeekDay]parser.WeekDayData,
) {
	for day, data := range s {
		start := hourToTime(data.Time.Start)
		end := hourToTime(data.Time.End)

		// evita ruido
		if start == nil && end == nil && data.Room == "" {
			continue
		}

		c.AddSchedule(
			courseOffering.WeekDay(day),
			start,
			end,
			data.Room,
		)
	}
}

// ============================================================
// Time utils
// ============================================================

func combineDateHour(d *parser.Date, h *parser.Hour) *time.Time {
	if d == nil {
		return nil
	}

	hour := 0
	min := 0

	if h != nil {
		hour = h.Hour
		min = h.Minute
	}

	t := time.Date(
		d.Year,
		time.Month(d.Month),
		d.Day,
		hour,
		min,
		0,
		0,
		timezone.ParaguayTZ,
	)

	return &t
}

func hourToTime(h *parser.Hour) *time.Time {
	if h == nil {
		return nil
	}

	t := time.Date(
		0,
		time.January,
		1,
		h.Hour,
		h.Minute,
		0,
		0,
		timezone.ParaguayTZ,
	)

	return &t
}

// ============================================================
// Normalization utils
// ============================================================

var roman = [...]string{
	"", "I", "II", "III", "IV", "V",
	"VI", "VII", "VIII", "IX", "X",
	"XI", "XII", "XIII", "XIV", "XV",
	"XVI", "XVII", "XVIII", "XIX", "XX",
}

var romanToInt = map[string]int{
	"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5,
	"VI": 6, "VII": 7, "VIII": 8, "IX": 9, "X": 10,
	"XI": 11, "XII": 12, "XIII": 13, "XIV": 14, "XV": 15,
	"XVI": 16, "XVII": 17, "XVIII": 18, "XIX": 19, "XX": 20,
}

var accents = strings.NewReplacer(
	"á", "a", "Á", "A",
	"é", "e", "É", "E",
	"í", "i", "Í", "I",
	"ó", "o", "Ó", "O",
	"ú", "u", "Ú", "U",
)

var parenRegex = regexp.MustCompile(`\s*\([^()]*\)`)

func normalizeSubjectName(val string) string {
	if i := strings.IndexByte(val, '-'); i >= 0 {
		val = val[:i]
	}
	val = strings.TrimSpace(val)

	val = parenRegex.ReplaceAllString(val, "")

	fields := strings.Fields(val)
	if len(fields) == 0 {
		return ""
	}

	for i := range fields {
		f := fields[i]

		if n, err := strconv.Atoi(f); err == nil && n >= 1 && n <= 20 {
			fields[i] = roman[n]
			continue
		}

		up := strings.ToUpper(f)
		if _, ok := romanToInt[up]; ok {
			fields[i] = up
			continue
		}

		fields[i] = strings.ToLower(accents.Replace(f))
	}

	return strings.Join(fields, " ")
}
