package schedules

import (
	"github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/go-chi/chi/v5"
)

func NewSchedulesRouter(
	scheduleService *schedule.Schedule,
	planService *academicPlan.AcademicPlan,
) func(r chi.Router) {

	handlers := newSchedulesHandlers(scheduleService, planService)

	return func(r chi.Router) {
		// STEP 1
		r.Get("/", handlers.Step1Get)
		r.Post("/", handlers.Step1Post)
		r.Post("/delete", handlers.DeleteSchedule)

		// STEP 2
		r.Get("/step2", handlers.Step2Get)
		r.Post("/step2", handlers.Step2Post)

		// STEP 3
		r.Get("/step3", handlers.Step3Get)
		r.Post("/step3", handlers.Step3Post)
	}
}
