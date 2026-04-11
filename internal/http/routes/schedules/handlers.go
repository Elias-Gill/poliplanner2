package schedules

import (
	"context"
	"html/template"
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
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
	"github.com/elias-gill/poliplanner2/logger"
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

type SchedulesHandlers struct {
	scheduleService *schedule.ScheduleService
	planService     *academicPlan.AcademicPlanService

	// templates
	step1 *template.Template
	step2 *template.Template
	step3 *template.Template
}

func newSchedulesHandlers(
	scheduleService *schedule.ScheduleService,
	planService *academicPlan.AcademicPlanService,
) *SchedulesHandlers {

	basePath := path.Join(
		config.Get().Paths.BaseDir,
		"web", "templates", "pages", "schedules",
	)

	return &SchedulesHandlers{
		scheduleService: scheduleService,
		planService:     planService,
		step1:           utils.ParseTemplateWithBaseLayout(path.Join(basePath, "step1.html")),
		step2:           utils.ParseTemplateWithBaseLayout(path.Join(basePath, "step2.html")),
		step3:           utils.ParseTemplateWithBaseLayout(path.Join(basePath, "step3.html")),
	}
}

// ==================== STEP 1 ====================

func (h *SchedulesHandlers) Step1Get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	careers, err := h.planService.ListCareers(ctx)
	if err != nil {
		logger.Error("cannot list careers", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}

	view := Step1View{
		Careers: careers,
	}

	if err := h.step1.Execute(w, view); err != nil {
		logger.Error("step1 template error", "error", err)
	}
}

func (h *SchedulesHandlers) Step1Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	careers, err := h.planService.ListCareers(ctx)
	if err != nil {
		logger.Error("cannot list careers", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}

	view := Step1View{
		Careers: careers,
		Title:   r.Form.Get("title"),
		Career:  r.Form.Get("career_id"),
	}

	title, err := utils.RequiredString(view.Title)
	if err != nil {
		view.Error = "el título es obligatorio"
		h.step1.Execute(w, view)
		return
	}

	careerID, err := utils.ParseID(view.Career)
	if err != nil {
		view.Error = "carrera seleccionada no válida"
		h.step1.Execute(w, view)
		return
	}

	userID := middleware.MustExtractUserID(r)
	available, err := h.scheduleService.TitleIsAvailable(ctx, userID, title)
	if err != nil {
		utils.CustomRedirect(w, r, "/500")
		return
	}
	if !available {
		view.Error = "ya existe un horario con este título"
		h.step1.Execute(w, view)
		return
	}

	q := url.Values{}
	q.Set("title", title)
	q.Set("career_id", strconv.FormatInt(careerID, 10))
	utils.CustomRedirect(w, r, "/schedule/step2?"+q.Encode())
}

// ==================== STEP 2 ====================

func (h *SchedulesHandlers) Step2Get(w http.ResponseWriter, r *http.Request) {
	title, err := utils.RequiredString(r.URL.Query().Get("title"))
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	careerID, err := utils.ParseID(r.URL.Query().Get("career_id"))
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	plan, err := h.planService.GetCareerPlan(ctx, planModel.CareerID(careerID))
	if err != nil {
		logger.Error("cannot get career plan", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}

	if plan == nil {
		utils.CustomRedirect(w, r, "/404")
		return
	}

	view := Step2View{
		Title:    title,
		CareerID: careerID,
		Plan:     plan,
	}

	if err := h.step2.Execute(w, view); err != nil {
		logger.Error("step2 template error", "error", err)
	}
}

func (h *SchedulesHandlers) Step2Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	title, err := utils.RequiredString(r.Form.Get("title"))
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	careerID, err := utils.ParseID(r.Form.Get("career_id"))
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	plan, err := h.planService.GetCareerPlan(ctx, planModel.CareerID(careerID))
	if err != nil {
		logger.Error("cannot get career plan", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}

	if plan == nil {
		utils.CustomRedirect(w, r, "/404")
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
		h.step2.Execute(w, view)
		return
	}

	if _, err := utils.ParseIDList(assignments); err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	q := url.Values{}
	q.Set("title", title)

	for _, id := range assignments {
		q.Add("assignment_ids", id)
	}

	utils.CustomRedirect(w, r, "/schedule/step3?"+q.Encode())
}

// ==================== STEP 3 ====================

func (h *SchedulesHandlers) Step3Get(w http.ResponseWriter, r *http.Request) {
	title, err := utils.RequiredString(r.URL.Query().Get("title"))
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	assignments := r.URL.Query()["assignment_ids"]

	ids, err := utils.ParseIDList(assignments)
	if err != nil || len(ids) == 0 {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	subjects := make([]planModel.SubjectID, len(ids))
	for i, id := range ids {
		subjects[i] = planModel.SubjectID(id)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	sections, err := h.planService.ListOffering(ctx, subjects)
	if err != nil {
		logger.Error("cannot list offerings", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}

	view := Step3View{
		Title:         title,
		AssignmentIDs: assignments,
		Sections:      sections,
	}

	if err := h.step3.Execute(w, view); err != nil {
		logger.Error("step3 template error", "error", err)
	}
}

func (h *SchedulesHandlers) Step3Post(w http.ResponseWriter, r *http.Request) {
	userID := middleware.MustExtractUserID(r)
	if userID == 0 {
		utils.CustomRedirect(w, r, "/login")
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	title, err := utils.RequiredString(r.Form.Get("title"))
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	sectionIDs := r.Form["section_ids"]

	ids, err := utils.ParseIDList(sectionIDs)
	if err != nil || len(ids) == 0 {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	courses := make([]courseOffering.CourseOfferingID, len(ids))
	for i, id := range ids {
		courses[i] = courseOffering.CourseOfferingID(id)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	available, err := h.scheduleService.TitleIsAvailable(ctx, userID, title)
	if err != nil {
		logger.Error("cannot check title existence", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}
	if !available {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	schedule, err := scheduleModel.NewSchedule(
		userID,
		title,
		courses,
	)
	if err != nil {
		utils.CustomRedirect(w, r, "/bad_form")
		return
	}

	scheduleID, err := h.scheduleService.Save(ctx, *schedule)
	if err != nil {
		logger.Error("cannot save schedule", "error", err)
		utils.CustomRedirect(w, r, "/500")
		return
	}

	// set this id into a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     utils.LatestSelectionCookie,
		Value:    strconv.FormatInt(int64(scheduleID), 10),
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Get().Security.SecureHTTP,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})

	utils.CustomRedirect(w, r, "/dashboard")
}
