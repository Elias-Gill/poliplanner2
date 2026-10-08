package academic

type YearSemester int8

const (
	FirstSemester  YearSemester = 1
	SecondSemester YearSemester = 2
)

func (s YearSemester) String() string {
	switch s {
	case FirstSemester:
		return "Primer periodo"
	case SecondSemester:
		return "Segundo periodo"
	default:
		return "Periodo desconocido"
	}
}

type PeriodID int64

type Period struct {
	ID       PeriodID
	Year     int
	Semester YearSemester
}
