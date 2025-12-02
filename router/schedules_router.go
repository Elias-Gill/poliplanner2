package router

import (
	"context"
	"fmt"
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

type ScheduleCreatePageData struct {
	Error        string
	Careers      []*model.Career
	SheetVersion *model.SheetVersion
}

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

			careers, err := service.FindCareersBySheetVersion(ctx, latestExcel.VersionID)
			if err != nil {
				logger.GetLogger().Error("Error finding careers", "error", err)
				http.Redirect(w, r, "/500", 500)
			}

			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/schedule/index.html"))
			tpl.Execute(w, &ScheduleCreatePageData{
				Careers:      careers,
				SheetVersion: latestExcel,
			})
		})

		r.Get("/create/details", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*400)
			defer cancel()

			rawId := r.URL.Query().Get("careerId")
			// TODO: que hacer
			// if careerID == "" {
			// 	http.Error(w, "careerId is required", http.StatusBadRequest)
			// 	return
			// }

			careerId, err := strconv.ParseInt(rawId, 10, 64)
			// TODO: agregar mensaje de error para el id invalido
			// if err != nil {
			// 	logger.GetLogger().Error("Error finding subjects", "error", err, "careerID", rawId)
			// 	http.Redirect(w, r, "/500", 500)
			// 	return
			// }

			subjects, err := service.FindSubjectsByCareerID(ctx, careerId)
			if err != nil {
				logger.GetLogger().Error("Error finding subjects", "error", err, "careerID", rawId)
				http.Redirect(w, r, "/500", 500)
				return
			}

			// Preparar datos para el template
			data := struct{ Subjects []*model.Subject }{
				Subjects: subjects,
			}

			tpl := template.Must(template.ParseFiles("web/templates/pages/schedule/details.html"))
			if err := tpl.Execute(w, data); err != nil {
				logger.GetLogger().Error("Error executing template", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		})

		// TODO: continuar
		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()

			fmt.Println("---- NUEVO HORARIO ----")
			fmt.Println("description =", r.Form.Get("description"))
			fmt.Println("careerId    =", r.Form.Get("careerId"))
			fmt.Println("subjectIds  =", r.Form["subjectIds"]) // slice

			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusNoContent)
		})
	}
}
