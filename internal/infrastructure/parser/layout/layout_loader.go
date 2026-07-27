package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	layoutsDir := filepath.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layout", "layouts")
	return &JsonLayoutLoader{layoutsDir: layoutsDir}
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
		return nil, fmt.Errorf("error reading layout directory: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no JSON files found in: %s", l.layoutsDir)
	}

	var layouts []Layout
	for _, file := range files {
		layout, err := l.loadSingleLayout(file)
		if err != nil {
			log.Warn("Skipping invalid layout file", "file", filepath.Base(file), "error", err)
			continue
		}
		layouts = append(layouts, *layout)
	}
	return layouts, nil
}

func (l *JsonLayoutLoader) loadSingleLayout(filePath string) (*Layout, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var jsonData jsonLayoutFile
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	headers := make([]string, 0, len(jsonData.List))
	patterns := make(map[string][]string)

	for _, entry := range jsonData.List {
		if entry.Header == "" {
			return nil, fmt.Errorf("empty header in %s", filepath.Base(filePath))
		}
		headers = append(headers, entry.Header)
		if len(entry.Patterns) > 0 {
			// Guardamos los patrones ya en minúsculas de forma permanente
			lowered := make([]string, len(entry.Patterns))
			for i, p := range entry.Patterns {
				lowered[i] = strings.ToLower(p)
			}
			patterns[entry.Header] = lowered
		}
	}

	return &Layout{
		FileName: filepath.Base(filePath),
		Headers:  headers,
		Patterns: patterns,
	}, nil
}
