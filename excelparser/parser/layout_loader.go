package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	var layouts []Layout

	files, err := filepath.Glob(filepath.Join(l.layoutsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("error reading layout directory: %v", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no JSON files found in: %s", l.layoutsDir)
	}

	for _, file := range files {
		layout, err := l.loadSingleLayout(file)
		if err != nil {
			fmt.Printf("Error loading layout %s: %v\n", file, err)
			continue
		}
		layouts = append(layouts, *layout)
	}

	if len(layouts) == 0 {
		return nil, fmt.Errorf("no valid layouts could be loaded")
	}

	return layouts, nil
}

func (l *JsonLayoutLoader) loadSingleLayout(filePath string) (*Layout, error) {
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

	return &Layout{
		FileName: filepath.Base(filePath),
		Headers:  headers,
		Patterns: patterns,
	}, nil
}
