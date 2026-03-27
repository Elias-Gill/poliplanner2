package courseOffering

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/teacher"
)

type SectionID int64

type Section struct {
	ID         SectionID
	Section    string
	CourseName string
	Teachers   []teacher.Teacher
	Type       CourseType
}

type OfferList struct {
	Subject string
	Offer   []Section
}

// =========================

type ExamClass struct {
	CourseName string
	Room       string
	Date       time.Time
	Revision   *time.Time
	Type       ExamType
	Instance   ExamInstance
}

type ExamsScheduleView struct {
	Partial1 []ExamClass
	Partial2 []ExamClass
	Final1   []ExamClass
	Final2   []ExamClass
}

// =========================

type CourseClass struct {
	CourseID CourseOfferingID
	Name     string

	Room  string
	Start time.Time
	End   time.Time
}

type CoursesScheduleView struct {
	Monday    []CourseClass
	Tuesday   []CourseClass
	Wednesday []CourseClass
	Thursday  []CourseClass
	Friday    []CourseClass
	Saturday  []CourseClass
}

// =========================

type TeacherInfo struct {
	Name  string
	Email string
}

type CourseSummary struct {
	Name       string
	Teachers   []TeacherInfo
	Section    string
	CourseType CourseType

	// Special saturday sessions (raw representation)
	SaturdayDates string

	// Exam committee
	CommitteeMember1   string
	CommitteeMember2   string
	CommitteePresident string
}
