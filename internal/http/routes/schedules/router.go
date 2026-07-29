package schedules

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
	"github.com/elias-gill/poliplanner2/internal/http/render"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/schedule"
	academicSrvs "github.com/elias-gill/poliplanner2/internal/service/academic"
	scheduleSrvs "github.com/elias-gill/poliplanner2/internal/service/schedule"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	tmpl              *render.TemplateManager
	scheduleService   *scheduleSrvs.ScheduleService
	careerService     *academicSrvs.CareerService
	courseService     *academicSrvs.CourseService
	curriculumService *academicSrvs.CurriculumService
}

func NewHandler(
	tmpl *render.TemplateManager,
	scheduleService *scheduleSrvs.ScheduleService,
	careerService *academicSrvs.CareerService,
	courseService *academicSrvs.CourseService,
	curriculumService *academicSrvs.CurriculumService,
) *Handler {
	return &Handler{
		tmpl:              tmpl,
		scheduleService:   scheduleService,
		careerService:     careerService,
		courseService:     courseService,
		curriculumService: curriculumService,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.index)
	r.Get("/subjects", h.listCurriculum)

	r.Get("/malla/{id}/courses", h.listOffering)

	r.Post("/", h.saveSchedule)

	r.Post("/delete", h.deleteSchedule)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	careers, err := h.careerService.ListCareers(ctx)
	if err != nil {
		logger.Error("cannot list careers for index schedule page", "error", err)
		utils.Redirect(w, r, "/500")
		return
	}

	data := map[string]any{
		"Title":   "Crear Horario",
		"Careers": careers,
	}

	if err := h.tmpl.RenderPage(w, "schedules/index.html", data); err != nil {
		logger.Error("cannot render schedule index page", "error", err)
		utils.Redirect(w, r, "/500")
	}
}

func (h *Handler) listCurriculum(w http.ResponseWriter, r *http.Request) {
	careerCode := r.URL.Query().Get("career_id")
	if careerCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	careerID, err := strconv.Atoi(careerCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	curriculum, err := h.curriculumService.GetCurriculum(ctx, academic.CareerID(careerID))
	if err != nil {
		logger.Error("cannot retrieve curriculum subjects", "error", err)
		utils.Redirect(w, r, "/500")
		return
	}

	data := map[string]any{
		"CareerCode": careerCode,
		"Plans":      curriculum.Plans,
		"Levels":     curriculum.Levels,
		"Subjects":   curriculum.Subjects,
		"Semesters":  curriculum.Semesters,
	}

	// FIX: cuando renderizo el partial, deberia de redirigir a la pagina principal de la que
	// deberia de salir
	if err := h.tmpl.RenderPartial(w, "schedules/index.html", "subjects_list", data); err != nil {
		logger.Error("cannot render subjects list partial", "error", err)
		utils.Redirect(w, r, "/500")
	}
}

// FIX: implementar filtros por enfasis
func (h *Handler) listOffering(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	mallaID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de materia inválido", http.StatusBadRequest)
		return
	}

	// Aditional parameters for the shop car
	careerCode := r.URL.Query().Get("career_code")
	subjectName := r.URL.Query().Get("subject_name")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	courses, err := h.courseService.GetOfferings(ctx, academic.CurriculumID(mallaID))
	if err != nil {
		logger.Error("cannot get course offerings", "malla_id", mallaID, "error", err)
		http.Error(w, "Error al obtener secciones", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Courses":     courses,
		"CareerCode":  careerCode,
		"SubjectName": subjectName,
	}

	if err := h.tmpl.RenderPartial(w, "schedules/index.html", "course_offerings", data); err != nil {
		logger.Error("cannot render course offerings partial", "error", err)
		http.Error(w, "Error al renderizar la plantilla", http.StatusInternalServerError)
	}
}

// REFACTOR: simplificar, demasiado json medio para nada a mi parecer
type selectedItem struct {
	ID int64 `json:"id"`
}

func (h *Handler) saveSchedule(w http.ResponseWriter, r *http.Request) {
	userID := utils.MustExtractUserID(r)
	if userID == 0 {
		w.Header().Set("HX-Redirect", "/login")
		utils.Redirect(w, r, "/login")
		return
	}

	if err := r.ParseForm(); err != nil {
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	// 1. Lectura de campos del formulario
	title, err := utils.RequiredString(r.Form.Get("title"))
	if err != nil {
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	jsonItems := r.Form.Get("selected_items_json")
	if jsonItems == "" {
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	// 2. Deserialización del payload enviado por Alpine.js
	var items []selectedItem
	if err := json.Unmarshal([]byte(jsonItems), &items); err != nil || len(items) == 0 {
		logger.Error("error al deserializar selected_items_json", "error", err)
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	// 3. Mapeo a []academic.CourseID
	courses := make([]academic.CourseID, len(items))
	for i, item := range items {
		courses[i] = academic.CourseID(item.ID)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	// 4. Delegar la creación completa al Service
	scheduleID, err := h.scheduleService.CreateSchedule(ctx, userID, title, courses)
	if err != nil {
		// Manejo diferencial según si fue un error de formulario/validación o fallo interno
		if errors.Is(err, scheduleSrvs.ErrTitleNotAvailable) {
			w.Header().Set("HX-Redirect", "/bad_form")
			utils.Redirect(w, r, "/bad_form")
			return
		}

		logger.Error("failed to create schedule via service", "error", err)
		w.Header().Set("HX-Redirect", "/500")
		utils.Redirect(w, r, "/500")
		return
	}

	// 5. Guardar Cookie de sesión/último horario
	cookie.SetLatestScheduleCookie(w, scheduleID)

	// 6. Redireccionar al Dashboard (Soporte HTMX y HTTP nativo)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.Redirect(w, r, "/dashboard")
}

func (h *Handler) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	userID := utils.MustExtractUserID(r)
	if userID == 0 {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusOK)
			return
		}
		utils.Redirect(w, r, "/login")
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("error al parsear formulario de eliminación", "error", err)
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	// 1. Obtener y validar el ID del horario
	idStr := r.Form.Get("id")
	if idStr == "" {
		logger.Error("id de horario no provisto para eliminación")
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("id de horario inválido", "id", idStr, "error", err)
		w.Header().Set("HX-Redirect", "/bad_form")
		utils.Redirect(w, r, "/bad_form")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	// 2. Ejecutar borrado a través del servicio
	err = h.scheduleService.Delete(ctx, userID, schedule.ScheduleID(id))
	if err != nil {
		if errors.Is(err, scheduleSrvs.ErrPermissionDenied) {
			logger.Error("permiso denegado para eliminar horario", "user_id", userID, "schedule_id", id)
			w.Header().Set("HX-Redirect", "/403")
			utils.Redirect(w, r, "/403")
			return
		}

		logger.Error("falló la eliminación del horario", "schedule_id", id, "error", err)
		w.Header().Set("HX-Redirect", "/500")
		utils.Redirect(w, r, "/500")
		return
	}

	// 3. Redirección exitosa (Soporte HTMX y HTTP nativo)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.Redirect(w, r, "/dashboard")
}
