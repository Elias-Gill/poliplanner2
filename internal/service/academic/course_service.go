package academic

import (
	"context"
	"fmt"

	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type CourseService struct {
	courseRepository academic.CourseRepository
	periodService    *PeriodService
}

func NewCourseService(
	courseRepo academic.CourseRepository,
	periodService *PeriodService,
) *CourseService {
	return &CourseService{
		periodService:    periodService,
		courseRepository: courseRepo,
	}
}

func (c *CourseService) GetOfferings(ctx context.Context, curriculum academicModel.CurriculumID) ([]academicModel.CourseSummaryView, error) {
	period, err := c.periodService.CalculateCurrentPeriod(ctx)
	if err != nil {
		return nil, err
	}

	courses, err := c.courseRepository.ListByCurriculumID(ctx, curriculum, academicModel.PeriodID(period))
	if err != nil {
		return nil, fmt.Errorf("get courses: %w", err)
	}

	for i := range courses {
		// Obtener profesores del curso
		teachers, err := c.courseRepository.GetCourseTeachers(ctx, courses[i].ID)
		if err != nil {
			return nil, fmt.Errorf("get teachers for course %v: %w", courses[i].ID, err)
		}

		// Obtener horarios del curso
		schedules, err := c.courseRepository.GetCourseSchedules(ctx, courses[i].ID)
		if err != nil {
			return nil, fmt.Errorf("get schedules for course %v: %w", courses[i].ID, err)
		}

		// Asignar los datos recopilados
		courses[i].Teachers = teachers
		courses[i].Schedules = schedules
	}

	return courses, nil
}
