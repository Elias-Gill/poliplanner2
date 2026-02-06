package router

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewCourseRouter(
	courseSvc *service.CourseService,
	careerSvc *service.CareerService,
) func(r chi.Router) {
	layout := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			careers, err := careerSvc.List(ctx)
			if err != nil {
				logger.Error("Failed to load careers", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			execTemplateWithLayout(w, "web/templates/pages/course/index.html", layout, careers)
		})

		r.Get("/{careerID}/list", func(w http.ResponseWriter, r *http.Request) {
			careerIDStr := chi.URLParam(r, "careerID")
			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid career", http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			courses, err := courseSvc.ListActiveByCareer(ctx, careerID)
			if err != nil {
				http.Error(w, "Courses not found", http.StatusNotFound)
				return
			}

			execTemplate(w, "web/templates/pages/course/list.html", courses)
		})

		r.Get("/{courseID}/detail", func(w http.ResponseWriter, r *http.Request) {
			courseIDStr := chi.URLParam(r, "courseID")
			courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid course", http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			detail, err := courseSvc.GetCourseDetail(ctx, courseID)
			if err != nil {
				http.Error(w, "Course not found", http.StatusNotFound)
				return
			}

			execTemplate(w, "web/templates/pages/course/modal.html", detail)
		})
	}
}
