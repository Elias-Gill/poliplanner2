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

	// FindCurriculumID resolves the curriculum id from the career code, the plan code
	// and the normalized subject name. Returns 0 when there is no match.
	FindBySubjectAndPlan(ctx context.Context, careerCode, planCode, subjectName string) (academic.CurriculumID, error)
}
