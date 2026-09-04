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

	var result []academicModel.CourseSummaryView
	for _, id := range courses {
		summary, err := c.GetCourseSummary(ctx, id)
		if err != nil {
			return nil, err
		}

		result = append(result, *summary)
	}

	return result, nil
}

func (c *CourseService) GetCourseSummary(ctx context.Context, courseID academicModel.CourseID) (*academicModel.CourseSummaryView, error) {
	basicData, err := c.courseRepository.GetBasicData(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("get basic data for course %v: %w", courseID, err)
	}

	teachers, err := c.courseRepository.GetCourseTeachers(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("get teachers for course %v: %w", courseID, err)
	}

	schedules, err := c.courseRepository.GetCourseSchedules(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("get schedules for course %v: %w", courseID, err)
	}

	exams, err := c.courseRepository.GetCourseExams(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("get exams for course %v: %w", courseID, err)
	}

	var course academicModel.CourseSummaryView

	course.ID = basicData.ID
	course.Section = basicData.Section
	course.Shift = basicData.Shift
	course.Name = basicData.Name
	course.Type = basicData.Type
	course.SaturdayDates = basicData.SaturdayDates
	course.Committee = basicData.Committee

	course.Teachers = teachers
	course.Schedules = schedules
	course.Exams = exams

	return &course, nil
}

type historicResult struct {
	Period   academicModel.Period
	Offering []academicModel.CourseSummaryView
}

func (c *CourseService) GetHistoricOfferings(ctx context.Context, curriculum academicModel.CurriculumID) ([]historicResult, error) {
	periods, err := c.periodService.ListPeriods(ctx)
	if err != nil {
		return nil, fmt.Errorf("list periods: %w", err)
	}

	// Limit to the last 5 periods
	if len(periods) > 5 {
		periods = periods[:5]
	}

	var result []historicResult
	for _, p := range periods {
		courses, err := c.courseRepository.ListByCurriculumID(ctx, curriculum, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get courses: %w", err)
		}

		var offer []academicModel.CourseSummaryView
		for _, id := range courses {
			summary, err := c.GetCourseSummary(ctx, id)
			if err != nil {
				return nil, err
			}
			offer = append(offer, *summary)
		}

		result = append(result, historicResult{Period: p, Offering: offer})
	}

	return result, nil
}
