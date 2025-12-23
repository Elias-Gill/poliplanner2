package router

import (
	"context"
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

type Data struct {
	Schedule    *model.Schedule
	NeedsUpdate bool
}

func NewDashboardRouter(
	scheduleService *service.ScheduleService,
	sheetVersionService *service.SheetVersionService,
) func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		// Dashboard layout
		type Data struct {
			Schedule    *model.Schedule
			NeedsUpdate bool
		}

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			userID := extractUserID(r)
			schedules, err := scheduleService.FindUserSchedules(ctx, userID)
			if err != nil {
				logger.Error("error finding user schedules", "user", userID, "error", err)
				customRedirect(w, r, "/500")
				return
			}

			// Extract last viewed schedule
			var latestSelection int64 = -1

			if cookie, err := r.Cookie(latest_selection_cookie); err == nil {
				if parsedValue, err := strconv.ParseInt(cookie.Value, 10, 64); err == nil {
					latestSelection = parsedValue
				}
			}

			if latestSelection == -1 && len(schedules) > 0 {
				latestSelection = schedules[0].ID
			}

			// FIX: error handling
			latestExcel, _ := sheetVersionService.FindLatestSheetVersion(r.Context())

			// Prepare slice of Data combining schedule and update flag
			var scheduleData []Data
			for _, sched := range schedules {
				needsUpdate := false
				if sched.SheetVersion < latestExcel.ID {
					needsUpdate = true
				}

				scheduleData = append(scheduleData, Data{
					Schedule:    sched,
					NeedsUpdate: needsUpdate,
				})
			}

			data := struct {
				Schedules          []Data
				SelectedScheduleID int64
			}{
				Schedules:          scheduleData,
				SelectedScheduleID: latestSelection,
			}

			execTemplateWithLayout(w, "web/templates/pages/dashboard/index.html", layouts, data)
		})

		r.Get("/view", func(w http.ResponseWriter, r *http.Request) {
			rawId := r.URL.Query().Get("id")
			mode := r.URL.Query().Get("mode")
			if rawId == "" || mode == "" {
				logger.Debug("empty id or mode")
				customRedirect(w, r, "/404")
				return
			}

			id, err := strconv.ParseInt(rawId, 10, 64)
			if err != nil {
				logger.Error("error finding schedule subjects", "schedule", id, "error", err)
				customRedirect(w, r, "/404")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			subjects, err := scheduleService.FindScheduleDetail(ctx, id)
			if err != nil {
				logger.Error("error finding schedule subjects", "schedule", id, "error", err)
				customRedirect(w, r, "/500")
				return
			}
			if subjects == nil {
				customRedirect(w, r, "/404")
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     latest_selection_cookie,
				Value:    strconv.FormatInt(id, 10),
				Path:     "/dashboard",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   60 * 60 * 24 * 30, // 30 days
			})

			data := struct {
				SelectedScheduleID int64
				Subjects           []*model.Subject
			}{
				SelectedScheduleID: id,
				Subjects:           subjects,
			}

			tpl := "web/templates/pages/dashboard/overview.html"
			switch mode {
			case "calendar":
				tpl = "web/templates/pages/dashboard/calendar.html"
			case "extra":
				tpl = "web/templates/pages/dashboard/extra.html"
			}

			execTemplate(w, tpl, data)
		})
	}
}
