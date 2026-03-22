package sqlite

import "database/sql"

type SqliteCourseOfferingStore struct {
	db *sql.DB
}

func NewSqliteCourseOfferingStore(connection *sql.DB) *SqliteCourseOfferingStore {
	return &SqliteCourseOfferingStore{db: connection}
}
