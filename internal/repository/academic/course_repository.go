package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
	// period "github.com/elias-gill/poliplanner2/internal/model/academic"
)

type CourseSaveParams struct {
	Name          string
	Type          academic.CourseType
	Section       string
	Shift         string
	Period        academic.PeriodID
	Curriculum    academic.CurriculumID
	SaturdayDates string
	Comitee       academic.Committee
}

type CourseRepository interface {
	// Upsert persists the base Course entity.
	// It only stores fields belonging to the course itself (course table).
	// It does not modify or manage related entities such as teachers, schedules, or exams.
	Upsert(ctx context.Context, course *CourseSaveParams) (academic.CourseID, error)

	// AssignTeachers replaces all existing teacher assignments for the given course.
	// After this operation, the course will be associated only with the provided teacher IDs.
	// This operation is destructive: previous assignments are removed.
	AssignTeachers(ctx context.Context, courseID academic.CourseID, teachers []academic.TeacherID) error

	// AssignSchedule replaces the full schedule for the given course.
	// Existing schedule entries are removed and replaced with the provided set.
	AssignSchedule(ctx context.Context, courseID academic.CourseID, schedule []academic.ClassSession) error

	// AssignExams replaces all exams associated with the given course.
	// After execution, only the provided exams will exist for the course.
	AssignExams(ctx context.Context, courseID academic.CourseID, exams []academic.Exam) error

	// -----------------------
	// -   READ OPERATIONS   -
	// -----------------------

	ListByCurriculumID(ctx context.Context, curriculum academic.CurriculumID, period academic.PeriodID) ([]academic.CourseSummaryView, error)

	GetCourseTeachers(ctx context.Context, courseID academic.CourseID) ([]academic.Teacher, error)

	GetCourseSchedules(ctx context.Context, courseID academic.CourseID) ([]academic.ClassSession, error)
}
