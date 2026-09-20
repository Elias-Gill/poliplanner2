package academic

import (
	"context"
	"fmt"

	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

// maxTeacherHistoryPeriods limits how many of the most recent academic periods are
// included in a teacher's history, mirroring the course offering history tool.
const maxTeacherHistoryPeriods = 5

type TeacherService struct {
	teacherRepository academicRepo.TeacherRepository
	courseRepository  academicRepo.CourseRepository
	periodService     *PeriodService
	courseService     *CourseService
}

func NewTeacherService(
	teacherRepo academicRepo.TeacherRepository,
	courseRepo academicRepo.CourseRepository,
	periodService *PeriodService,
	courseService *CourseService,
) *TeacherService {
	return &TeacherService{
		teacherRepository: teacherRepo,
		courseRepository:  courseRepo,
		periodService:     periodService,
		courseService:     courseService,
	}
}

// ListTeachers returns every registered teacher.
func (s *TeacherService) ListTeachers(ctx context.Context) ([]academicModel.Teacher, error) {
	teachers, err := s.teacherRepository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	return teachers, nil
}

// GetTeacherHistory builds the full profile of a teacher together with the courses
// they taught across the most recent academic periods. It returns (nil, nil) when
// the teacher does not exist.
func (s *TeacherService) GetTeacherHistory(ctx context.Context, teacherID academicModel.TeacherID) ([]academicModel.TeacherHistoricEntry, error) {
	// List the current available periods
	periods, err := s.periodService.ListPeriods(ctx)
	if err != nil {
		return nil, fmt.Errorf("list periods: %w", err)
	}

	if len(periods) > maxTeacherHistoryPeriods {
		periods = periods[:maxTeacherHistoryPeriods]
	}

	result := make([]academicModel.TeacherHistoricEntry, 0)

	// List all teacher courses by period
	for _, p := range periods {
		courseIDs, err := s.courseRepository.ListByTeacherAndPeriod(ctx, teacherID, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get courses for teacher %v in period %v: %w", teacherID, p.ID, err)
		}

		offering := make([]academicModel.CourseSummaryView, 0, len(courseIDs))
		for _, courseID := range courseIDs {
			summary, err := s.courseService.GetCourseSummary(ctx, courseID)
			if err != nil {
				return nil, err
			}
			offering = append(offering, *summary)
		}

		// Skip periods where the teacher had no activity.
		if len(offering) == 0 {
			continue
		}

		result = append(result, academicModel.TeacherHistoricEntry{
			Period:   p,
			Offering: offering,
		})
	}

	return result, nil
}
