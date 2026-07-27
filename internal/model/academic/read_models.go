package academic

import (
	"fmt"
	"strings"
)

// ==========================
// 	 Schedule creation view
// ==========================

type CareerCurriculumView struct {
	Plans     []Plan
	Semesters []int
	Levels    []int
	Subjects  []CurriculumSubjectItem
}

type CurriculumSubjectItem struct {
	ID       CurriculumID
	Plan     string
	Semester int
	Level    int
	Name     string
}

type CourseSummaryView struct {
	ID            CourseID
	Section       string
	Shift         string
	Name          string
	Type          CourseType
	Teachers      []Teacher
	Schedules     []ClassSession
	Exams         []Exam
	SaturdayDates string
	Committee     Committee
}

// FormattedSchedule directly used inside HTML templates with "{{ .FormattedSchedule }}"
// WARNING: modify it's name carefully
func (c CourseSummaryView) FormattedSchedule() string {
	if len(c.Schedules) == 0 {
		return "Sin horario asignado"
	}

	var parts []string
	for _, s := range c.Schedules {
		dayStr := s.Day.String()

		var startStr, endStr string
		if s.Time.Start != nil {
			startStr = s.Time.Start.Format("15:04")
		} else {
			startStr = "--:--"
		}

		if s.Time.End != nil {
			endStr = s.Time.End.Format("15:04")
		} else {
			endStr = "--:--"
		}

		parts = append(parts, fmt.Sprintf("%s %s-%s", dayStr, startStr, endStr))
	}

	return strings.Join(parts, " | ")
}
