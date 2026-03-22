package academicPlan

import (
	"context"

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
	// TODO: immplementar
	return []*academicPlan.Career{
		{ID: 780, Name: "IIN"},
		{ID: 790, Name: "LCIK"},
	}, nil
}

func (a AcademicPlanService) GetCareerPlan(
	ctx context.Context,
	career academicPlan.CareerID,
) (*academicPlan.AcademicPlan, error) {
	// TODO: immplementar
	return &academicPlan.AcademicPlan{
		CareerID: career,
		Semesters: []academicPlan.SemesterSubjects{
			{
				Semester: 1,
				Subjects: []academicPlan.Subject{
					{ID: 5, Department: "DCB", Name: "Calculo II", Level: 2},
					{ID: 6, Department: "DCB", Name: "Fisica II", Level: 2},
					{ID: 7, Department: "DCB", Name: "Programacion II", Level: 2},
					{ID: 8, Department: "DCB", Name: "Estadistica", Level: 2},
				},
			},
			{
				Semester: 2,
				Subjects: []academicPlan.Subject{
					{Department: "DCB", ID: 9, Name: "Estructura de Datos", Level: 3},
					{Department: "DCB", ID: 10, Name: "Base de Datos", Level: 3},
					{Department: "DCB", ID: 11, Name: "Matematica Discreta", Level: 3},
				},
			},
			{
				Semester: 3,
				Subjects: []academicPlan.Subject{
					{Department: "DCB", ID: 12, Name: "Sistemas Operativos", Level: 4},
					{Department: "DCB", ID: 13, Name: "Redes", Level: 4},
					{Department: "DCB", ID: 14, Name: "Ingenieria de Software", Level: 4},
				},
			},
			{
				Semester: 4,
				Subjects: []academicPlan.Subject{
					{Department: "DCB", ID: 15, Name: "Compiladores", Level: 5},
					{Department: "DCB", ID: 16, Name: "Inteligencia Artificial", Level: 5},
					{Department: "DCB", ID: 17, Name: "Electiva I", Level: 5},
				},
			},
		},
	}, nil
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
