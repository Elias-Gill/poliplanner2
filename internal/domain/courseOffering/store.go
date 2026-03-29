package courseOffering

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
)

type CourseStorer interface {
	FindOfferForSubject(ctx context.Context, subejctID academicPlan.SubjectID) ([]Section, error)

	GetCourseDetails(ctx context.Context, id CourseOfferingID) (*CourseSummary, error)

	GetCoursesSchedules(ctx context.Context, id []CourseOfferingID) ([]CourseClass, error)
	GetCoursesExams(ctx context.Context, id []CourseOfferingID) ([]ExamClass, error)
}
