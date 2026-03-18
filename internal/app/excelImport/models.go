package excelimport

// Package excelimport contains internal data structures used exclusively during
// the Excel import process.
//
// The types defined in this file are NOT part of the domain model. They do not
// represent domain aggregates, entities, or value objects. Instead, they are
// temporary persistence-oriented structures used to bridge the gap between the
// parsed Excel data and the database layer.
//
// Their purpose is to hold the raw or partially transformed information extracted
// from the Excel sheets in a shape that is convenient for batch persistence
// operations. These structures should be treated as import DTOs tailored for the
// storage process, not as representations of the business domain.
//
// In particular, they mirror the column layout of the Excel source and may contain
// denormalized or redundant fields that would not exist in the domain model.
// Any domain logic or invariants must remain implemented in the domain layer,
// not in these types.

import (
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

type subject struct {
	Career     string
	Name       string
	Department string
	Semester   int
	Level      int
}

type teacher struct {
	Name      string
	Email     string
	SearchKey string
}

func (t *teacher) IsSimilar(other string) bool {
	a := normalize(t.Name)
	b := normalize(other)

	// Exact match
	if a == b {
		return true
	}

	// Apply Levenshtein only with exact distance
	if len(a) != len(b) {
		return false
	}

	dist := levenshtein(a, b)
	return dist <= 2
}

type Offering struct {
	CourseName string
	Period     period.PeriodID
	Teachers   []teacher
	Subject    subject
	Section    string
	CourseType courseOffering.CourseType

	// Partial exams
	Partial1 examData
	Partial2 examData

	// Final exams
	Final1 examData
	Final2 examData

	// Weekly course schedule
	Schedule map[courseOffering.WeekDay]WeekDayData

	// Special saturday sessions (raw representation)
	SaturdayDates string

	// Exam committee
	CommitteeMember1   string
	CommitteeMember2   string
	CommitteePresident string
}

// ============================================================
// Value Objects
// ============================================================

type timeSlot struct {
	Start *time.Time
	End   *time.Time
}

type WeekDayData struct {
	Room string
	Time timeSlot
}

type examData struct {
	Date     *time.Time
	Revision *time.Time
	Room     string
}

func newExamData(date *time.Time, revDate *time.Time, room string) examData {
	return examData{
		Date:     date,
		Revision: revDate,
		Room:     room,
	}
}

func (c *Offering) AddSchedule(day courseOffering.WeekDay, start *time.Time, end *time.Time, room string) {
	c.Schedule[day] = WeekDayData{
		Room: room,
		Time: timeSlot{
			Start: start,
			End:   end,
		},
	}
}

// ==========================================================
// =                        UTILS                           =
// ==========================================================

func normalize(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

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
			// delete this char
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
				prev[j+1]+1,  // delete
				curr[j]+1,    // insert
				prev[j]+cost, // substitute
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}
