package service

import "time"

func calculateCurrentPeriod() int {
	// Determine on which period we currently are
	period := 2 // January -> July
	now := time.Now()
	if now.Month() > 6 {
		period = 1 // August -> December
	}

	return period
}
