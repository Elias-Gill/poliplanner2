package router

import (
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

const latestSelectionCookie = "latestScheduleSelection"

func NewDashboardRouter(
	scheduleService *service.ScheduleService,
) func(r chi.Router) {

	return func(r chi.Router) {
		// TODO: implementar
	}
}
