package academic

import (
	"fmt"
	"strings"
	"time"
)

// ====================
// Important constants
// ====================
// WARNING: Do not touch any of this constants unless you actually know what you are doing
// or it will destroy data consistency with the database models.
//
// Modify carefully.

// ----------- Courses ----------------

// CourseType defines the operational modality of a course within the university.
type CourseType int

const (
	// Normal represents a standard course with regular weekly lectures and exams.
	Normal CourseType = 0

	// ExamOnly represents a restricted course modality for students who have already
	// preserved their "exam signature" (firma) from a previous semester. No regular
	// classes are taught; it is used strictly to manage and sit for final exams.
	ExamOnly CourseType = 1

	// FUTURE: If laboratory schedules are ever available, they could be represented
	// as a new types here
)

func (d CourseType) String() string {
	switch d {
	case ExamOnly:
		return "Solo Examen Final"
	default:
		return "Normal"
	}
}

// ----------- Weekly schedule ----------------

type WeekDay int

const (
	Monday    WeekDay = 1
	Tuesday   WeekDay = 2
	Wednesday WeekDay = 3
	Thursday  WeekDay = 4
	Friday    WeekDay = 5
	Saturday  WeekDay = 6
)

func (d WeekDay) String() string {
	switch d {
	case Monday:
		return "Lunes"
	case Tuesday:
		return "Martes"
	case Wednesday:
		return "Miércoles"
	case Thursday:
		return "Jueves"
	case Friday:
		return "Viernes"
	case Saturday:
		return "Sábado"
	default:
		return ""
	}
}

// ----------- Exams ----------------

type ExamID int64

type ExamType string

const (
	ExamPartial ExamType = "partial"
	ExamFinal   ExamType = "final"
)

type ExamInstance int

const (
	Instance1 ExamInstance = 1
	Instance2 ExamInstance = 2
)

// ==================
// Basic course data
// ==================

type CourseID int64

// Course represents a concrete, operational instance of a Curriculum subject.
//
// A Course name is not unique, as multiple sections (classes) of the same course can coexist
// simultaneously.
//
// Each Course is bound to a specific academic period and curriculum, and encapsulates its
// assigned section, teaching staff (Teachers), evaluation board (Committee), weekly schedules
// (including special Saturday session dates), and exam dates.
type Course struct {
	Name    string // Course name (non-unique across different sections).
	Type    CourseType
	Section string // Specific class section or group identifier.
	Shift   string // Shift on wich the class is imparted

	Schedule      []ClassSession
	SaturdayDates string // Specific dates for Saturday classes, if applicable.

	Exams   []Exam
	Comitee Committee
}

type Committee struct {
	// Revision
	President string
	Member1   string
	Member2   string
}

// --------- Course schedule --------------

type ClassSession struct {
	Day  WeekDay
	Room string
	Time TimeSlot
}

type TimeSlot struct {
	Start *time.Time
	End   *time.Time
}

func (t TimeSlot) String() string {
	const layout = "15:04"

	startStr := "N/A"
	if t.Start != nil {
		startStr = t.Start.Format(layout)
	}

	endStr := "N/A"
	if t.End != nil {
		endStr = t.End.Format(layout)
	}

	return fmt.Sprintf("%shs a %shs", startStr, endStr)
}

// ------------- Exams --------------------

type Exam struct {
	date     *time.Time
	revision *time.Time
	Room     string
	Type     ExamType
	Instance ExamInstance
}

func NewExam(date *time.Time, revDate *time.Time, room string, examType ExamType, instance ExamInstance) Exam {
	return Exam{
		date:     date,
		revision: revDate,
		Room:     room,
		Type:     examType,
		Instance: instance,
	}
}

func (e Exam) HasDate() bool {
	return e.date != nil
}

func (e *Exam) SetDate(t *time.Time) {
	e.date = t
}

func (e Exam) HasRevisionDate() bool {
	return e.revision != nil
}

func (e *Exam) SetRevision(t *time.Time) {
	e.revision = t
}

func (e Exam) HasHour() bool {
	if e.date == nil {
		return false
	}
	return !(e.date.Hour() == 0 && e.date.Minute() == 0)
}

func (e Exam) HasRevHour() bool {
	if e.revision == nil {
		return false
	}
	return !(e.revision.Hour() == 0 && e.revision.Minute() == 0)
}

func (e Exam) Date() *time.Time {
	return e.date
}

func (e Exam) Revision() *time.Time {
	return e.revision
}

// ===========================
// Utility functions
// ===========================

func (c Course) String() string {
	var sb strings.Builder

	sb.WriteString("Offering{ ")

	sb.WriteString("CourseName=")
	sb.WriteString(c.Name)
	sb.WriteString(" ")
	sb.WriteString(" ")
	sb.WriteString("Section=")
	sb.WriteString(c.Section)
	sb.WriteString(" ")
	sb.WriteString("CourseType=")
	fmt.Fprintf(&sb, "%v", c.Type)
	sb.WriteString(" ")

	sb.WriteString("Schedule=[")

	for i, s := range c.Schedule {
		if i > 0 {
			sb.WriteString(", ")
		}

		start := ""
		end := ""

		if s.Time.Start != nil {
			start = s.Time.Start.String()
		}
		if s.Time.End != nil {
			end = s.Time.End.String()
		}

		fmt.Fprintf(&sb, "day=%v,start=%s,end=%s,room=%s", s.Day, start, end, s.Room)
	}

	sb.WriteString("] ")

	sb.WriteString("}")

	return sb.String()
}
