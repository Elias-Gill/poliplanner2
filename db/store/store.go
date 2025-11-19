package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/db/model"
)

type UserStorer interface {
	Insert(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, userID int64) error
	GetByID(ctx context.Context, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type SheetVersionStorer interface {
	Insert(ctx context.Context, s *model.SheetVersion) error
	GetNewest(ctx context.Context) (*model.SheetVersion, error)
}

type CareerStorer interface {
	Insert(ctx context.Context, c *model.Career) error
	Delete(ctx context.Context, careerID int64) error
	GetByID(ctx context.Context, careerID int64) (*model.Career, error)
	GetBySheetVersion(ctx context.Context, versionID int64) ([]*model.Career, error)
}

type SubjectStorer interface {
	Insert(ctx context.Context, s *model.Career) error
	GetByID(ctx context.Context, subjectID int64) (*model.Subject, error)
	GetByCareerId(ctx context.Context, careerID int64) ([]*model.Subject, error)
}

type ScheduleStorer interface {
	Insert(ctx context.Context, userID int64, s *model.Schedule) (int64, error)
	Delete(ctx context.Context, scheduleID int64) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error)
	GetByID(ctx context.Context, scheduleID int64) (*model.Schedule, error)
}

type ScheduleDetailStorer interface {
	Insert(ctx context.Context, scheduleID int64, subjectID int64) error
	GetSubjectsByScheduleID(ctx context.Context, scheduleID int64) (error, []*model.Subject)
}
