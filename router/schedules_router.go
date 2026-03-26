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
	scheduleModel "github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
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

// ---------- router ----------

func NewSchedulesRouter(
	scheduleService *schedule.ScheduleService,
	planService *academicPlan.AcademicPlanService,
) func(r chi.Router) {

	basePath := path.Join(
		config.Get().Paths.BaseDir,
		"web", "templates", "pages", "schedules",
	)

	step1 := parseTemplateWithBaseLayout(path.Join(basePath, "step1.html"))
	step2 := parseTemplateWithBaseLayout(path.Join(basePath, "step2.html"))
	step3 := parseTemplateWithBaseLayout(path.Join(basePath, "step3.html"))

	return func(r chi.Router) {

		// ================= STEP 1 =================

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()

			careers, err := planService.ListCareers(ctx)
			if err != nil {
				logger.Error("cannot list careers", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			view := Step1View{
				Careers: careers,
			}

			if err := step1.Execute(w, view); err != nil {
				logger.Error("step1 template error", "error", err)
			}
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

			if err := r.ParseForm(); err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()

			careers, err := planService.ListCareers(ctx)
			if err != nil {
				logger.Error("cannot list careers", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			view := Step1View{
				Careers: careers,
				Title:   r.Form.Get("title"),
				Career:  r.Form.Get("career_id"),
			}

			title, err := requiredString(view.Title)
			if err != nil {
				view.Error = "title is required"
				step1.Execute(w, view)
				return
			}

			careerID, err := parseID(view.Career)
			if err != nil {
				view.Error = "invalid career selected"
				step1.Execute(w, view)
				return
			}

			q := url.Values{}
			q.Set("title", title)
			q.Set("career_id", strconv.FormatInt(careerID, 10))

			customRedirect(w, r, "/schedule/step2?"+q.Encode())
		})

		// ================= STEP 2 =================

		r.Get("/step2", func(w http.ResponseWriter, r *http.Request) {

			title, err := requiredString(r.URL.Query().Get("title"))
			if err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			careerID, err := parseID(r.URL.Query().Get("career_id"))
			if err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			plan, err := planService.GetCareerPlan(ctx, planModel.CareerID(careerID))
			if err != nil {
				logger.Error("cannot get career plan", "error", err)
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
			}

			if err := step2.Execute(w, view); err != nil {
				logger.Error("step2 template error", "error", err)
			}
		})

		r.Post("/step2", func(w http.ResponseWriter, r *http.Request) {

			if err := r.ParseForm(); err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			title, err := requiredString(r.Form.Get("title"))
			if err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			careerID, err := parseID(r.Form.Get("career_id"))
			if err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			plan, err := planService.GetCareerPlan(ctx, planModel.CareerID(careerID))
			if err != nil {
				logger.Error("cannot get career plan", "error", err)
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
			}

			assignments := r.Form["assignment_ids"]

			if len(assignments) == 0 {
				view.Error = "select at least one assignment"
				step2.Execute(w, view)
				return
			}

			if _, err := parseIDList(assignments); err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			q := url.Values{}
			q.Set("title", title)

			for _, id := range assignments {
				q.Add("assignment_ids", id)
			}

			customRedirect(w, r, "/schedule/step3?"+q.Encode())
		})

		// ================= STEP 3 =================

		r.Get("/step3", func(w http.ResponseWriter, r *http.Request) {

			title, err := requiredString(r.URL.Query().Get("title"))
			if err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			assignments := r.URL.Query()["assignment_ids"]

			ids, err := parseIDList(assignments)
			if err != nil || len(ids) == 0 {
				customRedirect(w, r, "/bad_form")
				return
			}

			subjects := make([]planModel.SubjectID, len(ids))
			for i, id := range ids {
				subjects[i] = planModel.SubjectID(id)
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			sections, err := planService.ListOffering(ctx, subjects)
			if err != nil {
				logger.Error("cannot list offerings", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			view := Step3View{
				Title:         title,
				AssignmentIDs: assignments,
				Sections:      sections,
			}

			if err := step3.Execute(w, view); err != nil {
				logger.Error("step3 template error", "error", err)
			}
		})

		r.Post("/step3", func(w http.ResponseWriter, r *http.Request) {

			userID := extractUserID(r)
			if userID == 0 {
				customRedirect(w, r, "/login")
				return
			}

			if err := r.ParseForm(); err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			title, err := requiredString(r.Form.Get("title"))
			if err != nil {
				customRedirect(w, r, "/bad_form")
				return
			}

			sectionIDs := r.Form["section_ids"]

			ids, err := parseIDList(sectionIDs)
			if err != nil || len(ids) == 0 {
				customRedirect(w, r, "/bad_form")
				return
			}

			courses := make([]courseOffering.CourseOfferingID, len(ids))
			for i, id := range ids {
				courses[i] = courseOffering.CourseOfferingID(id)
			}

			ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
			defer cancel()

			schedule, err := scheduleModel.NewSchedule(
				user.UserID(userID),
				title,
				courses,
			)
			if err != nil {
				logger.Error("cannot create schedule", "error", err)
				customRedirect(w, r, "/bad_form")
				return
			}

			scheduleID, err := scheduleService.Save(ctx, *schedule)
			if err != nil {
				logger.Error("cannot save schedule", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			// set this id into a cookie
			http.SetCookie(w, &http.Cookie{
				Name:     latestSelectionCookie,
				Value:    strconv.FormatInt(int64(scheduleID), 64),
				Path:     "/",
				HttpOnly: true,
				Secure:   config.Get().Security.SecureHTTP,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Unix(0, 0),
			})

			customRedirect(w, r, "/dashboard")
		})
	}
}
