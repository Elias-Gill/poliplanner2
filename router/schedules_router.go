package router

import (
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewSchedulesRouter(
	courseService *service.CourseService,
	scheduleService *service.ScheduleService,
	careerService *service.CareerService,
	sheetVersionService *service.SheetVersionService,
) func(r chi.Router) {
	return func(r chi.Router) {
		// TODO: implementar
	}
}
