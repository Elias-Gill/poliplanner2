package router

import (
	"context"
	"html/template"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"

	"net/http"
)

const latest_selection_cookie = "latestScheduleSelection"

func NewDashboardRouter(service *service.ScheduleService) func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		// Dashboard layout
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			userID := extractUserID(r)
			schedules, err := service.FindUserSchedules(ctx, userID)
			if err != nil {
				logger.Error("error finding user schedules", "user", userID, "error", err)
				w.Header().Add("HX-Redirect", "/500")
				http.Redirect(w, r, "/500", 500)
				return
			}

			var latestSelection int64 = -1
			if len(schedules) != 0 {
				latestSelection = schedules[0].ID
			}

			if cookie, err := r.Cookie(latest_selection_cookie); err == nil {
				if parsedValue, err := strconv.ParseInt(cookie.Value, 10, 64); err == nil {
					latestSelection = parsedValue
				}
			}

			data := struct {
				Schedules          []*model.Schedule
				SelectedScheduleID int64
			}{
				Schedules:          schedules,
				SelectedScheduleID: latestSelection,
			}

			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/dashboard/index.html"))
			tpl.Execute(w, data)
		})

		r.Get("/view", func(w http.ResponseWriter, r *http.Request) {
			rawId := r.URL.Query().Get("id")
			mode := r.URL.Query().Get("mode")
			if rawId == "" || mode == "" {
				logger.Debug("empty id or mode")
				w.Header().Add("HX-Redirect", "/404")
				return
			}

			id, err := strconv.ParseInt(rawId, 10, 64)
			if err != nil {
				logger.Error("error finding schedule subjects", "schedule", id, "error", err)
				w.Header().Add("HX-Redirect", "/404")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			subjects, err := service.FindScheduleDetail(ctx, id)
			if err != nil {
				logger.Error("error finding schedule subjects", "schedule", id, "error", err)
				w.Header().Add("HX-Redirect", "/500")
				http.Redirect(w, r, "/500", 500)
				return
			}
			if subjects == nil {
				w.Header().Add("HX-Redirect", "/404")
				return
			}

			data := struct {
				SelectedScheduleID int64
				Subjects           []*model.Subject
			}{
				SelectedScheduleID: id,
				Subjects:           subjects,
			}

			var tpl *template.Template
			switch mode {
			case "calendar":
				tpl = template.Must(template.ParseFiles("web/templates/pages/dashboard/calendar.html"))
			case "extra":
				tpl = template.Must(template.ParseFiles("web/templates/pages/dashboard/extra.html"))
			default:
				tpl = template.Must(template.ParseFiles("web/templates/pages/dashboard/overview.html"))
			}

			err = tpl.Execute(w, data)
			if err != nil {
				logger.Debug("error rendering template", "error", err)
			}
		})
	}
}
