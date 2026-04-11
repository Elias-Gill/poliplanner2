package excelimport

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
)

type ImportRepository interface {
	RunImport(ctx context.Context, fn func(ImportWriter) error) error

	// SaveAudit records a new audit entry for an Excel import, capturing metadata,
	// whether the import succeeded, and any errors encountered during parsing.
	SaveAudit(ctx context.Context, version *sheetVersion.SheetVersion) (sheetVersion.SheetVersionID, error)
}

type ImportWriter interface {
	// Creates a new period. If the period already exists, returns its ID
	EnsurePeriod(p period.Period) (period.PeriodID, error)

	// Saves our course bundle. First inserts the subjet, ensures that the academic plan is
	// created for the current career. Creates career entries, teachers and finally persists the
	// final course information.
	SaveCourseOffering(off Offering) error
}
