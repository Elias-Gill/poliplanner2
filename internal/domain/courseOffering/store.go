package courseOffering

import (
	"context"
)

type CourseStorer interface {
	FindById(ctx context.Context, id int64) (*CourseOffering, error)
}
