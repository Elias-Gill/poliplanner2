package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/elias-gill/poliplanner2/internal/logger"
)

type Layout struct {
	FileName string
	Headers  []string
	Patterns map[string][]string
}

type JsonLayoutLoader struct {
	layoutsDir string
}

func NewJsonLayoutLoader(layoutsDir string) *JsonLayoutLoader {
	log.GetLogger().Debug("Creating JSON layout loader", "layouts_dir", layoutsDir)
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
	log.GetLogger().Debug("Loading JSON layouts", "directory", l.layoutsDir)
	var layouts []Layout

	files, err := filepath.Glob(filepath.Join(l.layoutsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("error reading layout directory: %v", err)
	}

	log.GetLogger().Debug("Found JSON files", "count", len(files), "files", files)

	if len(files) == 0 {
		return nil, fmt.Errorf("no JSON files found in: %s", l.layoutsDir)
	}

	loadedCount := 0
	for _, file := range files {
		layout, err := l.loadSingleLayout(file)
		if err != nil {
			log.GetLogger().Warn("Error loading layout file", "file", file, "error", err)
			continue
		}
		layouts = append(layouts, *layout)
		loadedCount++
		log.GetLogger().Debug("Successfully loaded layout", "file", file, "headers_count", len(layout.Headers))
	}

	log.GetLogger().Info("Layout loading completed", "loaded", loadedCount, "total_files", len(files))

	if len(layouts) == 0 {
		return nil, fmt.Errorf("no valid layouts could be loaded")
	}

	return layouts, nil
}

func (l *JsonLayoutLoader) loadSingleLayout(filePath string) (*Layout, error) {
	log.GetLogger().Debug("Loading single layout", "file", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %v", err)
	}

	var jsonData jsonLayoutFile
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}

	if len(jsonData.List) == 0 {
		return nil, fmt.Errorf("empty 'lista' in JSON file")
	}

	headers := make([]string, 0, len(jsonData.List))
	patterns := make(map[string][]string)

	for _, entry := range jsonData.List {
		if entry.Header == "" {
			return nil, fmt.Errorf("empty header in JSON file")
		}
		headers = append(headers, entry.Header)
		if len(entry.Patterns) > 0 {
			patterns[entry.Header] = entry.Patterns
		}
	}

	log.GetLogger().Debug("Layout parsed successfully",
		"file", filepath.Base(filePath),
		"headers_count", len(headers),
		"patterns_count", len(patterns))

	return &Layout{
		FileName: filepath.Base(filePath),
		Headers:  headers,
		Patterns: patterns,
	}, nil
}
