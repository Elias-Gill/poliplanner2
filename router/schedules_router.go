package router

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewSchedulesRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/create", func(w http.ResponseWriter, r *http.Request) {
			// IMPORTANT: All the database operations should be done in no more than 400ms
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*400)
			defer cancel()

			latestExcel, err := service.FindLatestSheetVersion(ctx)
			if err != nil {
				logger.GetLogger().Error("Error finding latest excel version", "error", err)
				http.Redirect(w, r, "/500", 500)
			}

			careers, err := service.FindCareersBySheetVersion(ctx, latestExcel.ID)
			if err != nil {
				logger.GetLogger().Error("Error finding careers", "error", err)
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

			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/schedule/index.html"))
			tpl.Execute(w, data)
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

			subjects, err := service.FindSubjectsByCareerID(ctx, careerId)
			if err != nil {
				logger.GetLogger().Error("Error finding subjects", "error", err, "careerID", rawId)
				http.Redirect(w, r, "/500", 500)
				return
			}

			// Template data
			data := struct{ Subjects []*model.Subject }{
				Subjects: subjects,
			}

			tpl := template.Must(template.ParseFiles("web/templates/pages/schedule/details.html"))
			tpl.Execute(w, data)
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

			err = service.CreateSchedule(ctx, userID, sheetVersionID, description, subjectIDs)
			if err != nil {
				logger.GetLogger().Error("cannot create schedule", "error", err)
				w.Header().Set("HX-Redirect", "/500")
				return
			}

			w.Header().Set("HX-Redirect", "/dashboard")
		})
	}
}
