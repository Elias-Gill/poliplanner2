package parser

import (
	"strings"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/commons"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type WeekDayData struct {
	Room string
	Time commons.TimeSlot
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

	Teachers     [4]commons.TeacherDTO
	TeacherCount int

	Partial1Date commons.Date
	Partial1Time commons.Hour
	Partial1Room string

	Partial2Date commons.Date
	Partial2Time commons.Hour
	Partial2Room string

	Final1Date    commons.Date
	Final1Time    commons.Hour
	Final1Room    string
	Final1RevDate commons.Date
	Final1RevTime commons.Hour

	Final2Date    commons.Date
	Final2Time    commons.Hour
	Final2Room    string
	Final2RevDate commons.Date
	Final2RevTime commons.Hour

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

	if strings.Contains(s.RawSubjectName, "(**)") {
		s.CourseType = academic.Laboratory
	} else if strings.Contains(s.RawSubjectName, "(*)") {
		s.CourseType = academic.ExamOnly
	}
}

func (s *SubjectDTO) SetSemester(val string) { s.Semester = commons.ConvertStringToNumber(val) }
func (s *SubjectDTO) SetLevel(val string)    { s.Level = commons.ConvertStringToNumber(val) }

func (s *SubjectDTO) SetSection(val string) {
	s.Section = strings.TrimSpace(strings.ToUpper(val))
}

func (s *SubjectDTO) SetShift(val string) {
	s.Shift = strings.TrimSpace(strings.ToUpper(val))
}

func (s *SubjectDTO) SetTeachersFirtNames(v string) {
	count := commons.ScanLines(v, func(i int, l string) { s.Teachers[i].FirstName = l })
	if count > s.TeacherCount {
		s.TeacherCount = count
	}
}
func (s *SubjectDTO) SetTeachersLastNames(v string) {
	count := commons.ScanLines(v, func(i int, l string) { s.Teachers[i].LastName = l })
	if count > s.TeacherCount {
		s.TeacherCount = count
	}
}
func (s *SubjectDTO) SetTeachersTitles(v string) {
	count := commons.ScanLines(v, func(i int, l string) { s.Teachers[i].Title = l })
	if count > s.TeacherCount {
		s.TeacherCount = count
	}
}
func (s *SubjectDTO) SetTeachersEmails(v string) {
	count := commons.ScanLines(v, func(i int, l string) { s.Teachers[i].Email = l })
	if count > s.TeacherCount {
		s.TeacherCount = count
	}
}

func (s *SubjectDTO) SetPartial1Date(val string)  { s.Partial1Date = commons.ParseDate(val) }
func (s *SubjectDTO) SetPartial1Time(val string)  { s.Partial1Time = commons.ParseTime(val) }
func (s *SubjectDTO) SetPartial1Room(val string)  { s.Partial1Room = val }
func (s *SubjectDTO) SetPartial2Date(val string)  { s.Partial2Date = commons.ParseDate(val) }
func (s *SubjectDTO) SetPartial2Time(val string)  { s.Partial2Time = commons.ParseTime(val) }
func (s *SubjectDTO) SetPartial2Room(val string)  { s.Partial2Room = val }
func (s *SubjectDTO) SetFinal1Date(val string)    { s.Final1Date = commons.ParseDate(val) }
func (s *SubjectDTO) SetFinal1Time(val string)    { s.Final1Time = commons.ParseTime(val) }
func (s *SubjectDTO) SetFinal1Room(val string)    { s.Final1Room = val }
func (s *SubjectDTO) SetFinal1RevDate(val string) { s.Final1RevDate = commons.ParseDate(val) }
func (s *SubjectDTO) SetFinal1RevTime(val string) { s.Final1RevTime = commons.ParseTime(val) }
func (s *SubjectDTO) SetFinal2Date(val string)    { s.Final2Date = commons.ParseDate(val) }
func (s *SubjectDTO) SetFinal2Time(val string)    { s.Final2Time = commons.ParseTime(val) }
func (s *SubjectDTO) SetFinal2Room(val string)    { s.Final2Room = val }
func (s *SubjectDTO) SetFinal2RevDate(val string) { s.Final2RevDate = commons.ParseDate(val) }
func (s *SubjectDTO) SetFinal2RevTime(val string) { s.Final2RevTime = commons.ParseTime(val) }

func (s *SubjectDTO) SetDayTime(day academic.WeekDay, val string) {
	s.Schedule[day] = WeekDayData{Room: s.Schedule[day].Room, Time: commons.ParseTimeSlot(val)}
}
func (s *SubjectDTO) SetDayRoom(day academic.WeekDay, room string) {
	s.Schedule[day] = WeekDayData{Room: strings.TrimSpace(room), Time: s.Schedule[day].Time}
}
func (s *SubjectDTO) SetSaturdayDates(dates string)    { s.SaturdayDates = dates }
func (s *SubjectDTO) SetCommitteePresident(val string) { s.CommitteePresident = val }
func (s *SubjectDTO) SetCommitteeMember1(val string)   { s.CommitteeMember1 = val }
func (s *SubjectDTO) SetCommitteeMember2(val string)   { s.CommitteeMember2 = val }

func (d *SubjectDTO) Reset() {
	*d = SubjectDTO{}
}
