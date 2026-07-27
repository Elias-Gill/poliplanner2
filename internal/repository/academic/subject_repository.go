package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SubjectRepository interface {
	Upsert(ctx context.Context, c academic.Subject) (academic.SubjectID, error)

	GetByID(ctx context.Context, id academic.SubjectID) (*academic.Subject, error)
}
