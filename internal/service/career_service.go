package service

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type CareerService struct {
	carreerStorer store.CareerStorer
}

func NewCareerService(
	careerStorer store.CareerStorer,
) *CareerService {
	return &CareerService{
		carreerStorer: careerStorer,
	}
}

func (s *CareerService) List(ctx context.Context) ([]*model.Career, error) {
	return s.carreerStorer.List(ctx)
}
