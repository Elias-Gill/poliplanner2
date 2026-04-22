package dashboard

import (
	"github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/go-chi/chi/v5"
)

func NewDashboardRouter(
	scheduleService *schedule.Schedule,
	planService *academicPlan.AcademicPlan,
) func(r chi.Router) {

	handlers := newDashboardHandlers(scheduleService, planService)

	return func(r chi.Router) {
		r.Get("/", handlers.Dashboard)
	}
}
