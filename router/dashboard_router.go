package router

import (
	"context"
	"html/template"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"

	"net/http"
)

func NewDashboardRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		// Dashboard layout
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			schedules, err := service.FindUserSchedules(ctx, extractUserID(r))
			if err != nil {
				w.Header().Add("HX-Redirect", "/500")
				http.Redirect(w, r, "/500", 500)
				return
			}

			data := struct {
				Schedules          []*model.Schedule
				SelectedScheduleID int64
				Error              string
				Success            string
				Warning            string
				HasNewExcel        string
			}{
				Schedules: schedules,
				// TODO: traer el dato de una cookie
				SelectedScheduleID: 1,
			}

			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/dashboard/index.html"))
			tpl.Execute(w, data)
		})

		r.Get("/{id}/details", func(w http.ResponseWriter, r *http.Request) {
			rawId := chi.URLParam(r, "id")
			if rawId == "" {
				w.Header().Add("HX-Redirect", "/404")
				return
			}
			id, err := strconv.ParseInt(rawId, 10, 64)
			if err != nil {
				w.Header().Add("HX-Redirect", "/404")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			subjects, err := service.FindScheduleDetail(ctx, id)
			if err != nil {
				w.Header().Add("HX-Redirect", "/500")
				http.Redirect(w, r, "/500", 500)
				return
			}
			if subjects == nil {
				w.Header().Add("HX-Redirect", "/404")
				return
			}

			data := struct {
				Subjects []*model.Subject
			}{
				Subjects: subjects,
			}

			tpl := template.Must(template.ParseFiles("web/templates/pages/dashboard/details.html"))
			tpl.Execute(w, data)
		})
	}
}
