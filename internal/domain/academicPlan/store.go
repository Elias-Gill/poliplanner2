package academicPlan

import (
	"context"
)

type AcademicPlanStorer interface {
	GetPlanByCareerID(ctx context.Context, career int64) (*AcademicPlan, error)
	ListCareers(ctx context.Context) ([]*Career, error)
}
