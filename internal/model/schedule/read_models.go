package schedule

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/user"
)

type ScheduleDetails struct {
	ID        ScheduleID
	Owner     user.UserID
	Title     string
	CreatedAt time.Time
	Courses   []academic.CourseSummaryView
}

type ScheduleSummaryView struct {
	ID    ScheduleID
	Title string
}

type StudentScheduleView struct {
	Weekly WeekScheduleView
	Exams  ExamMapView
	Info   []CourseDetailView
}

type WeekScheduleView struct {
	Monday    []ClassSlotView
	Tuesday   []ClassSlotView
	Wednesday []ClassSlotView
	Thursday  []ClassSlotView
	Friday    []ClassSlotView
	Saturday  []ClassSlotView
}

type ClassSlotView struct {
	Course string
	Room   string
	Time   academic.TimeSlot
}

type ExamSlotView struct {
	CourseName string
	Room       string
	Date       string
	Revision   string
}

type ExamMapView struct {
	Partial1 []ExamSlotView
	Partial2 []ExamSlotView
	Final1   []ExamSlotView
	Final2   []ExamSlotView
}

type TeacherContactView struct {
	Name  string
	Email string
}

type CourseDetailView struct {
	Name               string
	Section            string
	Shift              string
	Type               academic.CourseType
	Teachers           []TeacherContactView
	SaturdayDates      string
	CommitteePresident string
	CommitteeMember1   string
	CommitteeMember2   string
}
