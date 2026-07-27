package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type CareerRepository interface {
	Upsert(ctx context.Context, c academic.Career) (academic.CareerID, error)

	GetByID(ctx context.Context, id academic.CareerID) (*academic.Career, error)
	List(ctx context.Context) ([]*academic.Career, error)
	ListPlans(ctx context.Context, id academic.CareerID) ([]academic.Plan, error)
}
