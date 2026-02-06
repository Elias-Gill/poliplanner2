package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/scraper"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 8 << 20 // 8 MiB

func NewExcelRouter(excelService *service.ExcelService) func(r chi.Router) {
	cfg := config.Get()
	layout := web.BaseLayout
	updateKey := cfg.Security.UpdateKey

	return func(r chi.Router) {
		// Form to upload Excel or trigger sync
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			execTemplateWithLayout(w, "web/templates/pages/excel/sync-form.html", layout, nil)
		})

		// Protected endpoint for upload or forced sync
		r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
			if !validateAuthHeader(r.Header.Get("Authorization"), updateKey) {
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}

			contentType := r.Header.Get("Content-Type")
			if strings.Contains(contentType, "multipart/form-data") {
				handleExcelUpload(w, r, excelService)
			} else {
				handleForcedSync(w, r, excelService)
			}
		})
	}
}

// Validates the Bearer token for admin actions
func validateAuthHeader(header, expectedKey string) bool {
	if header == "" {
		return false
	}
	header = strings.TrimSpace(header)
	return header == "Bearer "+expectedKey
}

// Handles file upload (manual Excel version)
func handleExcelUpload(w http.ResponseWriter, r *http.Request, svc *service.ExcelService) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.Warn("Failed to parse multipart form", "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logger.Warn("Missing or invalid file", "error", err)
		http.Error(w, "Excel file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "excel-upload-*.xlsx")
	if err != nil {
		logger.Error("Failed to create temp file", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err = io.Copy(tmpFile, file); err != nil {
		logger.Error("Failed to save uploaded file", "error", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	downloadURL := strings.TrimSpace(r.FormValue("downloadUrl"))
	if downloadURL == "" {
		downloadURL = "manual-upload"
	}

	source := scraper.NewExcelDownloadSource(
		downloadURL,
		header.Filename,
		time.Now(),
		1, // TODO: ask user for period or detect from file name
	)

	if err := svc.ParseAndPersistExcelFile(r.Context(), tmpFile.Name(), source); err != nil {
		logger.Error("Failed to process Excel", "error", err)
		respondError(w, "Could not process the file: "+err.Error())
		return
	}

	respondSuccess(w, "File processed successfully")
}

// Triggers automatic scraper sync (forced update)
func handleForcedSync(w http.ResponseWriter, r *http.Request, svc *service.ExcelService) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Get().Excel.ScraperTimeout)
	defer cancel()

	if err := svc.SearchNewestExcel(ctx); err != nil {
		logger.Error("Forced sync failed", "error", err)
		respondError(w, "Sync failed: "+err.Error())
		return
	}

	respondSuccess(w, "Sync completed successfully")
}

// HTMX-friendly success response
func respondSuccess(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<div class="alert success">
			<span>%s</span>
			<button onclick="this.parentElement.remove()">×</button>
		</div>
	`, msg)
}

// HTMX-friendly error response
func respondError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `
		<div class="alert error">
			<span>%s</span>
			<button onclick="this.parentElement.remove()">×</button>
		</div>
	`, msg)
}
