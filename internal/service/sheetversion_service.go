package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

func FindLatestSheetVersion(ctx context.Context) (*model.SheetVersion, error) {
	version, err := sheetVersionStorer.GetNewest(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("error searching latest schedule: %w", err)
	}

	return version, nil
}
