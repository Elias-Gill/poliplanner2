package courseOffering

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

type CourseRepository interface {
	FindOfferForSubject(ctx context.Context, subejctID academicPlan.SubjectID, period period.Period) ([]Section, error)

	GetCourseDetails(ctx context.Context, id CourseOfferingID) (*CourseSummary, error)

	GetCoursesSchedules(ctx context.Context, id []CourseOfferingID) ([]CourseClass, error)
	GetCoursesExams(ctx context.Context, id []CourseOfferingID) ([]ExamClass, error)
}
