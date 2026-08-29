package academic

type YearSemester int8

const (
	FirstSemester  YearSemester = 1
	SecondSemester YearSemester = 2
)

type PeriodID int64

type Period struct {
	ID       PeriodID
	Year     int
	Semester YearSemester
}
