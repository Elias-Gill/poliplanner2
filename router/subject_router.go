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

func NewSubjectRouter(
	ss *service.SubjectService,
	sv *service.SheetVersionService,
	cs *service.CareerService,
) func(r chi.Router) {
	layout := web.BaseLayout

	return func(r chi.Router) {
		// Index page
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*300)
			defer cancel()

			version, err := sv.FindLatestSheetVersion(ctx)
			if err != nil {
				logger.Error("Error finding latest excel version", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			careers, err := cs.FindCareersBySheetVersion(ctx, version.ID)
			if err != nil {
				logger.Error("Error finding careers for sheet", "error", err, "sheet", version.ID)
				customRedirect(w, r, "/500")
				return
			}

			execTemplateWithLayout(w, "web/templates/pages/subject/index.html", layout, careers)
		})

		// HTMX: load subjects list for a career
		r.Get("/list/{careerID}", func(w http.ResponseWriter, r *http.Request) {
			rawCareerID := chi.URLParam(r, "careerID")

			careerID, err := strconv.ParseInt(rawCareerID, 10, 64)
			if err != nil {
				logger.Debug("/subjects cannot parse careerID", "error", err)
				customRedirect(w, r, "/404")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			subjects, err := ss.FindSubjectsByCareerID(ctx, careerID)
			if err != nil {
				logger.Debug("/subjects cannot find subjects", "error", err)
				customRedirect(w, r, "/404")
				return
			}

			execTemplate(w, "web/templates/pages/subject/filter.html", subjects)
		})

		// HTMX: subject detail modal
		r.Get("/info/{subjectID}", func(w http.ResponseWriter, r *http.Request) {
			rawSubjectID := chi.URLParam(r, "subjectID")

			subjectID, err := strconv.ParseInt(rawSubjectID, 10, 64)
			if err != nil {
				logger.Debug("/subjects cannot parse subjectID", "error", err)
				customRedirect(w, r, "/404")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			info, err := ss.FindByID(ctx, subjectID)
			if err != nil {
				logger.Debug("/subjects cannot find subject info", "error", err)
				customRedirect(w, r, "/404")
				return
			}

			execTemplate(w, "web/templates/pages/subject/modal.html", info)
		})
	}
}
