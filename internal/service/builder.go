package service

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/domain/teacher"
	"github.com/elias-gill/poliplanner2/internal/parser"
	"github.com/elias-gill/poliplanner2/internal/parser/metadata"
)

func BuildSubjectFromDTO(data parser.SubjectDTO, planLoader *metadata.AcademicPlanLoader) academicPlan.Subject {
	// Complete data if possible
	if planLoader != nil {
		if data.Semester == 0 {
			if m, err := planLoader.FindSubject(data.RawSubjectName); err == nil {
				data.Semester = m.Semester
			}
		}
	}

	return academicPlan.Subject{
		Name:       normalizeSubjectName(data.RawSubjectName),
		Department: data.Department,
		Level:      data.Level,
		Semester:   data.Semester,
	}
}

func BuildTeachersFromDTO(data parser.SubjectDTO) []teacher.Teacher {
	var teachers = make([]teacher.Teacher, len(data.Teachers))
	for _, t := range data.Teachers {
		teachers = append(teachers,
			teacher.Teacher{
				Name:  t.FirstName + t.LastName,
				Email: t.Email,
				// TODO: que carajos hago con la search key
			})
	}
	return teachers
}

func BuildCourseOfferingFromDTO(
	data parser.SubjectDTO,
	subjectID academicPlan.SubjectID,
	periodID period.PeriodID,
	teachersIDs []teacher.TeacherID,
) courseOffering.CourseOffering {
	cType := courseOffering.Normal
	if data.CourseType == parser.ExamOnlyCourse {
		cType = courseOffering.ExamOnly
	}

	return courseOffering.CourseOffering{
		CourseName: data.RawSubjectName,
		Period:     periodID,
		Teachers:   teachersIDs,
		SubjectID:  subjectID,
		Section:    data.Section,
		CourseType: cType,

		Partial1: courseOffering.NewExamData(data.Partial1Date, data.Partial1Time, data.Partial1Room),
	}
}

func BuildAggregatesFromDTO(data parser.SubjectDTO, planLoader *metadata.AcademicPlanLoader) (
	academicPlan.Subject,
	courseOffering.CourseOffering,
	[]teacher.Teacher,
) {
	teachers := make([]teacher.Teacher, len(data.Teachers))
	for _, t := range data.Teachers {
		teachers = append(teachers, teacher.Teacher{
			Name:  t.FirstName + " " + t.LastName,
			Email: t.Email,
			// FIX: ver que hacer con la searching key
		})
	}

	course := courseOffering.CourseOffering{
		CourseName: data.RawSubjectName,
		Section:    data.Section,
	}

	if course {

	}

	return subject, course, teachers
}

// -----------------
// -   UTILS       -
// -----------------

// The max amount of roman numbers we have is normally 10, but just to prevent any
// potential errors
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

func normalizeSubjectName(val string) string {
	// Parte antes del -
	if i := strings.IndexByte(val, '-'); i >= 0 {
		val = val[:i]
	}
	val = strings.TrimSpace(val)

	// delete perenthesis (), (*), (**)
	val = regexp.MustCompile(`\s*\([^()]*\)`).ReplaceAllString(val, "")

	fields := strings.Fields(val)
	if len(fields) == 0 {
		return ""
	}

	// Normalize arabic to roman numbers
	for i := range fields {
		f := fields[i]

		if n, err := strconv.Atoi(f); err == nil && n >= 1 && n <= 20 {
			fields[i] = roman[n]
			continue
		}

		// Uppercase roman numbers
		up := strings.ToUpper(f)
		if _, ok := romanToInt[up]; ok {
			fields[i] = up
			continue
		}

		// Delete accents and lowercase all non roman number fields
		fields[i] = strings.ToLower(accents.Replace(f))
	}

	return strings.Join(fields, " ")
}
