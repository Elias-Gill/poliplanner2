package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

type CourseOfferingService struct {
	periodStore period.PeriodStore
	courseStore courseOffering.CourseStorer
}

func NewCourseOfferingService(
	courseStore courseOffering.CourseStorer,
	periodStore period.PeriodStore,
) *CourseOfferingService {
	return &CourseOfferingService{
		periodStore: periodStore,
		courseStore: courseStore,
	}
}

// ListOfferingsBySubjectIDs returns courses for the given career in the current period,
// grouped by semester and subject.
// PERFORMANCE: review because this is full AI code and even the sql querys may be
// reviewed for this use cases. The actual code is just SHIT
// FIX: agregar el el period a la primera pantalla de creacion de horarios.
func (s *CourseOfferingService) ListOfferingsForSubjects(
	ctx context.Context,
	subjects academicPlan.SubjectID,
	period period.PeriodID,
) ([]courseOffering.SemesterGroup, error) {
	items, err := s.courseStore.ListByIDsAndPeriod(ctx, careerID, period.ID)
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

func (s *CourseOfferingService) GetCurrentPeriod(ctx context.Context) (*period.Period, error) {
	year := time.Now().Year()
	semNum := calculateCurrentSemester()
	return s.periodStore.FindByYearSemester(ctx, year, semNum)
}
