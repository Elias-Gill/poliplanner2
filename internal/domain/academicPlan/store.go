package academicPlan

import (
	"context"
)

type AcademicPlanStorer interface {
	GetOrCreateByCareer(ctx context.Context, career string) (PlanID, error)
	AddSubject(ctx context.Context, planID PlanID, sub Subject) (SubjectID, error)
}

type AcademicPlanReadStore interface {
	GetPlanByCareerID(ctx context.Context, career int64) (*AcademicPlan, error)
	ListCareers(ctx context.Context) ([]*CareerReadModel, error)
}
