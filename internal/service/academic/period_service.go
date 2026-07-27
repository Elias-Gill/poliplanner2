package academic

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/model/academic"

	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type PeriodService struct {
	periodRepo academicRepo.PeriodRepository
}

func NewPeriodService(periodRepo academicRepo.PeriodRepository) *PeriodService {
	return &PeriodService{
		periodRepo: periodRepo,
	}
}

func (p *PeriodService) CalculateCurrentPeriod(ctx context.Context) (academic.PeriodID, error) {
	period := p.NewPeriodFromTime(time.Now().In(timezone.ParaguayTZ))

	id, err := p.periodRepo.Upsert(ctx, period)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (p *PeriodService) NewPeriodFromTime(t time.Time) academic.Period {
	return academic.Period{
		Year:     t.Year(),
		Semester: calculateSemester(t),
	}
}

func calculateSemester(t time.Time) academic.YearSemester {
	if t.Month() > time.July || (t.Month() == time.July && t.Day() >= 23) {
		return academic.SecondSemester // 23 de July -> December
	}
	return academic.FirstSemester // January -> 22 de July
}
