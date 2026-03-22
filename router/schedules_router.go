package router

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/elias-gill/poliplanner2/internal/config"
	planModel "github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

type Step1View struct {
	Careers []*planModel.Career
	Title   string
	Career  string
	Error   string
}

type Step2View struct {
	Title    string
	CareerID int64
	Plan     *planModel.AcademicPlan
	Error    string
}

type Step3View struct {
	Title         string
	AssignmentIDs []string
	Sections      []courseOffering.OfferList
	Error         string
}

func NewSchedulesRouter(
	scheduleService *schedule.ScheduleService,
	planService *academicPlan.AcademicPlanService,
) func(r chi.Router) {
	basePath := path.Join(
		config.Get().Paths.BaseDir,
		"web", "templates", "pages", "schedules")

	step1 := parseTemplateWithBaseLayout(path.Join(basePath, "step1.html"))
	step2 := parseTemplateWithBaseLayout(path.Join(basePath, "step2.html"))
	step3 := parseTemplateWithBaseLayout(path.Join(basePath, "step3.html"))

	// GET endpoints are ment to just render the condition
	return func(r chi.Router) {

		// ================= STEP 1 =================
		// Career selection + schedule description

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()

			careers, err := planService.ListCareers(ctx)
			if err != nil {
				logger.Error("Cannot list careers for /schedule/", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			step1.Execute(w, Step1View{
				Careers: careers,
			})
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				customRedirect(w, r, "/schedule")
				return
			}

			title := r.Form.Get("title")
			careerID := r.Form.Get("career_id")

			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()

			// Print error messages for step 1
			careers, err := planService.ListCareers(ctx)
			if err != nil {
				logger.Error("Cannot list careers for /schedule/", "error", err)
				customRedirect(w, r, "/500")
			}

			// FIX: que muestre cual es el field que falta
			if title == "" || careerID == "" {
				step1.Execute(w, Step1View{
					Careers: careers,
					Title:   title,
					Career:  careerID,
					Error:   "Missing required fields",
				})
				return
			}

			q := url.Values{}
			q.Set("title", title)
			q.Set("career_id", careerID)

			customRedirect(w, r, "/schedule/step2?"+q.Encode())
		})

		// ================= STEP 2 =================
		r.Get("/step2", func(w http.ResponseWriter, r *http.Request) {
			careerIDStr := r.URL.Query().Get("career_id")
			title := r.URL.Query().Get("title")

			if title == "" || careerIDStr == "" {
				customRedirect(w, r, "/bad_form")
				return
			}

			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil || careerID <= 0 {
				customRedirect(w, r, "/bad_form")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			plan, err := planService.GetCareerPlan(ctx, planModel.CareerID(careerID))
			if err != nil { // DB error
				customRedirect(w, r, "/500")
				return
			}

			// Plan or career does not exists
			// FIX: seria bueno que el usuario pueda saber cuando un plan para carrera esta mal o
			// directamente no existe la carrera
			if plan == nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			view := Step2View{
				Title:    title,
				CareerID: careerID,
				Plan:     plan,
			}

			if err := step2.Execute(w, view); err != nil {
				logger.Error("Error executing step2 template", "error", err)
			}
		})

		r.Post("/step2", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			// Retrieve the acculumated data
			title := r.Form.Get("title")
			careerIDStr := r.Form.Get("career_id")
			assignmentIDs := r.Form["assignment_ids"]

			if title == "" || careerIDStr == "" {
				customRedirect(w, r, "/bad_form")
				return
			}

			careerID, err := strconv.ParseInt(careerIDStr, 10, 64)
			if err != nil || careerID <= 0 {
				customRedirect(w, r, "/bad_form")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			// Show errors if exists
			if len(assignmentIDs) == 0 {
				// FIX: seria bueno que el usuario pueda saber cuando un plan para carrera esta mal o
				plan, err := planService.GetCareerPlan(ctx, planModel.CareerID(careerID))
				if err != nil {
					customRedirect(w, r, "/500")
					return
				}

				if plan == nil {
					customRedirect(w, r, "/404")
					return
				}
				view := Step2View{
					Title:    title,
					CareerID: careerID,
					Plan:     plan,
					Error:    "select at least one assignment",
				}

				step2.Execute(w, view)
				return
			}

			for _, idStr := range assignmentIDs {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil || id <= 0 {
					// FIX: no se estan mostrando en los bad form el motivo.
					// TODO: Agregar mas tarde el motivo del bad form
					customRedirect(w, r, "/bad_form")
				}
			}

			q := url.Values{}
			q.Set("title", title)

			for _, id := range assignmentIDs {
				q.Add("assignment_ids", id)
			}

			customRedirect(w, r, "/schedule/step3?"+q.Encode())
		})

		// ================= STEP 3 =================
		r.Get("/step3", func(w http.ResponseWriter, r *http.Request) {
			assignmentIDs := r.URL.Query()["assignment_ids"]
			title := r.URL.Query().Get("title")

			if title == "" || len(assignmentIDs) == 0 {
				customRedirect(w, r, "/bad_form")
				return
			}

			// Validate ids
			courses := make([]planModel.SubjectID, len(assignmentIDs))
			for i, idStr := range assignmentIDs {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil || id <= 0 {
					customRedirect(w, r, "/bad_form")
					return
				}
				courses[i] = planModel.SubjectID(id)
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			// FIX: error handling
			sections, err := planService.ListOffering(ctx, courses)
			if err != nil {
				customRedirect(w, r, "/500")
				return
			}

			view := Step3View{
				Title:         title,
				AssignmentIDs: assignmentIDs,
				Sections:      sections,
			}

			if err := step3.Execute(w, view); err != nil {
				logger.Error("Error executing step3 template", "error", err)
			}
		})

		// POST
		r.Post("/step3", func(w http.ResponseWriter, r *http.Request) {
			// TODO: support this
			customRedirect(w, r, "/dashboard")
		})
	}
}
