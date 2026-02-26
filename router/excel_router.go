package router

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/internal/source"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 8 << 20 // 8 MiB

func NewExcelRouter(excelService *service.ExcelService) func(r chi.Router) {
	cfg := config.Get()

	updateKey := cfg.Security.UpdateKey
	baseDir := path.Join(cfg.Paths.BaseDir, "web", "templates", "pages")

	syncFormPath := path.Join(baseDir, "excel", "sync-form.html")
	syncFormTemplate := parseTemplateWithBaseLayout(syncFormPath)

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			syncFormTemplate.Execute(w, nil)
		})

		r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
			if !isAuthorized(r.Header.Get("Authorization"), updateKey) {
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}

			ct := r.Header.Get("Content-Type")
			if strings.Contains(ct, "multipart/form-data") {
				handleUpload(w, r, excelService)
			} else {
				handleSync(w, r, excelService, cfg.Excel.ScraperTimeout)
			}
		})
	}
}

func isAuthorized(authHeader, expected string) bool {
	return strings.TrimSpace(authHeader) == "Bearer "+expected
}

// handleUpload creates a manual ExcelSource and passes it to the service
func handleUpload(w http.ResponseWriter, r *http.Request, svc *service.ExcelService) {
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

	// Create a minimal ExcelSource from the uploaded file
	source := source.NewExcelSourceFromReader(file, source.ExcelSourceMetadata{
		Name:   header.Filename,
		URI:    downloadURL,
		Period: period,
		Date:   time.Now(),
	})

	if err := svc.ParseAndPersistNewExcel(r.Context(), source); err != nil {
		http.Error(w, "Could not process the file: "+err.Error(), http.StatusBadRequest)
		return
	}

	respondHTML(w, http.StatusOK, "File processed successfully")
}

func handleSync(w http.ResponseWriter, r *http.Request, svc *service.ExcelService, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	if err := svc.SearchNewestExcel(ctx); err != nil {
		http.Error(w, "Sync failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	respondHTML(w, http.StatusOK, "Sync completed successfully")
}

// REFACTOR: I DONT like this
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
