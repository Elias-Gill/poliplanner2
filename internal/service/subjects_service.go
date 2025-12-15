package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type SubjectService struct {
	db           *sql.DB
	subjectStore store.SubjectStorer
}

func NewSubjectService(db *sql.DB, subjectStore store.SubjectStorer) *SubjectService {
	return &SubjectService{
		db:           db,
		subjectStore: subjectStore,
	}
}

func (s *SubjectService) FindByID(ctx context.Context, subjectID int64) (*model.Subject, error) {
	return s.subjectStore.GetByID(ctx, s.db, subjectID)
}

func (s *SubjectService) FindSubjectsByCareerID(ctx context.Context, careerID int64) ([]*store.SubjectListItem, error) {
	subjects, err := s.subjectStore.GetByCareerID(ctx, s.db, careerID)
	if err != nil {
		return nil, fmt.Errorf("error searching subjects: %w", err)
	}
	return subjects, nil
}
