package dashboard

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	pdf "github.com/elias-gill/poliplanner2/internal/pdf"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
	"github.com/elias-gill/poliplanner2/internal/http/render"
	scheduleModel "github.com/elias-gill/poliplanner2/internal/model/schedule"
	"github.com/elias-gill/poliplanner2/internal/service/academic"
	"github.com/elias-gill/poliplanner2/internal/service/schedule"
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

	pdfLimiter := middleware.NewGlobalPDFLimiter(1, 10*time.Minute)

	r.With(pdfLimiter.Limit).Get("/pdf/{id}", h.DownloadPDF)

	return r
}

// DashboardPageData carries all the data required to render the main dashboard view.
type DashboardPageData struct {
	Schedules      []scheduleModel.ScheduleSummaryView
	SelectedID     int64
	ActiveSchedule *scheduleModel.StudentScheduleView
}

// dashboard renders the main dashboard page. If the user has existing schedules,
// it automatically fetches and loads the first one by default.
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.MustExtractUserID(r)

	// Si HTMX envía una petición GET a /dashboard/?id=123
	if idQuery := r.URL.Query().Get("id"); idQuery != "" && r.Header.Get("HX-Request") == "true" {
		scheduleID, err := strconv.ParseInt(idQuery, 10, 64)
		if err != nil {
			http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
			return
		}

		details, err := h.scheduleService.GetScheduleOverview(ctx, userID, scheduleModel.ScheduleID(scheduleID))
		if err != nil {
			log.Printf("Dashboard: error fetching schedule details for ID %d: %v", scheduleID, err)
			utils.Redirect(w, r, "/500")
			return
		}

		cookie.SetLatestScheduleCookie(w, scheduleModel.ScheduleID(scheduleID))

		// Renderizar SOLO el parcial
		err = h.tmpl.RenderPartial(w, "dashboard/index.html", "dashboard/schedule_content", details)
		if err != nil {
			log.Printf("Dashboard: error fetching schedule details for ID %d: %v", scheduleID, err)
			utils.Redirect(w, r, "/500")
			return
		}
		return
	}

	// Carga normal de la página completa
	userSchedules, err := h.scheduleService.ListUserSchedules(ctx, userID)
	if err != nil {
		log.Printf("[dashboard] Error fetching user schedules for userID %d: %v", userID, err)
		utils.Redirect(w, r, "/500")
		return
	}

	data := DashboardPageData{
		Schedules: userSchedules,
	}

	if len(userSchedules) > 0 {
		// 1. Intentar obtener el ID desde la cookie del último horario seleccionado
		selectedID := userSchedules[0].ID // Fallback por defecto al primero
		if cookieID, ok := cookie.GetLatestScheduleCookie(r); ok {
			// Validar que el scheduleID de la cookie pertenezca a la lista del usuario
			for _, s := range userSchedules {
				if s.ID == cookieID {
					selectedID = cookieID
					break
				}
			}
		}

		data.SelectedID = int64(selectedID)

		// 2. Obtener los detalles del schedule seleccionado
		details, err := h.scheduleService.GetScheduleOverview(ctx, userID, selectedID)
		if err != nil {
			log.Printf("Dashboard: error fetching schedule details for ID %d: %v", selectedID, err)
			utils.Redirect(w, r, "/500")
			return
		}

		data.ActiveSchedule = details
	}

	err = h.tmpl.RenderPage(w, "dashboard/index.html", data)
	if err != nil {
		log.Printf("Dashboard: error rendering page: %v", err)
	}
}

// dashboardSchedule fetches and renders a specific schedule view fragment for HTMX requests.
func (h *Handler) dashboardSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.MustExtractUserID(r)

	idStr := chi.URLParam(r, "id")
	scheduleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	details, err := h.scheduleService.GetScheduleOverview(ctx, userID, scheduleModel.ScheduleID(scheduleID))
	if err != nil {
		// FIX: DEBERIA  de dieferneciar entre error de server y no autorizado
		log.Printf("Dashboard: error fetching schedule details for ID %d: %v", scheduleID, err)
		utils.Redirect(w, r, "/500")
		return
	}

	cookie.SetLatestScheduleCookie(w, scheduleModel.ScheduleID(scheduleID))

	// Render partial HTML fragment for dynamic HTMX updates
	h.tmpl.RenderPartial(w, "dashboard/index.html", "dashboard/schedule_content", details)
}

func (h *Handler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.MustExtractUserID(r)

	idStr := chi.URLParam(r, "id")
	scheduleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	studentView, err := h.scheduleService.GetScheduleOverview(ctx, userID, scheduleModel.ScheduleID(scheduleID))
	if err != nil {
		http.Error(w, "Error generando vista", http.StatusInternalServerError)
		return
	}

	// 2. Configurar Headers HTTP
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"horario_%d.pdf\"", scheduleID))

	// 3. Exportar directamente al ResponseWriter
	exporter := pdf.NewSchedulePDFExporter()
	if _, err := exporter.Export(studentView, w); err != nil {
		log.Printf("Error al escribir el PDF: %v", err)
	}
}
