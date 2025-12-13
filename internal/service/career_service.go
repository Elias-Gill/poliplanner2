package service

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type CareerService struct {
	db           *sql.DB
	careerStorer store.CareerStorer
}

func NewCareerService(
	db *sql.DB,
	careerStorer store.CareerStorer,
) *CareerService {
	return &CareerService{
		db:           db,
		careerStorer: careerStorer,
	}
}

func (s *CareerService) FindCareersBySheetVersion(
	ctx context.Context,
	versionID int64,
) ([]*model.Career, error) {
	return s.careerStorer.GetBySheetVersion(ctx, s.db, versionID)
}
