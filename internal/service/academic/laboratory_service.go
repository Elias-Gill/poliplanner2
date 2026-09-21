package academic

import (
	"context"

	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type LaboratoryService struct {
	courseRepository academic.CourseRepository
	periodService    *PeriodService
}

func (c *LaboratoryService) GetOfferings(ctx context.Context, curriculum academicModel.CurriculumID) ([]academicModel.LaboratoryEntryView, error) {
	return nil, nil
}

func (c *LaboratoryService) GetOfferingsByCourse(ctx context.Context, curriculum academicModel.CourseID) ([]academicModel.LaboratoryEntryView, error) {
	return nil, nil
}
