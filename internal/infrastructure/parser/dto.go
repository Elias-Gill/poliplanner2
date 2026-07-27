package parser

import (
	"strings"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type Hour struct {
	Hour   int
	Minute int
	Valid  bool
}

type Date struct {
	Year  int
	Month int
	Day   int
	Valid bool
}

type TimeSlot struct {
	Start Hour
	End   Hour
}

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

type SubjectDTO struct {
	Department     string
	Plan           string
	Emphases       []string
	Semester       int
	Level          int
	Section        string
	Shift          string
	RawSubjectName string
	CourseType     academic.CourseType

	// PERFORMANCE: Fixed size array removes unneeded allocations
	Teachers     [4]TeacherDTO
	TeacherCount int

	Partial1Date Date
	Partial1Time Hour
	Partial1Room string

	Partial2Date Date
	Partial2Time Hour
	Partial2Room string

	Final1Date    Date
	Final1Time    Hour
	Final1Room    string
	Final1RevDate Date
	Final1RevTime Hour

	Final2Date    Date
	Final2Time    Hour
	Final2Room    string
	Final2RevDate Date
	Final2RevTime Hour

	// PERFORMANCE: Fixed size array removes unneeded allocations
	Schedule      [7]WeekDayData
	SaturdayDates string

	CommitteePresident string
	CommitteeMember1   string
	CommitteeMember2   string
}

func (s *SubjectDTO) SetDepartment(val string) {
	s.Department = strings.ToUpper(strings.TrimSpace(val))
}

func (s *SubjectDTO) SetEmphases(val string) {
	val = strings.TrimSpace(val)
	if val == "" {
		s.Emphases = nil
		return
	}

	parts := strings.Split(val, ",")
	emphases := make([]string, 0, len(parts))

	for _, p := range parts {
		var sb strings.Builder
		for _, r := range p {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				sb.WriteRune(unicode.ToUpper(r))
			}
		}

		cleaned := sb.String()
		if cleaned != "" {
			emphases = append(emphases, cleaned)
		}
	}

	s.Emphases = emphases
}

func (s *SubjectDTO) SetPlan(val string) {
	s.Plan = strings.ToLower(strings.TrimSpace(val))
}

func (s *SubjectDTO) SetSubjectName(val string) {
	s.RawSubjectName = strings.TrimSpace(val)
	s.CourseType = academic.Normal

	for i := len(val) - 1; i >= 0; i-- {
		if val[i] == '*' {
			s.CourseType = academic.ExamOnly
			break
		}
	}
}

func (s *SubjectDTO) SetSemester(val string) { s.Semester = convertStringToNumber(val) }

func (s *SubjectDTO) SetLevel(val string) { s.Level = convertStringToNumber(val) }

func (s *SubjectDTO) SetSection(val string) {
	s.Section = strings.TrimSpace(strings.ToUpper(val))
}

func (s *SubjectDTO) SetShift(val string) {
	s.Shift = strings.TrimSpace(strings.ToUpper(val))
}

// scanLines parses a multi-line input string to extract and index up to 4 non-empty lines
// via the assign callback, updating TeacherCount to track the maximum number of teachers found.
func (s *SubjectDTO) scanLines(input string, assign func(idx int, line string)) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}
	idx := 0
	// Hard limit: processes a maximum of 4 non-empty lines
	for idx < 4 {
		next := strings.IndexByte(input, '\n')
		var line string
		if next == -1 {
			line = strings.TrimSpace(input)
		} else {
			line = strings.TrimSpace(input[:next])
		}

		// Empty lines are skipped entirely and do not increment the index
		if line != "" {
			assign(idx, line)
			idx++
		}
		if next == -1 {
			break
		}
		input = input[next+1:]
	}
	// Acts as a high-water mark; TeacherCount only updates if the new count is higher
	if idx > s.TeacherCount {
		s.TeacherCount = idx
	}
}

func (s *SubjectDTO) SetTeachersFirtNames(v string) {
	s.scanLines(v, func(i int, l string) { s.Teachers[i].FirstName = l })
}
func (s *SubjectDTO) SetTeachersLastNames(v string) {
	s.scanLines(v, func(i int, l string) { s.Teachers[i].LastName = l })
}
func (s *SubjectDTO) SetTeachersTitles(v string) {
	s.scanLines(v, func(i int, l string) { s.Teachers[i].Title = l })
}
func (s *SubjectDTO) SetTeachersEmails(v string) {
	s.scanLines(v, func(i int, l string) { s.Teachers[i].Email = l })
}

func (s *SubjectDTO) SetPartial1Date(val string)  { s.Partial1Date = parseDate(val) }
func (s *SubjectDTO) SetPartial1Time(val string)  { s.Partial1Time = parseTime(val) }
func (s *SubjectDTO) SetPartial1Room(val string)  { s.Partial1Room = val }
func (s *SubjectDTO) SetPartial2Date(val string)  { s.Partial2Date = parseDate(val) }
func (s *SubjectDTO) SetPartial2Time(val string)  { s.Partial2Time = parseTime(val) }
func (s *SubjectDTO) SetPartial2Room(val string)  { s.Partial2Room = val }
func (s *SubjectDTO) SetFinal1Date(val string)    { s.Final1Date = parseDate(val) }
func (s *SubjectDTO) SetFinal1Time(val string)    { s.Final1Time = parseTime(val) }
func (s *SubjectDTO) SetFinal1Room(val string)    { s.Final1Room = val }
func (s *SubjectDTO) SetFinal1RevDate(val string) { s.Final1RevDate = parseDate(val) }
func (s *SubjectDTO) SetFinal1RevTime(val string) { s.Final1RevTime = parseTime(val) }
func (s *SubjectDTO) SetFinal2Date(val string)    { s.Final2Date = parseDate(val) }
func (s *SubjectDTO) SetFinal2Time(val string)    { s.Final2Time = parseTime(val) }
func (s *SubjectDTO) SetFinal2Room(val string)    { s.Final2Room = val }
func (s *SubjectDTO) SetFinal2RevDate(val string) { s.Final2RevDate = parseDate(val) }
func (s *SubjectDTO) SetFinal2RevTime(val string) { s.Final2RevTime = parseTime(val) }

func (s *SubjectDTO) SetDayTime(day academic.WeekDay, val string) {
	s.Schedule[day] = WeekDayData{Room: s.Schedule[day].Room, Time: parseTimeSlot(val)}
}
func (s *SubjectDTO) SetDayRoom(day academic.WeekDay, room string) {
	s.Schedule[day] = WeekDayData{Room: strings.TrimSpace(room), Time: s.Schedule[day].Time}
}
func (s *SubjectDTO) SetSaturdayDates(dates string)    { s.SaturdayDates = dates }
func (s *SubjectDTO) SetCommitteePresident(val string) { s.CommitteePresident = val }
func (s *SubjectDTO) SetCommitteeMember1(val string)   { s.CommitteeMember1 = val }
func (s *SubjectDTO) SetCommitteeMember2(val string)   { s.CommitteeMember2 = val }

func (d *SubjectDTO) Reset() {
	*d = SubjectDTO{} // Stack cleaning function
}
