package service

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

func FindCareersBySheetVersion(ctx context.Context, versionID int64) ([]*model.Career, error) {
	return careerStorer.GetBySheetVersion(ctx, db, versionID)
}
