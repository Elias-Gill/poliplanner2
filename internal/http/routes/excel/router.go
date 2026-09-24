package excel

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
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
	kind := parseSourceType(r.URL.Query().Get("type"))

	versions, err := h.excelService.ListVersions(r.Context(), kind)
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
		"Type":     kind,
		"IsLab":    kind == excelModel.SourceTypeLab,
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

	kind := parseSourceType(r.FormValue("type"))

	semester, err := parseSemester(r.FormValue("period"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadDate, err := parseUploadDate(r.FormValue("date"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	downloadURL, err := parseDownloadURL(r.FormValue("downloadUrl"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// The file is optional: when absent, the source is downloaded from the URL.
	file, header, fileErr := r.FormFile("file")
	hasFile := fileErr == nil
	if fileErr != nil && !errors.Is(fileErr, http.ErrMissingFile) {
		http.Error(w, "Invalid file upload: "+fileErr.Error(), http.StatusBadRequest)
		return
	}
	if hasFile {
		defer file.Close()
	}

	name := resolveSourceName(r.FormValue("name"), header, downloadURL)

	meta := source.SourceMetadata{
		Name:     name,
		URI:      downloadURL,
		Semester: semester,
		Date:     uploadDate.In(timezone.ParaguayTZ),
	}

	if err := h.persistSource(r.Context(), kind, file, hasFile, meta); err != nil {
		logger.Error("Could not process uploaded excel source", "type", kind, "error", err)
		http.Error(w, "Could not process the file: "+err.Error(), http.StatusBadRequest)
		return
	}

	respondHTML(w, http.StatusOK, "File processed successfully")
}

// persistSource builds the right source type from an uploaded file or a download
// URL and hands it to the excel service.
func (h *Handler) persistSource(
	ctx context.Context,
	kind excelModel.SourceType,
	file io.ReadCloser,
	hasFile bool,
	meta source.SourceMetadata,
) error {
	if kind == excelModel.SourceTypeLab {
		var src source.LabSource
		if hasFile {
			src = source.NewLabSourceFromReader(file, meta)
		} else {
			src = source.NewLabSourceFromURL(meta.URI, meta.Name, meta.Semester, meta.Date)
		}
		return h.excelService.PersistLabSource(ctx, src)
	}

	var src source.ScheduleSource
	if hasFile {
		src = source.NewScheduleSourceFromReader(file, meta)
	} else {
		src = source.NewScheduleSourceFromURL(meta.URI, meta.Name, meta.Semester, meta.Date)
	}
	return h.excelService.PersistScheduleSource(ctx, src)
}

// ==================== Form parsing helpers ====================

func parseSourceType(raw string) excelModel.SourceType {
	if strings.TrimSpace(strings.ToLower(raw)) == string(excelModel.SourceTypeLab) {
		return excelModel.SourceTypeLab
	}
	return excelModel.SourceTypeSchedule
}

func parseSemester(raw string) (academic.YearSemester, error) {
	semester, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || (semester != 1 && semester != 2) {
		return 0, errors.New("Invalid period, must be 1 or 2")
	}
	return academic.YearSemester(semester), nil
}

func parseUploadDate(raw string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, errors.New("Invalid date format, expected YYYY-MM-DD")
	}
	return date, nil
}

func parseDownloadURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("downloadUrl is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("downloadUrl must be a valid http(s) URL")
	}

	return raw, nil
}

// resolveSourceName prefers an explicit name, then the uploaded filename, and
// finally the basename of the download URL.
func resolveSourceName(raw string, header *multipart.FileHeader, downloadURL string) string {
	if name := strings.TrimSpace(raw); name != "" {
		return name
	}

	if header != nil && header.Filename != "" {
		return header.Filename
	}

	if parsed, err := url.Parse(downloadURL); err == nil {
		if base := filepath.Base(parsed.Path); base != "." && base != "/" && base != "" {
			return base
		}
	}

	return "excel-source"
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
