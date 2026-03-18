package courseOffering

import (
	"context"
)

type CourseStorer interface {
	FindById(ctx context.Context, id int64) (*CourseOffering, error)
	ListByCareerAndPeriod(ctx context.Context, careerID int64, periodID int64) ([]*CourseListItem, error)
}
