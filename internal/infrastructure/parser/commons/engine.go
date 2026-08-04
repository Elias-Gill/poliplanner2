package commons

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/exceptions"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/layout"
	"github.com/xuri/excelize/v2"
)

type parserType int

const (
	ScheduleParser = iota
	LabsParser
)

type FieldSetter[T any] func(item *T, value string)

type BaseExcelEngine[T any] struct {
	Layouts        []layout.Layout
	File           *excelize.File
	SheetNames     []string
	CurrentSheet   int
	HeaderKeywords []string
	FieldSetters   map[string]FieldSetter[T]
}

func NewBaseEngine[T any](file io.ReadCloser, t parserType, headerKeywords []string, setters map[string]FieldSetter[T]) (*BaseExcelEngine[T], error) {
	path := filepath.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layout", "schedules")
	if t == LabsParser {
		path = filepath.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layout", "labs")
	}

	loader := layout.NewJsonLayoutLoader(path)
	layouts, err := loader.LoadJsonLayouts()
	if err != nil {
		return nil, exceptions.NewExcelParserConfigurationException("Failed to load layouts", err)
	}

	engine := &BaseExcelEngine[T]{
		Layouts:        layouts,
		HeaderKeywords: headerKeywords,
		CurrentSheet:   -1,
		FieldSetters:   setters,
	}

	if err := engine.PrepareFile(file); err != nil {
		return nil, err
	}
	return engine, nil
}

func (e *BaseExcelEngine[T]) PrepareFile(file io.ReadCloser) error {
	if e.File != nil {
		e.Close()
	}

	f, err := excelize.OpenReader(file, excelize.Options{
		UnzipSizeLimit:    25 << 20,
		UnzipXMLSizeLimit: 8 << 20,
	})
	if err != nil {
		if os.IsNotExist(err) {
			return exceptions.NewExcelParserConfigurationException("Cannot read source", err)
		}
		return exceptions.NewExcelParserInputException("Error reading source: ", err)
	}

	e.File = f
	e.SheetNames = f.GetSheetList()
	e.CurrentSheet = -1
	return nil
}

func (e *BaseExcelEngine[T]) Close() {
	if e.File != nil {
		e.File.Close()
		e.File = nil
	}
}

func (e *BaseExcelEngine[T]) NextSheet(shouldParseFn func(string) bool) bool {
	e.CurrentSheet++
	for e.CurrentSheet < len(e.SheetNames) {
		name := e.SheetNames[e.CurrentSheet]
		if !shouldParseFn(name) {
			e.CurrentSheet++
			continue
		}
		return true
	}
	return false
}

func (e *BaseExcelEngine[T]) ParseSheetStream(sheetName string, acquireItem func() *T, releaseItem func(*T)) ([]T, error) {
	items := make([]T, 0, 250)

	stream, err := e.File.Rows(sheetName)
	if err != nil {
		return nil, exceptions.NewExcelParserInputException("Sheet not found: "+sheetName, err)
	}
	defer stream.Close()

	var lowerHeader []string
	var lay *layout.Layout
	var startingCell int

	for stream.Next() {
		row, err := stream.Columns()
		if err != nil {
			return nil, exceptions.NewExcelParserInputException("Error reading row", err)
		}

		if len(row) == 0 || e.IsEmptyRow(row) {
			continue
		}

		if lay == nil {
			if e.IsHeaderRow(row) {
				lowerHeader = e.BuildLowerHeader(row)
				startingCell = e.CalculateStartingCell(row)
				l, err := e.FindFittingLayout(lowerHeader)
				if err != nil {
					return nil, err
				}
				lay = l
			}
			continue
		}

		if e.IsEmptyRow(row) {
			break
		}

		item := acquireItem()
		current := startingCell - 1

		for _, field := range lay.Headers {
			current++
			if current >= len(row) {
				break
			}
			val := row[current]
			if len(val) == 0 {
				continue
			}
			if setter, ok := e.FieldSetters[field]; ok {
				setter(item, val)
			}
		}

		items = append(items, *item)
		releaseItem(item)
	}

	if lay == nil {
		return nil, exceptions.NewLayoutMatchException("No header row found in sheet: " + sheetName)
	}
	return items, nil
}

func (e *BaseExcelEngine[T]) FindFittingLayout(lowerHeader []string) (*layout.Layout, error) {
	for i := range e.Layouts {
		if e.LayoutMatches(&e.Layouts[i], lowerHeader) {
			return &e.Layouts[i], nil
		}
	}
	return nil, exceptions.NewLayoutMatchException("No matching layout found for sheet")
}

func (e *BaseExcelEngine[T]) LayoutMatches(l *layout.Layout, lower []string) bool {
	cellIdx, hdrIdx := 0, 0
	for hdrIdx < len(l.Headers) && cellIdx < len(lower) {
		val := lower[cellIdx]
		cellIdx++
		if val == "" {
			continue
		}
		patterns, ok := l.Patterns[l.Headers[hdrIdx]]
		if !ok {
			return false
		}
		match := false
		for _, p := range patterns {
			if strings.Contains(val, p) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
		hdrIdx++
	}
	return hdrIdx == len(l.Headers)
}

func (e *BaseExcelEngine[T]) IsHeaderRow(row []string) bool {
	for _, val := range row {
		trimmed := strings.TrimSpace(val)
		if len(trimmed) == 0 {
			continue
		}
		lowerVal := strings.ToLower(trimmed)
		for _, keyword := range e.HeaderKeywords {
			if strings.Contains(lowerVal, keyword) {
				return true
			}
		}
	}
	return false
}

func (e *BaseExcelEngine[T]) IsEmptyRow(row []string) bool {
	for _, val := range row {
		if len(strings.TrimSpace(val)) != 0 {
			return false
		}
	}
	return true
}

func (e *BaseExcelEngine[T]) BuildLowerHeader(row []string) []string {
	lower := make([]string, len(row))
	for i, val := range row {
		lower[i] = strings.ToLower(strings.TrimSpace(val))
	}
	return lower
}

func (e *BaseExcelEngine[T]) CalculateStartingCell(row []string) int {
	for i, val := range row {
		if len(val) > 0 {
			for _, r := range val {
				if !unicode.IsSpace(r) {
					return i
				}
			}
		}
	}
	return 0
}
