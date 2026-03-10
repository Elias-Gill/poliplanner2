package courseOffering

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/domain/teacher"
)

type CourseOfferingID int64

type timeSlot struct {
	Start *time.Time
	End   *time.Time
}

type examData struct {
	date *time.Time
	time *time.Time
	room string
}

func NewExamData(date *time.Time, examTime *time.Time, room string) examData {
	return examData{
		date: date,
		time: examTime,
		room: room,
	}
}

type weekDayData struct {
	room string
	time timeSlot
}

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
	// FUTURE: si alguna vez se consigue sacar los hroarios de laboratorio, entonces podria ir
	// agregar aca
)

type CourseOffering struct {
	ID         CourseOfferingID
	CourseName string
	Period     period.PeriodID
	Teachers   []teacher.TeacherID
	SubjectID  academicPlan.SubjectID
	Section    string
	CourseType CourseType

	// First partial
	Partial1 examData
	Partial2 examData

	Final1 examData
	Final2 examData

	// Weekly course schedule
	Schedule map[WeekDay]weekDayData

	SaturdayDates string

	CommitteeMember1   string
	CommitteeMember2   string
	CommitteePresident string
}

func (c *CourseOffering) AddTeacher(id teacher.TeacherID) {
	c.Teachers = append(c.Teachers, id)
}

func (c *CourseOffering) AddSchedule(day WeekDay, start *time.Time, end *time.Time, room string) {
	c.Schedule[day] = weekDayData{
		room: room,
		time: timeSlot{
			Start: start,
			End:   end,
		},
	}
}
