package router

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

type ScheduleCreatePageData struct {
	Error        string
	Careers      []model.Career
	SheetVersion *model.SheetVersion
}

func NewSchedulesRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/create", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			version, err := service.FindLatestSheetVersion(ctx)
			if err != nil {
				logger.GetLogger().Error("error enpoint /schedule/create", "error", err)
				tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/500.html"))
				tpl.Execute(w, nil)
				return
			}

			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/schedule/index.html"))
			tpl.Execute(w, ScheduleCreatePageData{
				Error:        "",
				Careers:      nil,
				SheetVersion: version,
			})
		})

		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/schedule/index.html"))
			tpl.Execute(w, nil)
		})
	}
}
