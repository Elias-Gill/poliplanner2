package excel

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	excelModel "github.com/elias-gill/poliplanner2/internal/model/excel"
	render "github.com/elias-gill/poliplanner2/internal/render/html"
	"github.com/elias-gill/poliplanner2/internal/service/excel"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 8 << 20 // 8 MiB

type Handler struct {
	tmpl           *render.TemplateManager
	excelService   *excel.ExcelService
	syncService    *excel.SyncService
	updateKey      string
	scraperTimeout time.Duration
}

func NewHandler(
	tmpl *render.TemplateManager,
	excelService *excel.ExcelService,
	syncService *excel.SyncService,
	updateKey string,
	scraperTimeout time.Duration,
) *Handler {
	return &Handler{
		tmpl:           tmpl,
		excelService:   excelService,
		syncService:    syncService,
		updateKey:      updateKey,
		scraperTimeout: scraperTimeout,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.syncForm)
	r.Post("/sync", h.sync)
	r.Get("/list", h.listVersions) // <-- Nuevo endpoint para listar las versiones

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) syncForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "excel/sync-form.html", nil); err != nil {
		logger.Error("Cannot render sync-form template", "error", err)
	}
}

func (h *Handler) sync(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("Authorization")) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		h.handleUpload(w, r)
	} else {
		h.handleSync(w, r)
	}
}

func (h *Handler) listVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.excelService.ListVersions(r.Context(), excelModel.SourceTypeSchedule)
	if err != nil {
		logger.Error("Error listing excel versions", "error", err)
		http.Error(w, "No se pudieron obtener las versiones de Excel", http.StatusInternalServerError)
		return
	}

	state, err := h.syncService.GetSyncState(r.Context(), excelModel.SourceTypeSchedule)
	if err != nil {
		logger.Error("Error listing excel versions", "error", err)
		http.Error(w, "No se pudo obtener el ultimo auto sync de versiones excel", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Versions": versions,
		"LastSync": state.LastSearchAt,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "excel/list-versions.html", data); err != nil {
		logger.Error("Cannot render list-versions template", "error", err)
	}
}

// ==================== Helper methods ====================

func (h *Handler) isAuthorized(authHeader string) bool {
	expected := "Bearer " + h.updateKey
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(authHeader)), []byte(expected)) == 1
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Invalid form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Excel file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	periodStr := r.FormValue("period")
	semester, err := strconv.Atoi(periodStr)
	if err != nil || (semester != 1 && semester != 2) {
		http.Error(w, "Invalid period, must be 1 or 2", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	uploadDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	downloadURL := strings.TrimSpace(r.FormValue("downloadUrl"))
	if downloadURL == "" {
		downloadURL = "manual-upload"
	}

	src := source.NewScheduleSourceFromReader(file, source.SourceMetadata{
		Name:     header.Filename,
		URI:      downloadURL,
		Semester: academic.YearSemester(semester),
		Date:     uploadDate.In(timezone.ParaguayTZ),
	})

	if err := h.excelService.PersistScheduleSource(r.Context(), src); err != nil {
		http.Error(w, "Could not process the file: "+err.Error(), http.StatusBadRequest)
		return
	}

	respondHTML(w, http.StatusOK, "File processed successfully")
}

func (h *Handler) handleSync(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.scraperTimeout)
	defer cancel()

	if err := h.syncService.Sync(ctx); err != nil {
		http.Error(w, "Sync failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	respondHTML(w, http.StatusOK, "Sync completed successfully")
}

func respondHTML(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(
		w,
		`<div class="alert %s"><span>%s</span><button onclick="this.parentElement.remove()">×</button></div>`,
		alertClass(status),
		msg,
	)
}

func alertClass(status int) string {
	if status >= 200 && status < 300 {
		return "success"
	}
	return "error"
}
