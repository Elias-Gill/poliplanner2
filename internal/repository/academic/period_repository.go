package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type PeriodRepository interface {
	Upsert(ctx context.Context, c academic.Period) (academic.PeriodID, error)
}
