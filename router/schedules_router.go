package router

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewSchedulesRouter(
	courseService *service.CourseService,
	scheduleService *service.ScheduleService,
	careerService *service.CareerService,
	sheetVersionService *service.SheetVersionService,
) func(r chi.Router) {
	// GET endpoints are ment to just render the condition 
	return func(r chi.Router) {
		// ================= STEP 1 =================
		// Career selection + schedule metadata (name and description)
		//
		// GET:
		//   - Renders the initial wizard page.
		//   - Loads available careers from the service layer.
		//   - Returns full page (layout) for normal navigation.
		//   - Returns only the form fragment if the request is HTMX.
		//
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		})
		// POST:
		//   - Validates submitted career ID and basic schedule info.
		//   - Initializes temporary wizard state (session, hidden inputs, or server store).
		//   - On success:
		//       * HTMX = returns Step 2 fragment (curriculum/subjects selection).
		//       * Non-HTMX = redirect to Step 2 route.
		//   - On validation error:
		//       * Returns Step 1 fragment with error messages.
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		})

		// ================= STEP 2 =================
		// Subjects from academic plan selection
		//
		// GET:
		//   - Requires career context from Step 1.
		//   - Loads curricula (mallas) and subjects for the selected career.
		//   - Displays selectable subject list.
		//   - Returns fragment or full page depending on HTMX.
		//
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

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
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

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
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

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
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

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
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

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
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		})
	}
}
