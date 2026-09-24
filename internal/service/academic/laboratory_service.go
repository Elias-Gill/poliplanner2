package academic

import (
	"context"
	"fmt"

	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/repository/academic"
)

// LaboratoryService exposes read access to the laboratory offerings attached to a
// curriculum within an academic period.
type LaboratoryService struct {
	laboratoryRepository academic.LaboratoryRepository
	courseRepository     academic.CourseRepository
}

func NewLaboratoryService(
	laboratoryRepository academic.LaboratoryRepository,
	courseRepository academic.CourseRepository,
) *LaboratoryService {
	return &LaboratoryService{
		laboratoryRepository: laboratoryRepository,
		courseRepository:     courseRepository,
	}
}

// GetOfferings returns every laboratory offering of a curriculum during the given period.
func (c *LaboratoryService) GetOfferings(
	ctx context.Context,
	curriculum academicModel.CurriculumID,
	period academicModel.PeriodID,
) ([]academicModel.Laboratory, error) {
	labs, err := c.laboratoryRepository.ListByCurriculum(ctx, curriculum, period)
	if err != nil {
		return nil, fmt.Errorf("get laboratory offerings: %w", err)
	}
	return labs, nil
}

// GetOfferingsByCourse resolves the curriculum and period of a course and returns its
// laboratory offerings.
func (c *LaboratoryService) GetOfferingsByCourse(
	ctx context.Context,
	courseID academicModel.CourseID,
) ([]academicModel.Laboratory, error) {
	basic, err := c.courseRepository.GetBasicData(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("get course basic data: %w", err)
	}
	return c.GetOfferings(ctx, basic.Curriculum, basic.Period)
}
