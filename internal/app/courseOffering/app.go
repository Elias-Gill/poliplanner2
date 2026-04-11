package courseOffering

import (
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
)

type CourseOfferingService struct {
	courseStorer courseOffering.CourseRepository
}

func NewAcademicPlanService(
	courseStorer courseOffering.CourseRepository,
) *CourseOfferingService {
	return &CourseOfferingService{courseStorer: courseStorer}
}
