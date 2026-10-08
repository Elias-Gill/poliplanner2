package tools

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	excelModel "github.com/elias-gill/poliplanner2/internal/model/excel"
	render "github.com/elias-gill/poliplanner2/internal/render/html"
	academicSrv "github.com/elias-gill/poliplanner2/internal/service/academic"
	excelSrv "github.com/elias-gill/poliplanner2/internal/service/excel"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

// excelVersionsLimit caps how many versions are listed per source type.
const excelVersionsLimit = 8

type Handler struct {
	tmpl              *render.TemplateManager
	careerService     *academicSrv.CareerService
	curriculumService *academicSrv.CurriculumService
	courseService     *academicSrv.CourseService
	teacherService    *academicSrv.TeacherService
	excelService      *excelSrv.ExcelService
	syncService       *excelSrv.SyncService
}

func NewHandler(
	tmpl *render.TemplateManager,
	careerSrv *academicSrv.CareerService,
	currSrv *academicSrv.CurriculumService,
	courseSrv *academicSrv.CourseService,
	teacherSrv *academicSrv.TeacherService,
	excelService *excelSrv.ExcelService,
	syncService *excelSrv.SyncService,
) *Handler {
	return &Handler{
		tmpl:              tmpl,
		careerService:     careerSrv,
		curriculumService: currSrv,
		courseService:     courseSrv,
		teacherService:    teacherSrv,
		excelService:      excelService,
		syncService:       syncService,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.index)
	r.Get("/calculator", h.calculator)
	r.Get("/interactive_graph", h.interactiveGraph)

	r.Get("/course-offering-history", h.courseOfferingHistory)
	r.Get("/course-offering-history/list-subjects", h.listSubjects)
	r.Get("/course-offering-history/timeline", h.getTimeline)

	r.Get("/teacher-history", h.teacherHistory)
	r.Get("/teacher-history/timeline", h.getTeacherTimeline)

	r.Get("/excel-list", h.excelList)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/index.html", nil); err != nil {
		logger.Error("Cannot render index tools template", "error", err)
	}
}

func (h *Handler) calculator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/calculator.html", nil); err != nil {
		logger.Error("Cannot render calculator template", "error", err)
	}
}

func (h *Handler) interactiveGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/interactive_graph.html", nil); err != nil {
		logger.Error("Cannot render interactive_graph template", "error", err)
	}
}

func (h *Handler) excelList(w http.ResponseWriter, r *http.Request) {
	kind := parseExcelSourceType(r.URL.Query().Get("type"))

	versions, err := h.excelService.ListVersions(r.Context(), kind, excelVersionsLimit)
	if err != nil {
		logger.Error("Error listing excel versions", "error", err)
		http.Error(w, "No se pudieron obtener las versiones de Excel", http.StatusInternalServerError)
		return
	}

	state, err := h.syncService.GetSyncState(r.Context(), kind)
	if err != nil {
		logger.Error("Error getting excel sync state", "error", err)
		http.Error(w, "No se pudo obtener el ultimo auto sync de versiones excel", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Versions": versions,
		"LastSync": state.LastSearchAt,
		"IsLab":    kind == excelModel.SourceTypeLab,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/excel_list.html", data); err != nil {
		logger.Error("Cannot render excel_list template", "error", err)
	}
}

func parseExcelSourceType(raw string) excelModel.SourceType {
	if strings.TrimSpace(strings.ToLower(raw)) == string(excelModel.SourceTypeLab) {
		return excelModel.SourceTypeLab
	}
	return excelModel.SourceTypeSchedule
}

func (h *Handler) courseOfferingHistory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	careers, err := h.careerService.ListCareers(ctx)
	if err != nil {
		logger.Error("cannot list careers for history page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Careers": careers,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/course_offering_history.html", data); err != nil {
		logger.Error("Cannot render course_offering_history template", "error", err)
	}
}

func (h *Handler) listSubjects(w http.ResponseWriter, r *http.Request) {
	careerIDStr := r.URL.Query().Get("career_id")
	careerID, err := strconv.Atoi(careerIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	curriculum, err := h.curriculumService.GetCurriculum(r.Context(), academicModel.CareerID(careerID))
	if err != nil {
		logger.Error("cannot retrieve curriculum subjects", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Subjects": curriculum.Subjects,
	}

	if err := h.tmpl.RenderPartial(w, "tools/course_offering_history.html", "subjects_select", data); err != nil {
		logger.Error("cannot render subjects select partial", "error", err)
	}
}

func (h *Handler) getTimeline(w http.ResponseWriter, r *http.Request) {
	curriculumIDStr := r.URL.Query().Get("curriculum_id")
	curriculumID, err := strconv.Atoi(curriculumIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	history, err := h.courseService.GetHistoricOfferings(r.Context(), academicModel.CurriculumID(curriculumID))
	if err != nil {
		logger.Error("cannot retrieve subject history", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"History": history,
	}

	if err := h.tmpl.RenderPartial(w, "tools/course_offering_history.html", "timeline", data); err != nil {
		logger.Error("cannot render timeline partial", "error", err)
	}
}

func (h *Handler) teacherHistory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	teachers, err := h.teacherService.ListTeachers(ctx)
	if err != nil {
		logger.Error("cannot list teachers for history page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Teachers": teachers,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/teacher_history.html", data); err != nil {
		logger.Error("Cannot render teacher_history template", "error", err)
	}
}

func (h *Handler) getTeacherTimeline(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacher_id")
	teacherID, err := strconv.Atoi(teacherIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	view, err := h.teacherService.GetTeacherHistory(r.Context(), academicModel.TeacherID(teacherID))
	if err != nil {
		logger.Error("cannot retrieve teacher history", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if view == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := map[string]any{
		"History": view,
	}

	if err := h.tmpl.RenderPartial(w, "tools/teacher_history.html", "teacher_timeline", data); err != nil {
		logger.Error("cannot render teacher timeline partial", "error", err)
	}
}
