package router

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewSchedulesRouter(
	courseService *service.CourseService,
	scheduleService *service.ScheduleService,
	careerService *service.CareerService,
	planService *service.AcademicPlanService,
) func(r chi.Router) {
	basePath := path.Join(
		config.Get().Paths.BaseDir,
		"web", "templates", "pages", "schedules")

	step1 := parseTemplateWithBaseLayout(path.Join(basePath, "step1.html"))
	step2 := parseTemplateWithBaseLayout(path.Join(basePath, "step2.html"))
	// step3 := parseTemplateWithBaseLayout(path.Join(basePath, "step3.html"))
	// step4 := parseTemplateWithBaseLayout(path.Join(basePath, "step4.html"))

	// GET endpoints are ment to just render the condition
	return func(r chi.Router) {
		// ================= STEP 1 =================
		// Career selection + schedule metadata (name and description)
		//
		// GET:
		//   - Renders the initial wizard page.
		//   - Loads available careers from the service layer.
		//   - Returns full page layout.
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
			defer cancel()

			careers, err := careerService.List(ctx)
			if err != nil {
				logger.Error("Error listing careers", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			err = step1.Execute(w, careers)

			if err != nil {
				logger.Error("Error executing template", "error", err)
			}
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "invalid form", http.StatusBadRequest)
				return
			}

			title := r.Form.Get("title")
			careerIDStr := r.Form.Get("career_id")

			if title == "" || careerIDStr == "" {
				// FIX: this http error things
				http.Error(w, "missing required fields", http.StatusBadRequest)
				return
			}

			// validar career_id numérico
			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil || careerID <= 0 {
				http.Error(w, "invalid career id", http.StatusBadRequest)
				return
			}

			customRedirect(w, r, "/schedule/step2?career_id="+careerIDStr+"&title="+url.QueryEscape(title))
		})

		// ================= STEP 2 =================
		// Subjects from academic plan selection
		//
		// GET:
		//   - Requires career context from Step 1.
		//   - Loads curricula (mallas) and subjects for the selected career.
		//   - Displays selectable subject list.
		//   - Returns fragment or full page depending on HTMX.
		// FIX: convertir para usar htmx para la validacion y mostrar errores y demas
		r.Get("/step2", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			careerIDStr := r.URL.Query().Get("career_id")
			// FIX: pasar title al siguiente endpoint
			// title := r.URL.Query().Get("title")

			if careerIDStr == "" {
				http.Error(w, "career_id is required", http.StatusBadRequest)
				return
			}

			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid career_id", http.StatusBadRequest)
				return
			}

			plan, err := planService.GetByCareerID(ctx, careerID)
			if err != nil {
				http.Error(w, "could not load academic plan", http.StatusInternalServerError)
				return
			}

			if plan == nil {
				http.Error(w, "academic plan not found", http.StatusNotFound)
				return
			}

			data, err := planService.GetByCareerID(ctx, careerID)
			if err != nil {
				// FIX: error handling
				return 
			}

			step2.Execute(w, data)
		})
		// POST:
		//   - Receives selected subject IDs.
		//   - Validates that subjects belong to the selected career.
		//   - Stores subject selection into wizard state.
		//   - On success:
		//       * HTMX = returns Step 3 fragment (sections selection).
		//       * Non-HTMX = redirect to Step 3 route.
		//   - On validation error:
		//       * Re-renders Step 2 with errors.
		r.Post("/step2", func(w http.ResponseWriter, r *http.Request) {

		})

		// ================= STEP 3 =================
		// Sections selection for previously chosen subjects
		//
		// GET:
		//   - Requires subject selection from Step 2.
		//   - Loads available course sections for the most recent academic period.
		//   - Groups sections by subject for easier UX.
		//   - Returns fragment or full page.
		//
		r.Get("/step3", func(w http.ResponseWriter, r *http.Request) {
			// TODO: support this
			customRedirect(w, r, "/schedule")
		})
		// POST:
		//   - Receives selected section/course IDs.
		//   - Validates consistency with prior selections.
		//   - Stores final selection into wizard state.
		//   - On success:
		//       * HTMX = returns Step 4 fragment (review & confirmation).
		//       * Non-HTMX = redirect to Step 4 route.
		//   - On validation error:
		//       * Re-renders Step 3 with errors.
		r.Post("/step3", func(w http.ResponseWriter, r *http.Request) {
			// TODO: support this
			customRedirect(w, r, "/schedule")
		})

		// ================= STEP 4 =================
		// Final review and confirmation
		//
		// GET:
		//   - Aggregates all previously collected wizard data.
		//   - Displays:
		//       * Career
		//       * Schedule metadata
		//       * Selected subjects
		//       * Selected sections
		//   - Allows user to confirm before persistence.
		//
		r.Get("/step4", func(w http.ResponseWriter, r *http.Request) {

		})
		// POST:
		//   - Performs final validation and invariants check.
		//   - Calls application service to create the Schedule aggregate.
		//   - Clears temporary wizard state.
		//   - On success:
		//       * HTMX = returns success fragment or redirect trigger.
		//       * Non-HTMX = redirect to schedule detail page.
		//   - On failure:
		//       * Re-renders review step with error message.
		r.Post("/step4", func(w http.ResponseWriter, r *http.Request) {
			// TODO: support this
			customRedirect(w, r, "/schedule")
		})
	}
}
