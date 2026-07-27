package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type CurriculumSaveParams struct {
	SubjectID  academic.SubjectID
	CareerID   academic.CareerID
	Curriculum academic.Curriculum
}

type CurriculumRepository interface {
	Upsert(ctx context.Context, c CurriculumSaveParams) (academic.CurriculumID, error)

	GetByCareerID(ctx context.Context, career academic.CareerID) ([]academic.CurriculumSubjectItem, error)
}
