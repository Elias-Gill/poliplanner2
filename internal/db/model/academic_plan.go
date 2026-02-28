package model

type Assignment struct {
	ID         int64
	Semester   int
	Level      int
	Name       string
	Department string
}

type Semester struct {
	Number      int
	Assignments []Assignment
}

type AcademicPlan struct {
	ID        int64
	Semesters []Semester
}
