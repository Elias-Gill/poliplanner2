package excel

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

// ============================================================
// Entry Points & Domain Builders
// ============================================================

func buildCareerFromDTO(code string) academic.Career {
	return academic.Career{
		Code: strings.ToUpper(strings.TrimSpace(code)),
	}
}

func buildSubject(data parser.SubjectDTO) academic.Subject {
	return academic.Subject{
		Name: normalizeSubjectName(data.RawSubjectName),
		Department: academic.Department{
			Code: data.Department,
		},
	}
}

func buildCurriculum(data parser.SubjectDTO) academic.Curriculum {
	return academic.Curriculum{
		Level:    data.Level,
		Semester: data.Semester,
		Emphases: buildEmphases(data.Emphases),
		Plan: academic.Plan{
			Code: data.Plan,
		},
	}
}

func buildEmphases(data []string) []academic.Emphasis {
	var list []academic.Emphasis
	for _, s := range data {
		list = append(list, academic.Emphasis{
			Code: s,
		})
	}
	return list
}

func buildOfferingFromDTO(data parser.SubjectDTO) academic.Course {
	return academic.Course{
		Name:     strings.TrimSpace(data.RawSubjectName),
		Section:  data.Section,
		Shift:    data.Shift,
		Type:     data.CourseType,
		Exams:    buildExams(data),
		Schedule: generateSchedule(data.Schedule),
		Comitee: academic.Committee{
			President: data.CommitteePresident,
			Member1:   data.CommitteeMember1,
			Member2:   data.CommitteeMember2,
		},
	}
}

func buildTeachers(src [4]parser.TeacherDTO, count int) []academic.Teacher {
	var teachers []academic.Teacher

	for i := range count {
		// If there is no email and name, then there is no teacher here
		if src[i].Email == "" && src[i].FirstName == "" {
			continue
		}

		teachers = append(teachers, academic.Teacher{
			Title:     src[i].Title,
			FirstName: src[i].FirstName,
			LastName:  src[i].LastName,
			Email:     src[i].Email,
		})
	}

	return teachers
}

func generateSchedule(s [7]parser.WeekDayData) []academic.ClassSession {
	entries := make([]academic.ClassSession, 0, 7)

	for day, data := range s {
		if !data.Time.Start.Valid && !data.Time.End.Valid && data.Room == "" {
			continue
		}

		entries = append(entries, academic.ClassSession{
			Day:  academic.WeekDay(day),
			Room: data.Room,
			Time: academic.TimeSlot{
				Start: hourToTime(data.Time.Start),
				End:   hourToTime(data.Time.End),
			},
		})
	}

	return entries
}

// ============================================================
// Exams Parsing Engine
// ============================================================

func buildExams(data parser.SubjectDTO) []academic.Exam {
	exams := make([]academic.Exam, 0, 4)

	configs := []struct {
		eType academic.ExamType
		inst  academic.ExamInstance
		d     parser.Date
		h     parser.Hour
		rd    parser.Date
		rh    parser.Hour
		room  string
	}{
		{academic.ExamPartial, 1, data.Partial1Date, data.Partial1Time, parser.Date{}, parser.Hour{}, data.Partial1Room},
		{academic.ExamPartial, 2, data.Partial2Date, data.Partial2Time, parser.Date{}, parser.Hour{}, data.Partial2Room},
		{academic.ExamFinal, 1, data.Final1Date, data.Final1Time, data.Final1RevDate, data.Final1RevTime, data.Final1Room},
		{academic.ExamFinal, 2, data.Final2Date, data.Final2Time, data.Final2RevDate, data.Final2RevTime, data.Final2Room},
	}

	for _, cfg := range configs {
		examDate := combineDateHour(cfg.d, cfg.h)
		if examDate == nil {
			continue
		}

		exam := academic.Exam{
			Type:     cfg.eType,
			Instance: cfg.inst,
			Room:     cfg.room,
		}
		exam.SetDate(examDate)
		exam.SetRevision(combineDateHour(cfg.rd, cfg.rh))

		exams = append(exams, exam)
	}

	return exams
}

// ============================================================
// Time Utils
// ============================================================

func combineDateHour(d parser.Date, h parser.Hour) *time.Time {
	if !d.Valid {
		return nil
	}

	var hour, min int
	if h.Valid {
		hour = h.Hour
		min = h.Minute
	}

	t := time.Date(d.Year, time.Month(d.Month), d.Day, hour, min, 0, 0, timezone.ParaguayTZ)
	return &t
}

func hourToTime(h parser.Hour) *time.Time {
	if !h.Valid {
		return nil
	}

	t := time.Date(0, time.January, 1, h.Hour, h.Minute, 0, 0, timezone.ParaguayTZ)
	return &t
}

// ============================================================
// Normalization Utils
// ============================================================

var romanNumbers = [...]string{
	"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X",
	"XI", "XII", "XIII", "XIV", "XV", "XVI", "XVII", "XVIII", "XIX", "XX",
}

var romanToInt = map[string]int{
	"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5, "VI": 6, "VII": 7, "VIII": 8, "IX": 9, "X": 10,
	"XI": 11, "XII": 12, "XIII": 13, "XIV": 14, "XV": 15, "XVI": 16, "XVII": 17, "XVIII": 18, "XIX": 19, "XX": 20,
}

var accentsReplacer = strings.NewReplacer(
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

	val = parenRegex.ReplaceAllString(val, "")
	fields := strings.Fields(val)
	if len(fields) == 0 {
		return ""
	}

	for i, f := range fields {
		// Convert numbers to roman
		if n, err := strconv.Atoi(f); err == nil && n >= 1 && n <= 20 {
			fields[i] = romanNumbers[n]
			continue
		}

		// Capitalize roman numbers
		upperField := strings.ToUpper(f)
		if _, isRoman := romanToInt[upperField]; isRoman {
			fields[i] = upperField
			continue
		}

		// Remove accents and to lower all the word
		cleaned := accentsReplacer.Replace(strings.ToLower(f))

		// Keep just the first letter uppercase (special Title case)
		if i == 0 {
			cleaned = capitalizeFirstRune(cleaned)
		}

		fields[i] = cleaned
	}

	return strings.Join(fields, " ")
}

func capitalizeFirstRune(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
