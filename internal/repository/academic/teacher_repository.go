package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type TeacherRepository interface {
	Upsert(ctx context.Context, c academic.Teacher) (academic.TeacherID, error)

	GetByID(ctx context.Context, id academic.TeacherID) (*academic.Teacher, error)
}
