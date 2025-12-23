package router

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 8 << 20 // 8MB

func NewExcelRouter(service *service.ExcelService) func(r chi.Router) {
	key := config.Get().Security.UpdateKey
	layout := web.BaseLayout

	return func(r chi.Router) {
		r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !isValidRequest(header, key) {
				http.Error(w, "invalid credentials", http.StatusForbidden)
				return
			}

			contentType := r.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "multipart/form-data") {
				handleFileUpload(w, r, service)
			} else {
				handleSyncRequest(w, r, service)
			}
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			err := execTemplateWithLayout(w, "web/templates/pages/excel/sync-form.html", layout, nil)
			if err != nil {
				http.Redirect(w, r, "/500", http.StatusFound)
			}
		})
	}
}

func handleFileUpload(w http.ResponseWriter, r *http.Request, service *service.ExcelService) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse form
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract and save the excel file into a tmp file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		http.Error(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		http.Error(w, "Failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Optional download url
	downloadUrl := r.FormValue("downloadUrl")

	// Parse and persist file data into database
	err = service.ParseExcelFile(r.Context(), tmpFile.Name(), fileHeader.Filename, downloadUrl)
	if err != nil {
		http.Error(w, "Failed to process uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded and processed successfully"))
}

func handleSyncRequest(w http.ResponseWriter, r *http.Request, service *service.ExcelService) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Get().Excel.ScraperTimeout)
	defer cancel()

	err := service.SearchNewestExcel(ctx)
	if err != nil {
		http.Error(w, "Cannot parse excel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync succeeded"))
}

func isValidRequest(authHeader, expectedKey string) bool {
	if authHeader == "" {
		return false
	}

	authHeader = strings.TrimSpace(authHeader)
	expected := "Bearer " + expectedKey

	return authHeader == expected
}
