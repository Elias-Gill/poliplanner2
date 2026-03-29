package academicPlan

import (
	"context"
)

type AcademicPlanStorer interface {
	GetPlanByCareerID(ctx context.Context, career CareerID) (*AcademicPlan, error)
	ListCareers(ctx context.Context) ([]*Career, error)

	GetSubject(ctx context.Context, id SubjectID) (*Subject, error)
}
