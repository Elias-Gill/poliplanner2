package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
)

type SqliteCourseOfferingStore struct {
	db *sql.DB
}

func NewSqliteCourseOfferingStore(connection *sql.DB) *SqliteCourseOfferingStore {
	return &SqliteCourseOfferingStore{db: connection}
}

func (s SqliteCourseOfferingStore) FindById(ctx context.Context, id int64) (*courseOffering.CourseOffering, error) {
	// TODO: implement
	return nil, nil
}

func (s SqliteCourseOfferingStore) FindOfferForSubject(ctx context.Context, subejctID academicPlan.SubjectID) ([]courseOffering.Section, error) {
	// TODO: implement
	return nil, nil
}

func (s SqliteCourseOfferingStore) GetCourseDetails(ctx context.Context, id courseOffering.CourseOfferingID) (*courseOffering.CourseSummary, error) {
	// TODO: implement
	return nil, nil
}

func (s SqliteCourseOfferingStore) GetCoursesSchedules(ctx context.Context, id []courseOffering.CourseOfferingID) ([]courseOffering.CourseClass, error) {
	// TODO: implement
	return nil, nil
}

func (s SqliteCourseOfferingStore) GetCoursesExams(ctx context.Context, id []courseOffering.CourseOfferingID) ([]courseOffering.ExamClass, error) {
	// TODO: implement
	return nil, nil
}
