package academicPlan

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/teacher"
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
	// TODO: implementar
	var offers []courseOffering.OfferList
	for _, subjectID := range courses {
		offers = append(offers, courseOffering.OfferList{
			Subject: mockSubjectName(subjectID),
			Offer: []courseOffering.Section{
				{
					ID:         courseOffering.SectionID(subjectID*10 + 1),
					Section:    "A",
					CourseName: mockSubjectName(subjectID),
					Type:       courseOffering.Normal,
					Teachers: []teacher.Teacher{
						{ID: 1, FirstName: "Juan", LastName: "Gonzalez"},
					},
				},
				{
					ID:         courseOffering.SectionID(subjectID*10 + 2),
					Section:    "B",
					CourseName: mockSubjectName(subjectID),
					Type:       courseOffering.Normal,
					Teachers: []teacher.Teacher{
						{ID: 2, FirstName: "Maria", LastName: "Lopez"},
						{ID: 3, FirstName: "Carlos", LastName: "Fernandez"},
					},
				},
				{
					ID:         courseOffering.SectionID(subjectID*10 + 3),
					Section:    "C",
					CourseName: mockSubjectName(subjectID),
					Type:       courseOffering.ExamOnly,
					Teachers: []teacher.Teacher{
						{ID: 4, FirstName: "Pedro", LastName: "Martinez"},
					},
				},
			},
		})
	}

	return offers, nil
}

func mockSubjectName(id academicPlan.SubjectID) string {
	switch id {
	case 1:
		return "Calculo I"
	case 2:
		return "Algebra Lineal"
	case 3:
		return "Fisica I"
	case 4:
		return "Programacion I"
	case 5:
		return "Calculo II"
	case 6:
		return "Fisica II"
	case 7:
		return "Programacion II"
	case 8:
		return "Estadistica"
	case 9:
		return "Estructura de Datos"
	case 10:
		return "Base de Datos"
	default:
		return "Asignatura Desconocida"
	}
}

func (a AcademicPlanService) ListCoursesExams(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) (courseOffering.ExamsScheduleView, error) {

	now := time.Now()
	later := now.Add(48 * time.Hour)

	return courseOffering.ExamsScheduleView{
		Partial1: []courseOffering.ExamClass{
			{
				CourseName: "Matematica 1",
				Date:       now,
				Revision:   nil,
				Room:       "A52",
			},
			{
				CourseName: "Fisica 1",
				Date:       later,
				Revision:   nil,
				Room:       "B12",
			},
		},
		Partial2: []courseOffering.ExamClass{
			{
				CourseName: "Programacion 1",
				Date:       later,
				Revision:   nil,
				Room:       "Lab 2",
			},
		},
		Final1: []courseOffering.ExamClass{
			{
				CourseName: "Matematica 1",
				Date:       later,
				Revision:   &now,
				Room:       "A52",
			},
		},
		Final2: []courseOffering.ExamClass{
			{
				CourseName: "Programacion 1",
				Date:       now,
				Revision:   nil,
				Room:       "Lab 2",
			},
		},
	}, nil
}

func (a AcademicPlanService) ListCoursesSchedule(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) (*courseOffering.CoursesScheduleView, error) {
	// TODO: implementar
	now := time.Now()
	view := &courseOffering.CoursesScheduleView{
		Monday: []courseOffering.CourseClass{
			{
				CourseID: 1,
				Name:     "Matematica 1",
				Room:     "Aula 3",
				Start:    now.Add(2 * time.Hour),
				End:      now.Add(4 * time.Hour),
			},
			{
				CourseID: 2,
				Name:     "Fisica 1",
				Room:     "Laboratorio",
				Start:    now.Add(5 * time.Hour),
				End:      now.Add(7 * time.Hour),
			},
			{
				CourseID: 1,
				Name:     "Matematica 1",
				Room:     "Aula 3",
				Start:    now.Add(2 * time.Hour),
				End:      now.Add(4 * time.Hour),
			},
			{
				CourseID: 2,
				Name:     "Fisica 1",
				Room:     "Laboratorio",
				Start:    now.Add(5 * time.Hour),
				End:      now.Add(7 * time.Hour),
			},
		},

		Tuesday: []courseOffering.CourseClass{
			{
				CourseID: 3,
				Name:     "Programacion",
				Room:     "Aula 5",
				Start:    now.Add(3 * time.Hour),
				End:      now.Add(5 * time.Hour),
			},
		},

		Wednesday: []courseOffering.CourseClass{},
		Thursday:  []courseOffering.CourseClass{},
		Friday:    []courseOffering.CourseClass{},
		Saturday:  []courseOffering.CourseClass{},
	}

	return view, nil
}

func (a AcademicPlanService) ListCoursesInfo(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) ([]courseOffering.CourseSummary, error) {

	result := []courseOffering.CourseSummary{
		{
			Name: "Matematica 1",
			Teachers: []courseOffering.TeacherInfo{
				{
					Name:  "Juan Perez",
					Email: "juan.perez@politecnica.edu.py",
				},
				{
					Name:  "Maria Gomez",
					Email: "maria.gomez@politecnica.edu.py",
				},
			},
			Section:    "TQ",
			CourseType: courseOffering.Normal,

			SaturdayDates: "10/05, 24/05",

			CommitteeMember1:   "Carlos Ruiz",
			CommitteeMember2:   "Ana Duarte",
			CommitteePresident: "Luis Fernandez",
		},
		{
			Name: "Fisica 1",
			Teachers: []courseOffering.TeacherInfo{
				{
					Name:  "Pedro Martinez",
					Email: "pedro.martinez@politecnica.edu.py",
				},
			},
			Section:       "B",
			CourseType:    courseOffering.Normal,
			SaturdayDates: "",

			CommitteeMember1:   "Miguel Benitez",
			CommitteeMember2:   "Rosa Acosta",
			CommitteePresident: "Jose Caballero",
		},
		{
			Name: "Programacion 1",
			Teachers: []courseOffering.TeacherInfo{
				{
					Name:  "Laura Diaz",
					Email: "laura.diaz@politecnica.edu.py",
				},
			},
			Section:       "A",
			CourseType:    courseOffering.ExamOnly,
			SaturdayDates: "15/09",

			CommitteeMember1:   "Andres Lopez",
			CommitteeMember2:   "Sofia Rojas",
			CommitteePresident: "Diego Vargas",
		},
	}

	return result, nil
}
