package model

import (
	"errors"
	"strings"
	"time"
)

type Teacher struct {
	Name      string
	Email     string
	searchKey string
}

func (t *Teacher) IsSimilar(other string) bool {
	a := normalize(t.Name)
	b := normalize(other)

	// Comparación exacta después de normalizar
	if a == b {
		return true
	}

	// Solo Levenshtein si tienen la misma longitud (evita comparaciones absurdas)
	if len(a) != len(b) {
		return false
	}

	// Distancia de Levenshtein ≤ 2
	dist := levenshtein(a, b)
	return dist <= 2
}

func (t *Teacher) GenerateSearchKey(firstName, lastName string) error {
	first := ""
	if fields := strings.Fields(firstName); len(fields) > 0 {
		first = strings.ToLower(fields[0])
	} else {
		return errors.New("first name is empty")
	}

	last := ""
	if fields := strings.Fields(lastName); len(fields) > 0 {
		last = strings.ToLower(fields[0])
	} else {
		return errors.New("last name is empty")
	}

	t.searchKey = strings.Join([]string{first, last}, "_")
	return nil
}

func (t Teacher) GetSearchKey() string {
	return t.searchKey
}

type Subject struct {
	Name       string
	Department string
}

type Curriculum struct {
	Semester int
	Subject  Subject
	// acronym of the career, should be all capitalized
	Career string
}

type TimeSlot struct {
	Start string
	End   string
}

type Period struct {
	ID     int64 // REFACTOR: no me gusta que este aca, mejor tener structs separados para cada cosa
	Year   int
	Period int
}

type CourseModel struct {
	ID         int64
	Name       string
	Period     Period
	Teachers   []Teacher
	Curriculum Curriculum
	Section    string
	CourseType  int

	// First partial
	Partial1Date *time.Time
	Partial1Time string
	Partial1Room string

	// Second partial
	Partial2Date *time.Time
	Partial2Time string
	Partial2Room string

	// First final
	Final1Date    *time.Time
	Final1Time    string
	Final1Room    string
	Final1RevDate *time.Time
	Final1RevTime string

	// Second final
	Final2Date    *time.Time
	Final2Time    string
	Final2Room    string
	Final2RevDate *time.Time
	Final2RevTime string

	// Weekly schedule
	MondayRoom string
	Monday     TimeSlot

	TuesdayRoom string
	Tuesday     TimeSlot

	WednesdayRoom string
	Wednesday     TimeSlot

	ThursdayRoom string
	Thursday     TimeSlot

	FridayRoom string
	Friday     TimeSlot

	SaturdayRoom  string
	Saturday      TimeSlot
	SaturdayDates string

	CommitteeMember1   string
	CommitteeMember2   string
	CommitteePresident string
}

// Light weight grade info, used to optimize database and network usage when listing a lot of
// grades
type CourseListItem struct {
	ID          int64
	SubjectName string
	Section     string
	Semester    int
	Teachers    string
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
