package store

import (
	"context"
	"errors"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type UserStorer interface {
	Insert(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, userID int64) error
	Update(ctx context.Context, userID int64, updateFn func(user *model.User) error) error

	GetByID(ctx context.Context, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByRecoveryToken(ctx context.Context, token string) (*model.User, error)
}

var ErrNoSheetVersion = errors.New("no sheet version found")

type SheetVersionStorer interface {
	GetNewest(ctx context.Context) (*model.SheetVersion, error)
	GetLastCheckedAt(ctx context.Context) (*time.Time, error)
	// BUG: depends if the server is correctly set
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

type CourseStorer interface {
	// Upsert opens a transaction and allows inserting multiple CourseModel records sequentially
	// via the insertFn callback. This keeps memory low during large excel imports by processing
	// and persisting one record at a time.
	//
	// The insertFn callback exposes a "persist" function as its argument. The caller must
	// use this function to persist each individual CourseModel inside the transaction.
	//
	// All inserts run atomically (everything commits or the whole operation rolls back).
	Upsert(ctx context.Context, insertFn func(persist func(model.CourseModel) error) error) error
	FindById(ctx context.Context, id int64) (*model.CourseModel, error)
	ListByCareerAndPeriod(ctx context.Context, careerID int64, periodID int64) ([]*model.CourseListItem, error)
}

type ScheduleStorer interface {
	Insert(ctx context.Context, s *model.ScheduleBasicData) (int64, error)
	Delete(ctx context.Context, scheduleID int64) error

	ListByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error)
	GetByID(ctx context.Context, scheduleID int64) (*model.ScheduleDetails, error)
}

type CareerStorer interface {
	List(ctx context.Context) ([]*model.Career, error)
}

type PeriodStore interface {
	FindByYearPeriod(ctx context.Context, year int, period int) (*model.Period, error)
}
