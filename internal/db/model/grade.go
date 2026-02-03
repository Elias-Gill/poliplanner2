package model

import "time"

type Teacher struct {
	Name  string
	Email string
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

type GradeModel struct {
	Name       string
	Period     Period
	Teachers   []Teacher
	Curriculum Curriculum
	Section    string

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
type GradeListItem struct {
	ID          int64
	SubjectName string
	Section     string
	Semester    int

	TeacherTitle    string
	TeacherName     string
	TeacherLastname string
}
