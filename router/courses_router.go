package router

import (
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewCourseRouter(
	courseSvc *service.CourseService,
	careerSvc *service.CareerService,
) func(r chi.Router) {
	return func(r chi.Router) {
		// TODO: complete
	}
}
