package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/db/models"
)

type UserStorer interface {
	Insert(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, userID int64) error
	GetByID(ctx context.Context, userID int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type SheetVersionStorer interface {
	Insert(ctx context.Context, s *models.SheetVersion) error
	GetNewest(ctx context.Context) (*models.SheetVersion, error)
}

type CareerStorer interface {
	Insert(ctx context.Context, c *models.Career) error
	Delete(ctx context.Context, careerID int64) error
	GetByID(ctx context.Context, careerID int64) (*models.Career, error)
	GetBySheetVersion(ctx context.Context, versionID int64) ([]*models.Career, error)
}

type SubjectStorer interface {
	Insert(ctx context.Context, s *models.Career) error
	GetByID(ctx context.Context, subjectID int64) (*models.Subject, error)
	GetByCareerId(ctx context.Context, careerID int64) ([]*models.Subject, error)
}

type ScheduleStorer interface {
	Insert(ctx context.Context, userID int64, s *models.Schedule) error
	Delete(ctx context.Context, scheduleID int64) error
	GetByUserID(ctx context.Context, userID int64) (error, []*models.Schedule)
}

type ScheduleDetailStorer interface {
	Insert(ctx context.Context, scheduleID int64, subjectID int64) error
	GetSubjectsByScheduleID(ctx context.Context, scheduleID int64) (error, []*models.Subject)
}
