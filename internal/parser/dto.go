package parser

import (
	"strings"
	"sync"
	"time"
)

type CourseType int

const ExamOnlyCourse CourseType = 1
const NormalCourse CourseType = 0

// Buffer pool for time formatting
var timeBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 10)
	},
}

type TeacherDTO struct {
	Title     string
	FirstName string
	LastName  string
	Email     string
}

type weekDayData struct {
	Room string
	Time TimeSlot
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

type TimeSlot struct {
	Start string
	End   string
}

type SubjectDTO struct {
	// General info
	Department string
	Semester   int
	Level      int
	Section    string

	// This is used as the name of the table "cursos", which is the final agreggate
	// with all schedule information of a subject in a specific time period.
	RawSubjectName string

	// Course type can be:
	// - Normal course
	// - Final exam only
	CourseType CourseType

	// Teachers info
	Teachers []TeacherDTO

	// Exams
	// First partial
	Partial1Date *time.Time
	Partial1Time *time.Time
	Partial1Room string

	// Second partial
	Partial2Date *time.Time
	Partial2Time *time.Time
	Partial2Room string

	// First final
	Final1Date    *time.Time
	Final1Time    *time.Time
	Final1Room    string
	Final1RevDate *time.Time
	Final1RevTime *time.Time

	// Second final
	Final2Date    *time.Time
	Final2Time    *time.Time
	Final2Room    string
	Final2RevDate *time.Time
	Final2RevTime *time.Time

	// Weekly schedule
	Schedule      map[WeekDay]weekDayData
	SaturdayDates string

	// Committee
	CommitteePresident string
	CommitteeMember1   string
	CommitteeMember2   string
}

// -----------------------------
// Setters with cleaning methods
// -----------------------------

func (s *SubjectDTO) SetDepartment(val string) {
	s.Department = strings.TrimSpace(val)
}

func (s *SubjectDTO) SetSubjectName(val string) {
	s.RawSubjectName = strings.TrimSpace(val)

	// Set course type based on the name
	s.CourseType = NormalCourse
	// If contains a (*) it is a closed grade with only final exam
	for i := len(val) - 1; i >= 0; i-- {
		if rune(val[i]) == '*' {
			s.CourseType = ExamOnlyCourse
			break
		}
	}
}

func (s *SubjectDTO) SetSemester(val string) {
	s.Semester = convertStringToNumber(val)
}

func (s *SubjectDTO) SetLevel(val string) {
	s.Level = convertStringToNumber(val)
}

func (s *SubjectDTO) SetSection(val string) {
	s.Section = strings.TrimSpace(val)
}

func (s *SubjectDTO) SetTeachersFirtNames(firstNames string) {
	list := strings.Split(firstNames, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].FirstName = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) SetTeachersLastNames(secondNames string) {
	list := strings.Split(secondNames, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].LastName = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) SetTeachersTitles(titles string) {
	list := strings.Split(titles, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].Title = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) SetTeachersEmails(emails string) {
	list := strings.Split(emails, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].Email = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) ensureTeachersLen(n int) {
	if len(s.Teachers) < n {
		s.Teachers = append(s.Teachers, make([]TeacherDTO, n-len(s.Teachers))...)
	}
}

func (s *SubjectDTO) SetPartial1Date(val string) {
	s.Partial1Date = parseDate(val)
}

func (s *SubjectDTO) SetPartial1Time(val string) {
	s.Partial1Time = cleanTime(val)
}

func (s *SubjectDTO) SetPartial1Room(val string) {
	s.Partial1Room = val
}

func (s *SubjectDTO) SetPartial2Date(val string) {
	s.Partial2Date = parseDate(val)
}

func (s *SubjectDTO) SetPartial2Time(val string) {
	s.Partial2Time = cleanTime(val)
}

func (s *SubjectDTO) SetPartial2Room(val string) {
	s.Partial2Room = val
}

func (s *SubjectDTO) SetFinal1Date(val string) {
	s.Final1Date = parseDate(val)
}

func (s *SubjectDTO) SetFinal1Time(val string) {
	s.Final1Time = cleanTime(val)
}

func (s *SubjectDTO) SetFinal1Room(val string) {
	s.Final1Room = val
}

func (s *SubjectDTO) SetFinal1RevDate(val string) {
	s.Final1RevDate = parseDate(val)
}

func (s *SubjectDTO) SetFinal1RevTime(val string) {
	s.Final1RevTime = cleanTime(val)
}

func (s *SubjectDTO) SetFinal2Date(val string) {
	s.Final2Date = parseDate(val)
}

func (s *SubjectDTO) SetFinal2Time(val string) {
	s.Final2Time = cleanTime(val)
}

func (s *SubjectDTO) SetFinal2Room(val string) {
	s.Final2Room = val
}

func (s *SubjectDTO) SetFinal2RevDate(val string) {
	s.Final2RevDate = parseDate(val)
}

func (s *SubjectDTO) SetFinal2RevTime(val string) {
	s.Final2RevTime = cleanTime(val)
}

func (s *SubjectDTO) SetCommitteePresident(val string) {
	s.CommitteePresident = val
}

func (s *SubjectDTO) SetCommitteeMember1(val string) {
	s.CommitteeMember1 = val
}

func (s *SubjectDTO) SetCommitteeMember2(val string) {
	s.CommitteeMember2 = val
}

func (s *SubjectDTO) SetSaturdayDates(dates string) {
	s.SaturdayDates = dates
}

func (s *SubjectDTO) SetDayTime(day WeekDay, val string) {
	s.Schedule[day] = weekDayData{
		Room: s.Schedule[day].Room,
		Time: convertIntoTimeSlot(val),
	}
}

func (s *SubjectDTO) SetDayRoom(day WeekDay, room string) {
	s.Schedule[day] = weekDayData{
		Room: strings.TrimSpace(room),
		Time: s.Schedule[day].Time,
	}
}

func (d *SubjectDTO) Reset() {
	*d = SubjectDTO{}
	d.Schedule = make(map[WeekDay]weekDayData, 6) // pre-initialize map
	d.Teachers = d.Teachers[:0]                   // reutilize the memmory slice
}
