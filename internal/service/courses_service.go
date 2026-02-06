package service

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type CourseService struct {
	courseStore store.CourseStorer
	periodStore store.PeriodStore
}

func NewCourseService(courseStore store.CourseStorer, periodStore store.PeriodStore) *CourseService {
	return &CourseService{
		courseStore: courseStore,
		periodStore: periodStore,
	}
}

func (s *CourseService) GetCurrentPeriod(ctx context.Context) (*model.Period, error) {
	year := time.Now().Year()
	periodNum := calculateCurrentPeriod()
	return s.periodStore.FindByYearPeriod(ctx, year, periodNum)
}

func (s *CourseService) ListActiveByCareer(ctx context.Context, careerID int64) ([]*model.CourseListItem, error) {
	period, err := s.GetCurrentPeriod(ctx)
	if err != nil {
		return nil, err
	}
	return s.courseStore.ListByCareerAndPeriod(ctx, careerID, period.ID)
}

func (s *CourseService) GetCourseDetail(ctx context.Context, id int64) (*model.CourseModel, error) {
	return s.courseStore.FindById(ctx, id)
}
