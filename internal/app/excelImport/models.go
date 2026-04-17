package excelimport

import (
	"fmt"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

// ============================================================
// Core structs
// ============================================================

type subject struct {
	Career     string
	Name       string
	Department string
	Semester   int
	Level      int
}

type teacher struct {
	FirstName string
	LastName  string
	Email     string
	SearchKey string
}

func (t *teacher) IsSimilar(other string) bool {
	a := normalize(t.FullName())
	b := normalize(other)

	if a == b {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	return levenshtein(a, b) <= 2
}

func (t teacher) FullName() string {
	if t.FirstName == "" {
		return t.LastName
	}
	if t.LastName == "" {
		return t.FirstName
	}
	return t.FirstName + " " + t.LastName
}

func newTeacher(firstName, lastName, email string) teacher {
	return teacher{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		SearchKey: generateSearchKey(firstName, lastName),
	}
}

type Offering struct {
	CourseName string
	Period     period.PeriodID
	Teachers   []teacher
	Subject    subject
	Section    string
	CourseType courseOffering.CourseType

	// Normalized collections
	Exams    []Exam
	Schedule []ScheduleEntry

	// Raw extras
	SaturdayDates string

	// Committee
	CommitteeMember1   string
	CommitteeMember2   string
	CommitteePresident string
}

// ============================================================
// Value Objects (flat, DB-friendly)
// ============================================================

type ScheduleEntry struct {
	Day   courseOffering.WeekDay
	Start *time.Time
	End   *time.Time
	Room  string
}

type Exam struct {
	Type     courseOffering.ExamType
	Instance courseOffering.ExamInstance
	Date     *time.Time
	Room     string
	Revision *time.Time
}

// ============================================================
// Mutators (simple, no hidden logic)
// ============================================================

func (o *Offering) AddSchedule(
	day courseOffering.WeekDay,
	start *time.Time,
	end *time.Time,
	room string,
) {
	o.Schedule = append(o.Schedule, ScheduleEntry{
		Day:   day,
		Start: start,
		End:   end,
		Room:  room,
	})
}

// ============================================================
// Utils
// ============================================================

func normalize(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	name = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'ä', 'â':
			return 'a'
		case 'é', 'è', 'ë', 'ê':
			return 'e'
		case 'í', 'ì', 'ï', 'î':
			return 'i'
		case 'ó', 'ò', 'ö', 'ô':
			return 'o'
		case 'ú', 'ù', 'ü', 'û':
			return 'u'
		case 'ñ':
			return 'n'
		case '.', ',', '-', '_':
			return -1
		}
		return r
	}, name)

	return strings.TrimSpace(name)
}

func levenshtein(s1 string, s2 string) int {
	if len(s1) < len(s2) {
		return levenshtein(s2, s1)
	}

	if len(s2) == 0 {
		return len(s1)
	}

	prev := make([]int, len(s2)+1)
	for i := range prev {
		prev[i] = i
	}

	curr := make([]int, len(s2)+1)
	for i, r1 := range s1 {
		curr[0] = i + 1
		for j, r2 := range s2 {
			cost := 0
			if r1 != r2 {
				cost = 1
			}
			curr[j+1] = min(
				prev[j+1]+1,
				curr[j]+1,
				prev[j]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

func generateSearchKey(firstName, lastName string) string {
	first := ""
	if fields := strings.Fields(firstName); len(fields) > 0 {
		first = strings.ToLower(fields[0])
	} else {
		return ""
	}

	last := ""
	if fields := strings.Fields(lastName); len(fields) > 0 {
		last = strings.ToLower(fields[0])
	} else {
		return ""
	}

	return first + "_" + last
}

func (o Offering) String() string {
	var sb strings.Builder

	sb.WriteString("Offering{ ")

	sb.WriteString("CourseName=" + o.CourseName + " ")
	sb.WriteString("Period=" + fmt.Sprintf("%v", o.Period) + " ")
	sb.WriteString("Section=" + o.Section + " ")
	sb.WriteString("CourseType=" + fmt.Sprintf("%v", o.CourseType) + " ")

	sb.WriteString("Subject={")
	sb.WriteString("Career=" + o.Subject.Career + " ")
	sb.WriteString("Name=" + o.Subject.Name + " ")
	sb.WriteString("Department=" + o.Subject.Department + " ")
	sb.WriteString("Semester=" + fmt.Sprintf("%d", o.Subject.Semester) + " ")
	sb.WriteString("Level=" + fmt.Sprintf("%d", o.Subject.Level))
	sb.WriteString("} ")

	sb.WriteString("Teachers=[")

	for i, t := range o.Teachers {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(t.FirstName + " " + t.LastName + "<" + t.Email + ">")
	}

	sb.WriteString("] ")

	sb.WriteString("Schedule=[")

	for i, s := range o.Schedule {
		if i > 0 {
			sb.WriteString(", ")
		}

		start := ""
		end := ""

		if s.Start != nil {
			start = s.Start.String()
		}
		if s.End != nil {
			end = s.End.String()
		}

		sb.WriteString(fmt.Sprintf("day=%v,start=%s,end=%s,room=%s", s.Day, start, end, s.Room))
	}

	sb.WriteString("] ")

	sb.WriteString("}")

	return sb.String()
}
