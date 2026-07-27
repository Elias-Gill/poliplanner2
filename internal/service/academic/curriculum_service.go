package academic

import (
	"context"
	"fmt"
	"slices"

	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type CurriculumService struct {
	curriculumRepository academicRepo.CurriculumRepository
	careerRepository     academicRepo.CareerRepository
}

func NewCurriculumService(curriculumRepo academicRepo.CurriculumRepository, careerRepo academicRepo.CareerRepository) *CurriculumService {
	return &CurriculumService{
		curriculumRepository: curriculumRepo,
		careerRepository:     careerRepo,
	}
}

func (c *CurriculumService) GetCurriculum(ctx context.Context, career academicModel.CareerID) (*academicModel.CareerCurriculumView, error) {
	subjects, err := c.curriculumRepository.GetByCareerID(ctx, career)
	if err != nil {
		return nil, err
	}

	plans, err := c.careerRepository.ListPlans(ctx, career)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans for career %v: %w", career, err)
	}

	semesters, levels := extractUniqueSemesters(subjects)

	return &academicModel.CareerCurriculumView{
		Plans:     plans,
		Subjects:  subjects,
		Levels:    levels,
		Semesters: semesters,
	}, nil
}

func extractUniqueSemesters(subjects []academicModel.CurriculumSubjectItem) ([]int, []int) {
	if len(subjects) == 0 {
		return []int{}, []int{}
	}

	seenSem := make(map[int]struct{})
	seenLvl := make(map[int]struct{})
	for _, subject := range subjects {
		if subject.Semester >= 0 {
			seenSem[subject.Semester] = struct{}{}
		}
		if subject.Level >= 0 {
			seenLvl[subject.Level] = struct{}{}
		}
	}

	semesters := make([]int, 0, len(seenSem))
	for sem := range seenSem {
		semesters = append(semesters, sem)
	}

	levels := make([]int, 0, len(seenLvl))
	for lvl := range seenLvl {
		levels = append(levels, lvl)
	}

	slices.Sort(semesters)
	slices.Sort(levels)

	return semesters, levels
}
