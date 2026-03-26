package courseOffering

import (
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
)

type CourseOfferingService struct {
	courseStorer courseOffering.CourseStorer
}

func NewAcademicPlanService(
	courseStorer courseOffering.CourseStorer,
) *CourseOfferingService {
	return &CourseOfferingService{courseStorer: courseStorer}
}
