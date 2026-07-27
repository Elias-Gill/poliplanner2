package academic

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type SubjectService struct {
	subjecRepo academicRepo.SubjectRepository
}

func NewSubjectService(repo academicRepo.SubjectRepository) *SubjectService {
	return &SubjectService{
		subjecRepo: repo,
	}
}

func (s *SubjectService) GetSubject(ctx context.Context, id int64) (*academic.Subject, error) {
	return s.subjecRepo.GetByID(ctx, academic.SubjectID(id))
}
