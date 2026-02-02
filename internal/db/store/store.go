package store

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type UserStorer interface {
	Insert(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, userID int64) error
	Update(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByRecoveryToken(ctx context.Context, token string) (*model.User, error)
}

type SheetVersionStorer interface {
	GetNewest(ctx context.Context) (*model.SheetVersion, error)
	GetLastCheckedAt(ctx context.Context) (*time.Time, error)
	SetLastCheckedAt(ctx context.Context, t time.Time) error

	// TODO: refactor to use a struct
	Save(
		ctx context.Context,
		fileName string,
		localPath string,
		url string,
		processedSheets int,
		succeededSheets int,
		errors []error,
	) (int64, error)
}

type GradeStorer interface {
	// Upsert opens a transaction and allows inserting multiple GradeModel records sequentially
	// via the insertFn callback. This keeps memory low during large excel imports by processing
	// and persisting one record at a time.
	//
	// The insertFn callback exposes a "persist" function as its argument. The caller must
	// use this function to persist each individual GradeModel inside the transaction.
	//
	// All inserts run atomically (everything commits or the whole operation rolls back).
	Upsert(ctx context.Context, insertFn func(persist func(model.GradeModel) error) error) error
	FindById(ctx context.Context, id int64) (*model.GradeModel, error)
}

type ScheduleStorer interface {
	Insert(ctx context.Context, s *model.Schedule) (int64, error)
	Delete(ctx context.Context, scheduleID int64, validateFn func(owner int64) error) error

	// TODO: This should return a basic schedule info
	GetByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error)

	// TODO: This should return a complete aggregate of grades for the given schedule
	GetByID(ctx context.Context, scheduleID int64, validateFn func(owner int64) error) (*model.ScheduleDetails, error)
}

type CareerStorer interface {
	List(ctx context.Context) ([]*model.Career, error)
}
