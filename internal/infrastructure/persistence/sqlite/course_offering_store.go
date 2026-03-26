package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
)

type SqliteCourseOfferingStore struct {
	db *sql.DB
}

func NewSqliteCourseOfferingStore(connection *sql.DB) *SqliteCourseOfferingStore {
	return &SqliteCourseOfferingStore{db: connection}
}

func (s SqliteCourseOfferingStore) FindById(ctx context.Context, id int64) (*courseOffering.CourseOffering, error) {
	return nil, nil
}
