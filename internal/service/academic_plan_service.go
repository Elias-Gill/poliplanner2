package service

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
)

type AcademicPlanService struct {
	planStorer    academicPlan.AcademicPlanStorer
	planReadStore academicPlan.AcademicPlanReadStore
}

func NewAcademicPlanService(
	planStorer academicPlan.AcademicPlanStorer,
	planReadStore academicPlan.AcademicPlanReadStore,
) *AcademicPlanService {
	return &AcademicPlanService{
		planStorer:    planStorer,
		planReadStore: planReadStore,
	}
}

func (s *AcademicPlanService) GetPlanByCareerID(ctx context.Context, careerID int64) (*academicPlan.AcademicPlan, error) {
	return s.planReadStore.GetPlanByCareerID(ctx, careerID)
}

func (s *AcademicPlanService) ListCareers(ctx context.Context) ([]*academicPlan.CareerReadModel, error) {
	return s.planReadStore.ListCareers(ctx)
}
