package service

import (
	"context"
	"database/sql"
	"fmt"

	sheetversion "github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
)

type SheetVersion struct {
	db                 *sql.DB
	sheetVersionStorer sheetversion.SheetVersionRepository
}

func New(
	sheetVersionStorer sheetversion.SheetVersionRepository,
) *SheetVersion {
	return &SheetVersion{
		sheetVersionStorer: sheetVersionStorer,
	}
}

func (s *SheetVersion) FindLatestSheetVersion(
	ctx context.Context,
) (*sheetversion.SheetVersion, error) {
	version, err := s.sheetVersionStorer.GetNewest(ctx)
	if err != nil {
		return nil, fmt.Errorf("error searching latest schedule: %w", err)
	}

	return version, nil
}

// Ideas para mas funciones:
// ListVersions (solo las que funcionaron el parseo)
// ListAudit (el historial de parseo y los fallos y demas)
