package router

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewSchedulesRouter(
	subjectService *service.SubjectService,
	scheduleService *service.ScheduleService,
	sheetVersionService *service.SheetVersionService,
	careerService *service.CareerService,
) func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/create", func(w http.ResponseWriter, r *http.Request) {
			// IMPORTANT: All the database operations should be done in no more than 400ms
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*400)
			defer cancel()

			latestExcel, err := sheetVersionService.FindLatestSheetVersion(ctx)
			if err != nil {
				logger.Error("Error finding latest excel version", "error", err)
				http.Redirect(w, r, "/500", 500)
			}

			careers, err := careerService.FindCareersBySheetVersion(ctx, latestExcel.ID)
			if err != nil {
				logger.Error("Error finding careers", "error", err)
				http.Redirect(w, r, "/500", 500)
			}

			// Template data
			data := &struct {
				Error        string
				Careers      []*model.Career
				SheetVersion *model.SheetVersion
			}{
				Careers:      careers,
				SheetVersion: latestExcel,
			}

			execTemplateWithLayout(w, "web/templates/pages/schedule/index.html", layouts, data)
		})

		r.Get("/create/details", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*400)
			defer cancel()

			rawId := r.URL.Query().Get("careerId")
			if rawId == "" {
				http.Error(w, "CareerID is required", http.StatusBadRequest)
				return
			}

			careerId, err := strconv.ParseInt(rawId, 10, 64)
			if err != nil {
				http.Error(w, "Invalid careerID", http.StatusBadRequest)
				return
			}

			subjects, err := subjectService.FindSubjectsByCareerID(ctx, careerId)
			if err != nil {
				logger.Error("Error finding subjects", "error", err, "careerID", rawId)
				http.Redirect(w, r, "/500", 500)
				return
			}

			// Template data
			data := struct{ Subjects []*store.SubjectListItem }{
				Subjects: subjects,
			}

			execTemplate(w, "web/templates/pages/schedule/details.html", data)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")
			if idParam == "" {
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			scheduleID, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			userID := extractUserID(r)

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			err = scheduleService.DeleteSchedule(ctx, userID, scheduleID)
			if err != nil {
				logger.Error("cannot delete schedule", "error", err)
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			// Just redirect to the dashboard so we dont have to handle anything else here
			w.Header().Set("HX-Redirect", "/dashboard")
		})

		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()

			description := r.Form.Get("description")
			rawSheetVersionID := r.Form.Get("sheetVersionID")
			rawSubjectIDs := r.Form["subjectIds"]

			if description == "" {
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			sheetVersionID, err := strconv.ParseInt(rawSheetVersionID, 10, 64)
			if err != nil {
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			subjectIDs := make([]int64, 0, len(rawSubjectIDs))
			for _, sID := range rawSubjectIDs {
				id, err := strconv.ParseInt(sID, 10, 64)
				if err != nil {
					w.Header().Set("HX-Redirect", "/500")
					return
				}
				subjectIDs = append(subjectIDs, id)
			}

			// extract user from session
			userID := extractUserID(r)

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			err = scheduleService.CreateSchedule(ctx, userID, sheetVersionID, description, subjectIDs)
			if err != nil {
				logger.Error("cannot create schedule", "error", err)
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			w.Header().Set("HX-Redirect", "/dashboard")
		})

		r.Get("/{scheduleID}/update", func(w http.ResponseWriter, r *http.Request) {
			userID := extractUserID(r)

			rawScheduleID := chi.URLParam(r, "scheduleID")
			scheduleID, err := strconv.ParseInt(rawScheduleID, 10, 64)
			if err != nil {
				customRedirect(w, r, "/404")
				return
			}

			err = scheduleService.MigrateSchedule(r.Context(), userID, scheduleID)
			if err != nil {
				// Log the error with context for debugging
				logger.Error("Failed to migrate schedule", "user", userID, "scheduleID", scheduleID, "error", err)
				customRedirect(w, r, "/500")
				return
			}

			// redirect to dashboard
			customRedirect(w, r, "/")
		})
	}
}
