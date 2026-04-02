package parser

import (
	"strings"
)

// ============================================================
// Enums
// ============================================================

type CourseType int

const (
	NormalCourse   CourseType = 0
	ExamOnlyCourse CourseType = 1
)

type WeekDay int

const (
	Monday WeekDay = 1 + iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// ============================================================
// Time primitives
// ============================================================

type Hour struct {
	Hour   int
	Minute int
}

type Date struct {
	Year  int
	Month int
	Day   int
}

type TimeSlot struct {
	Start *Hour
	End   *Hour
}

// ============================================================
// Supporting DTOs
// ============================================================

type TeacherDTO struct {
	Title     string
	FirstName string
	LastName  string
	Email     string
}

type WeekDayData struct {
	Room string
	Time TimeSlot
}

// ============================================================
// Main DTO: SubjectDTO
// ============================================================

type SubjectDTO struct {

	// --------------------------------------------------------
	// General course information
	// --------------------------------------------------------

	Department string
	Semester   int
	Level      int
	Section    string

	// This is used as the name of the table "cursos", which is the final aggregate
	// with all schedule information of a subject in a specific time period.
	RawSubjectName string

	// Course type can be:
	// - Normal course
	// - Final exam only
	CourseType CourseType

	// --------------------------------------------------------
	// Teachers
	// --------------------------------------------------------

	Teachers []TeacherDTO

	// --------------------------------------------------------
	// Exams
	// --------------------------------------------------------

	// First partial
	Partial1Date *Date
	Partial1Time *Hour
	Partial1Room string

	// Second partial
	Partial2Date *Date
	Partial2Time *Hour
	Partial2Room string

	// First final
	Final1Date    *Date
	Final1Time    *Hour
	Final1Room    string
	Final1RevDate *Date
	Final1RevTime *Hour

	// Second final
	Final2Date    *Date
	Final2Time    *Hour
	Final2Room    string
	Final2RevDate *Date
	Final2RevTime *Hour

	// --------------------------------------------------------
	// Weekly schedule
	// --------------------------------------------------------

	Schedule      map[WeekDay]WeekDayData
	SaturdayDates string

	// --------------------------------------------------------
	// Exam committee
	// --------------------------------------------------------

	CommitteePresident string
	CommitteeMember1   string
	CommitteeMember2   string
}

// ============================================================
// Setters with cleaning methods
// ============================================================

func (s *SubjectDTO) SetDepartment(val string) {
	s.Department = strings.ToUpper(strings.TrimSpace(val))
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

// ============================================================
// Teacher setters
// ============================================================

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

// ============================================================
// Exam setters
// ============================================================

func (s *SubjectDTO) SetPartial1Date(val string) {
	s.Partial1Date = parseDate(val)
}

func (s *SubjectDTO) SetPartial1Time(val string) {
	s.Partial1Time = parseTime(val)
}

func (s *SubjectDTO) SetPartial1Room(val string) {
	s.Partial1Room = val
}

func (s *SubjectDTO) SetPartial2Date(val string) {
	s.Partial2Date = parseDate(val)
}

func (s *SubjectDTO) SetPartial2Time(val string) {
	s.Partial2Time = parseTime(val)
}

func (s *SubjectDTO) SetPartial2Room(val string) {
	s.Partial2Room = val
}

func (s *SubjectDTO) SetFinal1Date(val string) {
	s.Final1Date = parseDate(val)
}

func (s *SubjectDTO) SetFinal1Time(val string) {
	s.Final1Time = parseTime(val)
}

func (s *SubjectDTO) SetFinal1Room(val string) {
	s.Final1Room = val
}

func (s *SubjectDTO) SetFinal1RevDate(val string) {
	s.Final1RevDate = parseDate(val)
}

func (s *SubjectDTO) SetFinal1RevTime(val string) {
	s.Final1RevTime = parseTime(val)
}

func (s *SubjectDTO) SetFinal2Date(val string) {
	s.Final2Date = parseDate(val)
}

func (s *SubjectDTO) SetFinal2Time(val string) {
	s.Final2Time = parseTime(val)
}

func (s *SubjectDTO) SetFinal2Room(val string) {
	s.Final2Room = val
}

func (s *SubjectDTO) SetFinal2RevDate(val string) {
	s.Final2RevDate = parseDate(val)
}

func (s *SubjectDTO) SetFinal2RevTime(val string) {
	s.Final2RevTime = parseTime(val)
}

// ============================================================
// Schedule setters
// ============================================================

func (s *SubjectDTO) SetDayTime(day WeekDay, val string) {
	s.Schedule[day] = WeekDayData{
		Room: s.Schedule[day].Room,
		Time: parseTimeSlot(val),
	}
}

func (s *SubjectDTO) SetDayRoom(day WeekDay, room string) {
	s.Schedule[day] = WeekDayData{
		Room: strings.TrimSpace(room),
		Time: s.Schedule[day].Time,
	}
}

func (s *SubjectDTO) SetSaturdayDates(dates string) {
	s.SaturdayDates = dates
}

// ============================================================
// Committee setters
// ============================================================

func (s *SubjectDTO) SetCommitteePresident(val string) {
	s.CommitteePresident = val
}

func (s *SubjectDTO) SetCommitteeMember1(val string) {
	s.CommitteeMember1 = val
}

func (s *SubjectDTO) SetCommitteeMember2(val string) {
	s.CommitteeMember2 = val
}

// ============================================================
// Lifecycle helpers
// ============================================================

func (d *SubjectDTO) Reset() {
	*d = SubjectDTO{}
	d.Schedule = make(map[WeekDay]WeekDayData, 6) // pre-initialize map
	d.Teachers = d.Teachers[:0]                   // reuse slice memory
}
