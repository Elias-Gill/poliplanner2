package model

// Light weight subjects info, used to optimize database retrieve when listing a lot of
// subjects
type SubjectListItem struct {
	ID          int64
	SubjectName string
	Semester    int
	Section     string

	TeacherTitle    string
	TeacherName     string
	TeacherLastname string
}

