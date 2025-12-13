package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type UserStorer interface {
	Insert(ctx context.Context, exec Executor, u *model.User) error
	Delete(ctx context.Context, exec Executor, userID int64) error
	GetByID(ctx context.Context, exec Executor, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, exec Executor, username string) (*model.User, error)
	GetByEmail(ctx context.Context, exec Executor, email string) (*model.User, error)
}

type SheetVersionStorer interface {
	Insert(ctx context.Context, exec Executor, s *model.SheetVersion) error
	GetNewest(ctx context.Context, exec Executor) (*model.SheetVersion, error)
	HasToUpdate(ctx context.Context, exec Executor) bool
}

type CareerStorer interface {
	Insert(ctx context.Context, exec Executor, c *model.Career) error
	Delete(ctx context.Context, exec Executor, careerID int64) error
	GetByID(ctx context.Context, exec Executor, careerID int64) (*model.Career, error)
	GetBySheetVersion(ctx context.Context, exec Executor, versionID int64) ([]*model.Career, error)
}

type SubjectStorer interface {
	Insert(ctx context.Context, exec Executor, careerID int64, s *model.Subject) error
	GetByID(ctx context.Context, exec Executor, subjectID int64) (*model.Subject, error)
	GetByCareerID(ctx context.Context, exec Executor, careerID int64) ([]*model.Subject, error)
}

type ScheduleStorer interface {
	Insert(ctx context.Context, exec Executor, s *model.Schedule) (int64, error)
	Delete(ctx context.Context, exec Executor, scheduleID int64) error
	GetByUserID(ctx context.Context, exec Executor, userID int64) ([]*model.Schedule, error)
	GetByID(ctx context.Context, exec Executor, scheduleID int64) (*model.Schedule, error)
}

type ScheduleDetailStorer interface {
	Insert(ctx context.Context, exec Executor, scheduleID int64, subjectID int64) error
	GetSubjectsByScheduleID(ctx context.Context, exec Executor, scheduleID int64) ([]*model.Subject, error)
}

// Executor abstracts over sql.DB and sql.Tx, allowing them to be used interchangeably.
// This design lets the service layer manage transaction boundaries explicitly,
// since it has the knowledge of which operations must be grouped within a transaction.
type Executor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
