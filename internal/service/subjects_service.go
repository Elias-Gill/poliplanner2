package service

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type GradeService struct {
	gradeStore  store.GradeStorer
	periodStore store.PeriodStore
}

func NewSubjectService(gradeStore store.GradeStorer) *GradeService {
	return &GradeService{
		gradeStore: gradeStore,
	}
}

func (s *GradeService) FindByID(ctx context.Context, id int64) (*model.GradeModel, error) {
	return s.gradeStore.FindById(ctx, id)
}

// REFACTOR: rever porque esta feo
func (s *GradeService) LightListByCareerCurrent(ctx context.Context, careerID int64) ([]*model.GradeListItem, error) {
	// Determine on which period we currently are
	period := 2 // January -> July
	now := time.Now()
	if now.Month() > 6 {
		period = 1 // August -> December
	}

	p, err := s.periodStore.FindByYearPeriod(ctx, now.Year(), period)
	if err != nil {
		return nil, err
	}

	return s.gradeStore.ListByCareerAndPeriod(ctx, careerID, p.ID)
}
