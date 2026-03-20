package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/elias-gill/poliplanner2/internal/config"
	log "github.com/elias-gill/poliplanner2/logger"
)

type Layout struct {
	FileName string
	Headers  []string
	Patterns map[string][]string
}

type JsonLayoutLoader struct {
	layoutsDir string
}

func NewJsonLayoutLoader() *JsonLayoutLoader {
	layoutsDir := path.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layout", "layouts")
	return &JsonLayoutLoader{
		layoutsDir: layoutsDir,
	}
}

type jsonLayoutFile struct {
	List []struct {
		Header   string   `json:"encabezado"`
		Patterns []string `json:"patron"`
	} `json:"lista"`
}

func (l *JsonLayoutLoader) LoadJsonLayouts() ([]Layout, error) {
	files, err := filepath.Glob(filepath.Join(l.layoutsDir, "*.json"))
	if err != nil {
		log.Error("Failed to read layouts directory", "dir", l.layoutsDir, "error", err)
		return nil, fmt.Errorf("error reading layout directory: %w", err)
	}

	if len(files) == 0 {
		log.Error("No layout files found", "dir", l.layoutsDir)
		return nil, fmt.Errorf("no JSON files found in: %s", l.layoutsDir)
	}

	var layouts []Layout
	var errorsCount int

	for _, file := range files {
		layout, err := l.loadSingleLayout(file)
		if err != nil {
			log.Warn("Skipping invalid layout file", "file", filepath.Base(file), "error", err)
			errorsCount++
			continue
		}
		layouts = append(layouts, *layout)
	}

	if len(layouts) == 0 {
		log.Error("No valid layouts could be loaded", "files_checked", len(files), "errors", errorsCount)
		return nil, fmt.Errorf("no valid layouts could be loaded from %d files", len(files))
	}

	log.Info("Layouts loaded successfully",
		"count", len(layouts),
		"files_checked", len(files),
		"invalid_files", errorsCount,
	)

	return layouts, nil
}

func (l *JsonLayoutLoader) loadSingleLayout(filePath string) (*Layout, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", filepath.Base(filePath), err)
	}

	var jsonData jsonLayoutFile
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON format in %s: %w", filepath.Base(filePath), err)
	}

	if len(jsonData.List) == 0 {
		return nil, fmt.Errorf("empty 'lista' array in %s", filepath.Base(filePath))
	}

	headers := make([]string, 0, len(jsonData.List))
	patterns := make(map[string][]string)

	for i, entry := range jsonData.List {
		if entry.Header == "" {
			return nil, fmt.Errorf("empty header found at index %d in %s", i, filepath.Base(filePath))
		}
		headers = append(headers, entry.Header)
		if len(entry.Patterns) > 0 {
			patterns[entry.Header] = entry.Patterns
		}
	}

	return &Layout{
		FileName: filepath.Base(filePath),
		Headers:  headers,
		Patterns: patterns,
	}, nil
}
