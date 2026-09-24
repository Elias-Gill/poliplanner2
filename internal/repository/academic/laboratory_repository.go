package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

// LaboratorySaveParams groups the data needed to persist a laboratory offering.
type LaboratorySaveParams struct {
	Curriculum academic.CurriculumID
	Period     academic.PeriodID
	Section    string
	Schedule   []academic.ClassSession
}

type LaboratoryRepository interface {
	// Replace upserts the laboratory offering and replaces its full schedule.
	Upsert(ctx context.Context, params LaboratorySaveParams) (academic.LaboratoryID, error)
}
