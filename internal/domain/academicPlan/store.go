package academicPlan

import (
	"context"
)

type AcademicPlanStorer interface {
	GetPlanByCareerID(ctx context.Context, career CareerID) ([]Subject, error)
	ListCareers(ctx context.Context) ([]*Career, error)

	GetSubject(ctx context.Context, id SubjectID) (*Subject, error)
	GetCareer(ctx context.Context, id CareerID) (*Career, error)
}
