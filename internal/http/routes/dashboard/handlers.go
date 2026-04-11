package dashboard

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	scheduleDomain "github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
	"github.com/elias-gill/poliplanner2/logger"
)

type overviewData struct {
	Info   []courseOffering.CourseSummary
	Weekly *courseOffering.CoursesScheduleView
	Exams  *courseOffering.ExamsScheduleView
}

type dashboardPage struct {
	Mode       string
	SelectedID int64
	Data       any
	Schedules  []scheduleDomain.ScheduleSummary
}

type DashboardHandlers struct {
	scheduleService *schedule.Schedule
	planService     *academicPlan.AcademicPlan

	// templates
	dashboardTemplate *template.Template
	calendarTemplate  *template.Template
}

func newDashboardHandlers(
	scheduleService *schedule.Schedule,
	planService *academicPlan.AcademicPlan,
) *DashboardHandlers {

	templateDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "dashboard")

	// Parse base layout
	base := utils.ParseTemplateWithBaseLayout(path.Join(templateDir, "index.html"))
	dashboardTemplate := template.Must(template.Must(base.Clone()).ParseFiles(path.Join(templateDir, "overview.html")))
	calendarTemplate := template.Must(template.Must(base.Clone()).ParseFiles(path.Join(templateDir, "calendar.html")))

	return &DashboardHandlers{
		scheduleService:   scheduleService,
		planService:       planService,
		dashboardTemplate: dashboardTemplate,
		calendarTemplate:  calendarTemplate,
	}
}

// ==================== Helper functions (internal) ====================

func (h *DashboardHandlers) executeTemplate(w http.ResponseWriter, tpl *template.Template, data any, tplName string) {
	if err := tpl.Execute(w, data); err != nil {
		logger.Debug("template execution failed", "template", tplName, "error", err)
		http.Error(w, "Ocurrió un error al cargar la página", http.StatusInternalServerError)
	}
}

func (h *DashboardHandlers) serveOverview(ctx context.Context, userID user.UserID, selectedID scheduleDomain.ScheduleID) (any, error) {
	sch, err := h.scheduleService.GetSchedule(ctx, userID, selectedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	weekly, err := h.planService.ListCoursesSchedule(ctx, sch.Courses)
	if err != nil {
		return nil, fmt.Errorf("failed to list weekly schedule: %w", err)
	}

	exams, err := h.planService.GetScheduleExamsView(ctx, sch.Courses)
	if err != nil {
		return nil, fmt.Errorf("failed to list exams: %w", err)
	}

	info, err := h.planService.ListCoursesInfo(ctx, sch.Courses)
	if err != nil {
		return nil, fmt.Errorf("failed to list courses info: %w", err)
	}

	return overviewData{Info: info, Weekly: weekly, Exams: exams}, nil
}

func (h *DashboardHandlers) serveCalendar(ctx context.Context, userID user.UserID, selectedID scheduleDomain.ScheduleID) (any, error) {
	sch, err := h.scheduleService.GetSchedule(ctx, userID, selectedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	exams, err := h.planService.ListCoursesExams(ctx, sch.Courses)
	if err != nil {
		return nil, fmt.Errorf("failed to list exams: %w", err)
	}

	return exams, nil
}

// ==================== HTTP Handlers ====================

func (h *DashboardHandlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode != "calendar" {
		mode = "overview"
	}

	// Determine selected schedule
	selected, present := getLatestSelectionCookie(r)
	queryID := r.URL.Query().Get("id")
	userID := middleware.MustExtractUserID(r)

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	schedules, err := h.scheduleService.ListUserSchedules(ctx, user.UserID(userID))
	if err != nil {
		utils.CustomRedirect(w, r, "/500")
		return
	}

	// If no schedule created, then execute the template with empty data
	if len(schedules) == 0 {
		h.dashboardTemplate.Execute(w, dashboardPage{
			Mode:      mode,
			Data:      nil,
			Schedules: schedules,
		})
		return
	}

	var selectedID scheduleDomain.ScheduleID
	if queryID != "" {
		if qid, err := strconv.ParseInt(queryID, 10, 64); err == nil {
			selectedID = scheduleDomain.ScheduleID(qid)
		}
	} else if present {
		selectedID = scheduleDomain.ScheduleID(selected)
	} else if len(schedules) > 0 {
		selectedID = schedules[len(schedules)-1].ID
	}

	// Load page data
	var data any
	if mode == "calendar" {
		data, err = h.serveCalendar(ctx, user.UserID(userID), selectedID)
		if err != nil {
			utils.CustomRedirect(w, r, "/500")
			return
		}
		h.executeTemplate(w, h.calendarTemplate, dashboardPage{
			Mode:       mode,
			SelectedID: int64(selectedID),
			Data:       data,
			Schedules:  schedules,
		}, "calendar.html")
		return
	}

	data, err = h.serveOverview(ctx, user.UserID(userID), selectedID)
	if err != nil {
		logger.Debug(err.Error())
		utils.CustomRedirect(w, r, "/500")
		return
	}

	h.executeTemplate(w, h.dashboardTemplate, dashboardPage{
		Mode:       mode,
		SelectedID: int64(selectedID),
		Data:       data,
		Schedules:  schedules,
	}, "overview.html")
}

// getLatestSelectionCookie returns the last selected schedule from cookie
func getLatestSelectionCookie(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(utils.LatestSelectionCookie)
	if err != nil {
		return -1, false
	}

	id, err := strconv.ParseInt(cookie.Value, 10, 64)
	if err != nil || id < 1 {
		return -1, false
	}

	return id, true
}
