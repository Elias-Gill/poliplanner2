package schedule

import "github.com/elias-gill/poliplanner2/internal/domain/schedule"

type ScheduleService struct {
	storer schedule.ScheduleStorer
}

func NewScheduleService(storer schedule.ScheduleStorer) *ScheduleService {
	return &ScheduleService{
		storer: storer,
	}
}
