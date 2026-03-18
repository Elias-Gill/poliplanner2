package period

import (
	"context"
)

type PeriodStore interface {
	Save(ctx context.Context, p Period) (PeriodID, error)
}
