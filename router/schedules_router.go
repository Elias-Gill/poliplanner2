package router

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/logger"
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
	step3 := parseTemplateWithBaseLayout(path.Join(basePath, "step3.html"))
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

			// basic validations
			title := r.Form.Get("title")
			careerIDStr := r.Form.Get("career_id")

			if title == "" || careerIDStr == "" {
				executeFragment(w, r, "messages/error_message", "missing required fields")
				return
			}

			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil || careerID <= 0 {
				executeFragment(w, r, "messages/error_message", "invalid career id")
				return
			}

			q := url.Values{}
			q.Add("career_id", careerIDStr)
			q.Add("title", title)
			customRedirect(w, r, "/schedule/step2?"+q.Encode())
		})

		// ================= STEP 2 =================
		// Subjects from academic plan selection
		//
		// GET:
		//   - Requires career context from Step 1.
		//   - Loads curricula (mallas) and subjects for the selected career.
		//   - Displays selectable subject list.
		//   - Returns fragment or full page depending on HTMX.
		r.Get("/step2", func(w http.ResponseWriter, r *http.Request) {
			careerIDStr := r.URL.Query().Get("career_id")
			title := r.URL.Query().Get("title")

			if title == "" {
				http.Error(w, "title is required", http.StatusBadRequest)
				return
			}

			// FIX: convertir para usar htmx para la validacion y mostrar errores y demas
			if careerIDStr == "" {
				http.Error(w, "career_id is required", http.StatusBadRequest)
				return
			}

			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid career_id", http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			// REFACTOR: differentiate "plan form career not exists" and DB errors
			plan, err := planService.GetByCareerID(ctx, careerID)
			if err != nil {
				customRedirect(w, r, "/500")
				return
			}

			// REFACTOR: Return a "bad form" page instead of 404
			if plan == nil {
				customRedirect(w, r, "/404")
				return
			}

			step2.Execute(w, map[string]any{
				"Plan":  plan,
				"Title": title,
			})
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
			// REFACTOR: mostrar pagina de bad form
			if err := r.ParseForm(); err != nil {
				http.Error(w, "invalid form", http.StatusBadRequest)
				return
			}

			title := r.Form.Get("title")
			assignmentIDs := r.Form["assignment_ids"]

			if title == "" {
				executeFragment(w, r, "messages/error_message", "missing title")
				return
			}

			// FIX: mostrar mensaje de error con htmx
			if len(assignmentIDs) == 0 {
				executeFragment(w, r, "messages/error_message", "select at least one assignment")
				return
			}

			for _, idStr := range assignmentIDs {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil || id <= 0 {
					executeFragment(w, r, "messages/error_message", "invalid assignment id")
					return
				}
			}

			// Add query params
			q := url.Values{}
			q.Set("title", title)

			for _, id := range assignmentIDs {
				q.Add("assignment_ids", id)
			}

			customRedirect(w, r, "/schedule/step3?"+q.Encode())
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
			assignmentIDs := r.URL.Query()["assignment_ids"] // slice de strings
			title := r.URL.Query().Get("title")

			if title == "" {
				executeFragment(w, r, "messages/error_message", "missing title")
				return
			}

			// FIX: convertir para usar htmx para la validacion y mostrar errores y demas
			if len(assignmentIDs) == 0 {
				executeFragment(w, r, "messages/error_message", "select at least one assignment")
				return
			}

			// Validate ids
			for _, idStr := range assignmentIDs {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil || id <= 0 {
					executeFragment(w, r, "messages/error_message", "invalid assignment id")
					return
				}
			}

			// TODO: get the data
			// ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			// defer cancel()

			// Datos de prueba para la plantilla
			fakeSections := []model.SectionsList{
				{
					Assignment: "Cálculo I",
					Sections: []model.Section{
						{ID: 1, Section: "TQ", Name: "Cálculo I", Professor: "Dr. Pérez", Type: 0},
						{ID: 2, Section: "MI", Name: "Cálculo I", Professor: "Dra. Sánchez", Type: 0},
					},
				},
				{
					Assignment: "Física I",
					Sections: []model.Section{
						{ID: 3, Section: "TR", Name: "Física I", Professor: "Ing. Gómez", Type: 0},
					},
				},
				{
					Assignment: "Química Final",
					Sections: []model.Section{
						{ID: 4, Section: "TR", Name: "Química Final (*)", Professor: "Dra. López", Type: 1},
					},
				},
			}

			// Ejecutar la plantilla
			step3.Execute(w, map[string]any{
				"Sections": fakeSections,
				"Title":    "Horario de prueba",
			})
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
