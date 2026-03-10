package period

import (
	"context"
)

type PeriodStore interface {
	FindByYearSemester(ctx context.Context, year int, periodNum Semester) (*Period, error)
}
