package excel

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	excelimport "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
)

const maxUploadSize = 8 << 20 // 8 MiB

type ExcelHandlers struct {
	excelService     *excelimport.ImportService
	syncFormTemplate *template.Template // se usará el parse que ya tienes en utils
	updateKey        string
	scraperTimeout   time.Duration
}

func newExcelHandlers(excelService *excelimport.ImportService) *ExcelHandlers {
	cfg := config.Get()

	baseDir := path.Join(cfg.Paths.BaseDir, "web", "templates", "pages")
	syncFormPath := path.Join(baseDir, "excel", "sync-form.html")
	syncFormTemplate := utils.ParseTemplateWithBaseLayout(syncFormPath)

	return &ExcelHandlers{
		excelService:     excelService,
		syncFormTemplate: syncFormTemplate,
		updateKey:        cfg.Security.UpdateKey,
		scraperTimeout:   cfg.Excel.ScraperTimeout,
	}
}

// ==================== Handlers ====================

func (h *ExcelHandlers) SyncForm(w http.ResponseWriter, r *http.Request) {
	h.syncFormTemplate.Execute(w, nil)
}

func (h *ExcelHandlers) Sync(w http.ResponseWriter, r *http.Request) {
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

// ==================== Helper methods ====================

func (h *ExcelHandlers) isAuthorized(authHeader string) bool {
	return strings.TrimSpace(authHeader) == "Bearer "+h.updateKey
}

func (h *ExcelHandlers) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Excel file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	periodStr := r.FormValue("period")
	period, err := strconv.Atoi(periodStr)
	if err != nil || (period != 1 && period != 2) {
		http.Error(w, "Invalid period, must be 1 or 2", http.StatusBadRequest)
		return
	}

	downloadURL := strings.TrimSpace(r.FormValue("downloadUrl"))
	if downloadURL == "" {
		downloadURL = "manual-upload"
	}

	source := source.NewExcelSourceFromReader(file, source.ExcelSourceMetadata{
		Name:   header.Filename,
		URI:    downloadURL,
		Period: period,
		Date:   time.Now().In(timezone.ParaguayTZ),
	})

	if err := h.excelService.PersistSource(r.Context(), source); err != nil {
		http.Error(w, "Could not process the file: "+err.Error(), http.StatusBadRequest)
		return
	}

	respondHTML(w, http.StatusOK, "File processed successfully")
}

func (h *ExcelHandlers) handleSync(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.scraperTimeout)
	defer cancel()

	if err := h.excelService.Sync(ctx); err != nil {
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
