package schedule

type ScheduleService struct {
	storer ScheduleStorer
}

func NewScheduleService(storer ScheduleStorer) *ScheduleService {
	return &ScheduleService{
		storer: storer,
	}
}
