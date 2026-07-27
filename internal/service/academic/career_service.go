package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type CareerService struct {
	careerRepository academicRepo.CareerRepository
}

func NewCareerService(planStorer academicRepo.CareerRepository) *CareerService {
	return &CareerService{careerRepository: planStorer}
}

func (a CareerService) ListCareers(ctx context.Context) ([]*academic.Career, error) {
	return a.careerRepository.List(ctx)
}
