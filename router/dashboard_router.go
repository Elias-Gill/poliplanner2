package router

import (
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

const latestSelectionCookie = "latestScheduleSelection"

func NewDashboardRouter(
	scheduleService *service.ScheduleService,
) func(r chi.Router) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "dashboard")

	// Main page
	base := parseTemplateWithBaseLayout(path.Join(baseDir, "index.html"))

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			base.Execute(w, nil)
		})

		r.Get("/sections", func(w http.ResponseWriter, r *http.Request) {
			executeFragment(w, "dashboard/subjects_section", nil)
		})
	}
}
