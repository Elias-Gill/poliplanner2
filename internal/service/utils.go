package service

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

// Determine on which academic period we currently are
func calculateCurrentSemester() period.Semester {
	now := time.Now()
	if now.Month() > 6 {
		return period.FirstSemester // August -> December
	} else {
		return period.SecondSemester // January -> July
	}
}
