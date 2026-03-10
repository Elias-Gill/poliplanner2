package teacher

import (
	"errors"
	"strings"
)

type TeacherID int64

type Teacher struct {
	Name      string
	Email     string
	searchKey string
}

func (t *Teacher) IsSimilar(other string) bool {
	a := normalize(t.Name)
	b := normalize(other)

	// Exact match after normalization
	if a == b {
		return true
	}

	// Only use Levenshtein distance if both strings have the same size
	if len(a) != len(b) {
		return false
	}

	if levenshtein(a, b) > 2 {
		return false
	}

	return true
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
