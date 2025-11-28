package router

import (
	"context"
	"html/template"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"

	// "html/template"
	"net/http"
)

type DashboardData struct {
	Schedules        []*model.Schedule
	SelectedSchedule *model.Schedule
	Error            string
	Success          string
	Warning          string
	HasNewExcel      string
}

func NewDashboardRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		// Dashboard layout
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			schedules, err := service.FindUserSchedules(ctx, extractUserID(r))
			if err != nil {
				// TODO: Print error page
				w.Write([]byte("<h1>Aca paso algo muy feo</h1>"))
				return
			}

			if len(schedules) == 0 {
				tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/dashboard/index.html"))
				tpl.Execute(w, DashboardData{
					Schedules: schedules,
				})
				return
			}

			// TODO: imprimir el dashboard normal
		})
	}
}
