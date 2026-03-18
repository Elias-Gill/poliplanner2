package academicPlan

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
)

type AcademicPlanService struct {
	planStorer   academicPlan.AcademicPlanStorer
	courseStorer courseOffering.CourseStorer
}

func (a AcademicPlanService) ListCareers(ctx context.Context) (*academicPlan.CareerReadModel, error) {
	// TODO: immplementar
	return nil, nil
}

func (a AcademicPlanService) GetCareerPlan(
	ctx context.Context,
	career academicPlan.CareerID,
) (*academicPlan.AcademicPlan, error) {
	return nil, nil
}

func (a AcademicPlanService) ListOffering(
	ctx context.Context,
	courses []courseOffering.CourseOfferingID,
) ([]courseOffering.SectionsList, error) {

	// TODO: implementar
	return []courseOffering.SectionsList{
		{
			Assignment: "Cálculo I",
			Sections: []courseOffering.Section{
				{ID: 1, Section: "TQ", Name: "Cálculo I", Professor: "Dr. Pérez", Type: 0},
				{ID: 2, Section: "MI", Name: "Cálculo I", Professor: "Dra. Sánchez", Type: 0},
			},
		},
		{
			Assignment: "Física I",
			Sections: []courseOffering.Section{
				{ID: 3, Section: "TR", Name: "Física I", Professor: "Ing. Gómez", Type: 0},
			},
		},
		{
			Assignment: "Química Final",
			Sections: []courseOffering.Section{
				{ID: 4, Section: "TR", Name: "Química Final (*)", Professor: "Dra. López", Type: 1},
			},
		},
	}, nil
}
