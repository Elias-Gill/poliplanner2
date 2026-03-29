package academicPlan

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
)

type AcademicPlanService struct {
	planStorer   academicPlan.AcademicPlanStorer
	courseStorer courseOffering.CourseStorer
}

func NewAcademicPlanService(
	planStorer academicPlan.AcademicPlanStorer,
	courseStorer courseOffering.CourseStorer,
) *AcademicPlanService {
	return &AcademicPlanService{planStorer: planStorer, courseStorer: courseStorer}
}

func (a AcademicPlanService) ListCareers(ctx context.Context) ([]*academicPlan.Career, error) {
	return a.planStorer.ListCareers(ctx)
}

func (a AcademicPlanService) GetCareerPlan(
	ctx context.Context,
	career academicPlan.CareerID,
) (*academicPlan.AcademicPlan, error) {
	return a.planStorer.GetPlanByCareerID(ctx, career)
}

func (a AcademicPlanService) ListOffering(
	ctx context.Context,
	courses []academicPlan.SubjectID,
) ([]courseOffering.OfferList, error) {
	offers := make([]courseOffering.OfferList, 0, len(courses))

	for _, subjectID := range courses {
		// Get subject info
		subject, err := a.planStorer.GetSubject(ctx, subjectID)
		if err != nil {
			return nil, fmt.Errorf("get subject %d: %w", subjectID, err)
		}

		// List sections
		sections, err := a.courseStorer.FindOfferForSubject(ctx, subjectID)
		if err != nil {
			return nil, fmt.Errorf("find offering for subject %d: %w", subjectID, err)
		}

		offers = append(offers, courseOffering.OfferList{
			Subject: subject.Name,
			Offer:   sections,
		})
	}

	return offers, nil
}

func (a AcademicPlanService) ListCoursesSchedule(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) (*courseOffering.CoursesScheduleView, error) {
	if len(courses) == 0 {
		return &courseOffering.CoursesScheduleView{}, nil
	}

	schedules, err := a.courseStorer.GetCoursesSchedules(ctx, courses)
	if err != nil {
		return nil, fmt.Errorf("get courses schedules: %w", err)
	}

	view := &courseOffering.CoursesScheduleView{}

	for _, class := range schedules {
		switch class.Day {
		case courseOffering.Monday:
			view.Monday = append(view.Monday, class)

		case courseOffering.Tuesday:
			view.Tuesday = append(view.Tuesday, class)

		case courseOffering.Wednesday:
			view.Wednesday = append(view.Wednesday, class)

		case courseOffering.Thursday:
			view.Thursday = append(view.Thursday, class)

		case courseOffering.Friday:
			view.Friday = append(view.Friday, class)

		case courseOffering.Saturday:
			view.Saturday = append(view.Saturday, class)

		default:
			return nil, fmt.Errorf("invalid weekday in course class: %v", class.Day)
		}
	}

	return view, nil
}

func (a AcademicPlanService) ListExams(
	ctx context.Context,
	id courseOffering.CourseOfferingID,
) ([]courseOffering.ExamClass, error) {
	// FIX: error handling
	exams, _ := a.courseStorer.GetCoursesExams(ctx, []courseOffering.CourseOfferingID{id})
	return exams, nil
}

func (a AcademicPlanService) ListCoursesExams(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) (*courseOffering.ExamsScheduleView, error) {

	if len(courses) == 0 {
		return &courseOffering.ExamsScheduleView{}, nil
	}

	exams, err := a.courseStorer.GetCoursesExams(ctx, courses)
	if err != nil {
		return nil, fmt.Errorf("get courses exams: %w", err)
	}

	view := &courseOffering.ExamsScheduleView{}

	for _, exam := range exams {
		switch exam.Type {

		case courseOffering.ExamPartial:
			switch exam.Instance {
			case courseOffering.Instance1:
				view.Partial1 = append(view.Partial1, exam)

			case courseOffering.Instance2:
				view.Partial2 = append(view.Partial2, exam)

			default:
				return nil, fmt.Errorf(
					"invalid partial exam instance for course %s: %v",
					exam.CourseName,
					exam.Instance,
				)
			}

		case courseOffering.ExamFinal:
			switch exam.Instance {
			case courseOffering.Instance1:
				view.Final1 = append(view.Final1, exam)

			case courseOffering.Instance2:
				view.Final2 = append(view.Final2, exam)

			default:
				return nil, fmt.Errorf(
					"invalid final exam instance for course %s: %v",
					exam.CourseName,
					exam.Instance,
				)
			}

		default:
			return nil, fmt.Errorf(
				"invalid exam type for course %s: %v",
				exam.CourseName,
				exam.Type,
			)
		}
	}

	return view, nil
}

func (a AcademicPlanService) ListCoursesInfo(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) ([]courseOffering.CourseSummary, error) {

	if len(courses) == 0 {
		return []courseOffering.CourseSummary{}, nil
	}

	result := make([]courseOffering.CourseSummary, 0, len(courses))

	for _, courseID := range courses {
		details, err := a.courseStorer.GetCourseDetails(ctx, courseID)
		if err != nil {
			return nil, fmt.Errorf("get course details %d: %w", courseID, err)
		}

		if details == nil {
			return nil, fmt.Errorf("course details %d not found", courseID)
		}

		result = append(result, *details)
	}

	return result, nil
}
