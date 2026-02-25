package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type CourseService struct {
	courseStore store.CourseStorer
	periodStore store.PeriodStore
}

type SubjectGroup struct {
	SubjectName string
	Courses     []*model.CourseListItem
}

type SemesterGroup struct {
	Semester int
	Subjects []SubjectGroup
}

func NewCourseService(courseStore store.CourseStorer, periodStore store.PeriodStore) *CourseService {
	return &CourseService{
		courseStore: courseStore,
		periodStore: periodStore,
	}
}

func (s *CourseService) GetCurrentPeriod(ctx context.Context) (*model.Period, error) {
	year := time.Now().Year()
	periodNum := calculateCurrentPeriod()
	return s.periodStore.FindByYearPeriod(ctx, year, periodNum)
}

// ListActiveByCareer returns courses for the given career in the current period, grouped by semester and subject.
// PERFORMANCE: review because this is full AI code and even the sql querys may be
// reviewed for this use cases. The actual code is just SHIT
func (s *CourseService) ListActiveByCareer(
	ctx context.Context,
	careerID int64,
) ([]SemesterGroup, error) {

	period, err := s.GetCurrentPeriod(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period: %w", err)
	}

	items, err := s.courseStore.ListByCareerAndPeriod(ctx, careerID, period.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list courses: %w", err)
	}

	// semestre -> materia -> cursos
	bySemester := make(map[int]map[string][]*model.CourseListItem)

	for _, c := range items {
		if bySemester[c.Semester] == nil {
			bySemester[c.Semester] = make(map[string][]*model.CourseListItem)
		}
		bySemester[c.Semester][c.SubjectName] =
			append(bySemester[c.Semester][c.SubjectName], c)
	}

	var semesters []SemesterGroup

	for semester, subjectsMap := range bySemester {
		var subjects []SubjectGroup

		for subjectName, courses := range subjectsMap {
			sort.Slice(courses, func(i, j int) bool {
				return courses[i].Section < courses[j].Section
			})

			subjects = append(subjects, SubjectGroup{
				SubjectName: subjectName,
				Courses:     courses,
			})
		}

		sort.Slice(subjects, func(i, j int) bool {
			return subjects[i].SubjectName < subjects[j].SubjectName
		})

		semesters = append(semesters, SemesterGroup{
			Semester: semester,
			Subjects: subjects,
		})
	}

	sort.Slice(semesters, func(i, j int) bool {
		return semesters[i].Semester < semesters[j].Semester
	})

	return semesters, nil
}

func (s *CourseService) GetCourseDetail(ctx context.Context, id int64) (*model.CourseAggregate, error) {
	return s.courseStore.FindById(ctx, id)
}
