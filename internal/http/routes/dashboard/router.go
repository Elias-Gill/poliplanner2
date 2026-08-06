package dashboard

import (
	"errors"
	"net/http"
	"strconv"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
	scheduleModel "github.com/elias-gill/poliplanner2/internal/model/schedule"
	render "github.com/elias-gill/poliplanner2/internal/render/html"
	"github.com/elias-gill/poliplanner2/internal/service/academic"
	"github.com/elias-gill/poliplanner2/internal/service/schedule"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests related to the student dashboard.
type Handler struct {
	tmpl            *render.TemplateManager
	scheduleService *schedule.ScheduleService
	planService     *academic.CourseService
}

// NewHandler constructs a new Handler instance.
func NewHandler(
	tmpl *render.TemplateManager,
	scheduleService *schedule.ScheduleService,
	planService *academic.CourseService,
) *Handler {
	return &Handler{
		tmpl:            tmpl,
		scheduleService: scheduleService,
		planService:     planService,
	}
}

// Routes sets up the HTTP router for the dashboard endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.dashboard)
	r.Get("/{id}", h.dashboardSchedule)

	return r
}

// DashboardPageData carries all the data required to render the main dashboard view.
type DashboardPageData struct {
	Schedules      []scheduleModel.ScheduleSummaryView
	SelectedID     int64
	ActiveSchedule *scheduleModel.StudentScheduleView
}

// dashboard renders the main dashboard page or partials if requested via HTMX query parameter.
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.MustExtractUserID(r)

	// Handle HTMX partial requests sent via query parameter (?id=123)
	if idQuery := r.URL.Query().Get("id"); idQuery != "" && r.Header.Get("HX-Request") == "true" {
		scheduleID, err := strconv.ParseInt(idQuery, 10, 64)
		if err != nil {
			utils.Redirect(w, r, "/404")
			return
		}

		details, err := h.scheduleService.GetScheduleOverview(ctx, userID, scheduleModel.ScheduleID(scheduleID))
		if err != nil {
			h.handleOverviewError(w, r, err)
			return
		}

		cookie.SetLatestScheduleCookie(w, scheduleModel.ScheduleID(scheduleID))

		err = h.tmpl.RenderPartial(w, "dashboard/index.html", "dashboard/schedule_content", details)
		if err != nil {
			logger.Error("Failed to render HTMX schedule partial", "scheduleID", scheduleID, "error", err)
		}
		return
	}

	// Full page initial load
	userSchedules, err := h.scheduleService.ListUserSchedules(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch user schedules", "userID", userID, "error", err)
		utils.Redirect(w, r, "/500")
		return
	}

	data := DashboardPageData{
		Schedules: userSchedules,
	}

	if len(userSchedules) > 0 {
		// 1. Try to get default schedule ID from cookie fallback
		selectedID := userSchedules[0].ID
		if cookieID, ok := cookie.GetLatestScheduleCookie(r); ok {
			for _, s := range userSchedules {
				if s.ID == cookieID {
					selectedID = cookieID
					break
				}
			}
		}

		data.SelectedID = int64(selectedID)

		// 2. Load active schedule overview
		details, err := h.scheduleService.GetScheduleOverview(ctx, userID, selectedID)
		if err != nil {
			h.handleOverviewError(w, r, err)
			return
		}

		data.ActiveSchedule = details
	}

	err = h.tmpl.RenderPage(w, "dashboard/index.html", data)
	if err != nil {
		logger.Error("Failed to render dashboard page", "userID", userID, "error", err)
	}
}

// dashboardSchedule fetches and renders a specific schedule view fragment for HTMX requests.
func (h *Handler) dashboardSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.MustExtractUserID(r)

	idStr := chi.URLParam(r, "id")
	scheduleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Redirect(w, r, "/404")
		return
	}

	details, err := h.scheduleService.GetScheduleOverview(ctx, userID, scheduleModel.ScheduleID(scheduleID))
	if err != nil {
		h.handleOverviewError(w, r, err)
		return
	}

	cookie.SetLatestScheduleCookie(w, scheduleModel.ScheduleID(scheduleID))

	err = h.tmpl.RenderPartial(w, "dashboard/index.html", "dashboard/schedule_content", details)
	if err != nil {
		logger.Error("Failed to render schedule partial fragment", "scheduleID", scheduleID, "error", err)
	}
}

// handleOverviewError logs errors and triggers appropriate HTMX redirects using utils.Redirect.
func (h *Handler) handleOverviewError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, schedule.ErrNotFound):
		utils.Redirect(w, r, "/404")

	case errors.Is(err, schedule.ErrPermissionDenied):
		utils.Redirect(w, r, "/permission_denied")

	default:
		utils.Redirect(w, r, "/500")
	}
}
