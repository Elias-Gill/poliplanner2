package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

func FindSubjectsByCareerID(ctx context.Context, careerID int64) ([]*model.Subject, error) {
	subjects, err := subjectStorer.GetByCareerID(ctx, db, careerID)

	if err != nil {
		return nil, fmt.Errorf("error searching subjects: %w", err)
	}

	return subjects, err
}
