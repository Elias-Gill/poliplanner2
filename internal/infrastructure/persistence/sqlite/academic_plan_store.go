package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
)

type SqliteAcademicPlanStore struct {
	db *sql.DB
}

func NewSqliteAcademicPlanStorer(connection *sql.DB) *SqliteAcademicPlanStore {
	return &SqliteAcademicPlanStore{db: connection}
}

func (s SqliteAcademicPlanStore) ListCareers(ctx context.Context) ([]*academicPlan.Career, error) {
	// TODO: immplementar
	return []*academicPlan.Career{
		{ID: 780, Name: "IIN"},
		{ID: 790, Name: "LCIK"},
	}, nil
}

func (s SqliteAcademicPlanStore) GetSubject(ctx context.Context, id academicPlan.SubjectID) (*academicPlan.Subject, error) {
	// TODO: implementar
	return nil, nil
}

func (s SqliteAcademicPlanStore) GetPlanByCareerID(ctx context.Context, career academicPlan.CareerID) (*academicPlan.AcademicPlan, error) {
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
