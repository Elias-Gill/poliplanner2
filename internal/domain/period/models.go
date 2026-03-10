package period

type PeriodID int64

type Semester int64

const (
	FirstSemester  Semester = 1
	SecondSemester Semester = 2
)

type Period struct {
	ID       PeriodID
	Year     int
	Semester Semester
}
