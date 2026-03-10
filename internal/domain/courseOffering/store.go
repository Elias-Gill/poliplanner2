package courseOffering

import (
	"context"
)

type CourseStorer interface {
	// Upsert opens a transaction and allows inserting multiple CourseModel records sequentially
	// via the insertFn callback. This keeps memory low during large excel imports by processing
	// and persisting one record at a time.
	//
	// The insertFn callback exposes a "persist" function as its argument. The caller must
	// use this function to persist each individual CourseModel inside the transaction.
	//
	// All inserts run atomically (everything commits or the whole operation rolls back).
	// REFACTOR: veremos que hacer con esta funcion despues
	Upsert(ctx context.Context, insertFn func(persist func(*CourseOffering) error) error) error

	FindById(ctx context.Context, id int64) (*CourseOffering, error)
	ListByCareerAndPeriod(ctx context.Context, careerID int64, periodID int64) ([]*CourseListItem, error)
}
