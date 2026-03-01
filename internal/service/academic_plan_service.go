package service

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type AcademicPlanService struct {
	planStorer store.AcademicPlanStore
}

func NewAcademicPlanService(
	planStorer store.AcademicPlanStore,
) *AcademicPlanService {
	return &AcademicPlanService{
		planStorer: planStorer,
	}
}

// (*AcademicPlan, nil) = existe plan
//
// (nil, nil) = no existe plan para esa carrera
//
// (nil, err) = error en la DB
func (s *AcademicPlanService) GetByCareerID(ctx context.Context, careerID int64) (*model.AcademicPlan, error) {
	return s.planStorer.GetByCareerID(ctx, careerID)
}
