package period

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
)

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

func NewPeriodFromTime(now time.Time) Period {
	return Period{
		Year:     now.Year(),
		Semester: calculateCurrentSemester(),
	}
}

// Determine on which academic period we currently are
func calculateCurrentSemester() Semester {
	now := time.Now().In(timezone.ParaguayTZ)
	if now.Month() > 6 {
		return FirstSemester // August -> December
	} else {
		return SecondSemester // January -> July
	}
}
