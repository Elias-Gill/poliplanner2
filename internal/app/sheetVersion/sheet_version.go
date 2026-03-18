package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type SheetVersionService struct {
	db                 *sql.DB
	sheetVersionStorer store.SheetVersionStorer
}

func NewSheetVersionService(
	sheetVersionStorer store.SheetVersionStorer,
) *SheetVersionService {
	return &SheetVersionService{
		sheetVersionStorer: sheetVersionStorer,
	}
}

func (s *SheetVersionService) FindLatestSheetVersion(
	ctx context.Context,
) (*model.SheetVersion, error) {
	version, err := s.sheetVersionStorer.GetNewest(ctx)
	if err != nil {
		return nil, fmt.Errorf("error searching latest schedule: %w", err)
	}

	return version, nil
}
