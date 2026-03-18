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

func buildOfferingFromDTO(career string, periodID period.PeriodID, data parser.SubjectDTO, planLoader *metadata.AcademicPlanLoader) Offering {
	// Complete metadata if possible
	if planLoader != nil {
		if data.Semester == 0 {
			if m, err := planLoader.FindSubject(data.RawSubjectName); err == nil {
				data.Semester = m.Semester
			}
		}
	}

	subject := subject{
		Career:     strings.ToUpper(strings.TrimSpace(career)),
		Name:       normalizeSubjectName(data.RawSubjectName),
		Department: data.Department,
		Level:      data.Level,
		Semester:   data.Semester,
	}

	teachers := make([]teacher, 0, len(data.Teachers))
	for _, t := range data.Teachers {
		teachers = append(teachers, teacher{
			Name:      t.FirstName + " " + t.LastName,
			Email:     t.Email,
			SearchKey: generateSearchKey(t.FirstName, t.LastName),
		})
	}

	cType := courseOffering.Normal
	if data.CourseType == parser.ExamOnlyCourse {
		cType = courseOffering.ExamOnly
	}

	offer := Offering{
		CourseName: data.RawSubjectName,
		Period:     periodID,
		Teachers:   teachers,
		Subject:    subject,
		Section:    data.Section,
		CourseType: cType,

		Partial1: convertToExamData(data.Partial1Date, data.Partial1Time, nil, nil, data.Partial1Room),
		Partial2: convertToExamData(data.Partial2Date, data.Partial2Time, nil, nil, data.Partial2Room),

		Final1: convertToExamData(data.Final1Date, data.Final1Time, data.Final1RevDate, data.Final1RevTime, data.Final1Room),
		Final2: convertToExamData(data.Final2Date, data.Final2Time, data.Final2RevDate, data.Final2RevTime, data.Final2Room),

		CommitteePresident: data.CommitteePresident,
		CommitteeMember1:   data.CommitteeMember1,
		CommitteeMember2:   data.CommitteeMember2,
	}

	applySchedule(&offer, data.Schedule)

	return offer
}

// ---------------------------------------
// -    Utils and conversion methods     -
// ---------------------------------------

func convertToExamData(
	date *parser.Date,
	hour *parser.Hour,
	revDate *parser.Date,
	revTime *parser.Hour,
	room string,
) examData {
	exam := combineDateHour(date, hour)
	revision := combineDateHour(revDate, revTime)

	return newExamData(exam, revision, room)
}

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

	t := time.Date(d.Year, time.Month(d.Month), d.Day, hour, min, 0, 0, timezone.ParaguayTZ)

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

func applySchedule(
	c *Offering,
	s map[parser.WeekDay]parser.WeekDayData,
) {
	for day, data := range s {
		start := hourToTime(data.Time.Start)
		end := hourToTime(data.Time.End)

		c.AddSchedule(
			courseOffering.WeekDay(day),
			start,
			end,
			data.Room)
	}
}

func generateSearchKey(firstName, lastName string) string {
	first := ""
	if fields := strings.Fields(firstName); len(fields) > 0 {
		first = strings.ToLower(fields[0])
	} else {
		// first name is empty
		return ""
	}

	last := ""
	if fields := strings.Fields(lastName); len(fields) > 0 {
		last = strings.ToLower(fields[0])
	} else {
		// last name is empty
		return ""
	}

	return strings.Join([]string{first, last}, "_")
}
