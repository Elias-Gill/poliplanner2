package teacher

import "context"

type TeacherStorer interface {
	Save(ctx context.Context, teachers []Teacher) ([]TeacherID, error)
}
