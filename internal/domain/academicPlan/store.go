package academicPlan

import (
	"context"
)

type AcademicPlanStorer interface {
	GetOrCreateByCareer(ctx context.Context, career string) (AcademicPlanID, error)
	AddSubject(ctx context.Context, planID AcademicPlanID, sub subject)
}

type AcademicPlanReadStore interface {
	GetPlanByCareerID(ctx context.Context, career int64) (*AcademicPlan, error)
	ListCareers(ctx context.Context) ([]*CareerReadModel, error)
}
