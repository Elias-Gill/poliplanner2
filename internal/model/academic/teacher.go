package academic

import "strings"

type TeacherID int64

type Teacher struct {
	ID        TeacherID
	Title     string
	FirstName string
	LastName  string
	Email     string
}

// FullName returns the teacher's display name including their title, trimming
// any extra whitespace when optional fields are empty.
func (t Teacher) FullName() string {
	return strings.TrimSpace(strings.Join([]string{t.Title, t.FirstName, t.LastName}, " "))
}
