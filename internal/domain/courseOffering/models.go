package courseOffering

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/domain/teacher"
)

// ============================================================
// Identifiers
// ============================================================

type CourseOfferingID int64

// ============================================================
// Enums
// ============================================================

type WeekDay int

const (
	Monday WeekDay = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

type CourseType int

const (
	Normal CourseType = iota
	ExamOnly
	// FUTURE: if laboratory schedules are ever available,
	// they could be represented as a new type here.
)

// ============================================================
// Value Objects
// ============================================================

type timeSlot struct {
	Start *time.Time
	End   *time.Time
}

type weekDayData struct {
	room string
	time timeSlot
}

// ============================================================
// Exam Data
// ============================================================

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

type ExamData struct {
	date     *time.Time
	revision *time.Time
	room     string
	examType ExamType
	instance ExamInstance
}

func NewExamData(date *time.Time, revDate *time.Time, room string, examType ExamType, instance ExamInstance) ExamData {
	return ExamData{
		date:     date,
		revision: revDate,
		room:     room,
		examType: examType,
		instance: instance,
	}
}

// ------------------------------------------------------------
// Exam convenience methods
// ------------------------------------------------------------

func (e ExamData) HasDate() bool {
	return e.date != nil
}

func (e ExamData) HasRevisionDate() bool {
	return e.revision != nil
}

func (e ExamData) HasHour() bool {
	if e.date == nil {
		return false
	}
	return !(e.date.Hour() == 0 && e.date.Minute() == 0)
}

func (e ExamData) HasRevHour() bool {
	if e.revision == nil {
		return false
	}
	return !(e.revision.Hour() == 0 && e.revision.Minute() == 0)
}

func (e ExamData) Date() *time.Time {
	return e.date
}

func (e ExamData) Revision() *time.Time {
	return e.revision
}

// ============================================================
// Aggregate: CourseOffering
// ============================================================

type CourseOffering struct {
	ID         CourseOfferingID
	CourseName string
	Period     period.PeriodID
	Teachers   []teacher.TeacherID
	SubjectID  academicPlan.SubjectID
	Section    string
	CourseType CourseType

	// Partial exams
	Partial1 ExamData
	Partial2 ExamData

	// Final exams
	Final1 ExamData
	Final2 ExamData

	// Weekly course schedule
	Schedule map[WeekDay]weekDayData

	// Special saturday sessions (raw representation)
	SaturdayDates string

	// Exam committee
	CommitteeMember1   string
	CommitteeMember2   string
	CommitteePresident string
}

// ------------------------------------------------------------
// Aggregate behavior
// ------------------------------------------------------------
