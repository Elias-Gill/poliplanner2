package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type TeacherRepository interface {
	Upsert(ctx context.Context, c academic.Teacher) (academic.TeacherID, error)

	GetByID(ctx context.Context, id academic.TeacherID) (*academic.Teacher, error)

	// List returns every registered teacher ordered by last name and first name.
	List(ctx context.Context) ([]academic.Teacher, error)
}
