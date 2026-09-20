package engine

import (
	"io"
	"os"
	"runtime"
	"strings"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/exceptions"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/xuri/excelize/v2"
)

type ParserEngine struct {
	layouts        []Layout
	file           *excelize.File
	sheetNames     []string
	currentSheet   int
	headerKeywords []string

	SheetFilter func(string) bool
}

func NewParser(file io.ReadCloser, layoutsDir string) (*ParserEngine, error) {
	loader := NewJsonLayoutLoader(layoutsDir)
	layouts, err := loader.LoadJsonLayouts()
	if err != nil {
		return nil, exceptions.NewExcelParserConfigurationException("Failed to load layouts", err)
	}

	p := &ParserEngine{
		layouts:        layouts,
		headerKeywords: []string{"item", "ítem", "DPTO.", "dpto"},
		currentSheet:   -1,
	}

	memUsageStatus("Excel parser loading", func() {
		err = p.prepareParser(file)
	})

	if err != nil {
		return nil, err
	}
	return p, nil
}

func (ep *ParserEngine) Close() {
	if ep.file != nil {
		ep.file.Close()
		ep.file = nil
	}
}

func (ep *ParserEngine) NextSheet() (string, bool) {
	ep.currentSheet++

	for ep.currentSheet < len(ep.sheetNames) {
		name := ep.sheetNames[ep.currentSheet]
		shouldParse := ep.SheetFilter(name)
		if shouldParse {
			return name, true
		}
		ep.currentSheet++
	}

	return "", false
}

// ParseCurrentSheet iterates over rows and delegates processing to rowHandler.
// It detects the header row and matching layout automatically.
func (ep *ParserEngine) ParseCurrentSheet(sheetName string, rowHandler func(row []string, lay *Layout, startingCell int) error) error {
	stream, err := ep.file.Rows(sheetName)
	if err != nil {
		return exceptions.NewExcelParserInputException("Sheet not found: "+sheetName, err)
	}
	defer stream.Close()

	var lay *Layout
	var startingCell int

	for stream.Next() {
		row, err := stream.Columns()
		if err != nil {
			return exceptions.NewExcelParserInputException("Error reading row", err)
		}

		if len(row) == 0 || ep.isEmptyRow(row) {
			continue
		}

		if lay == nil {
			if ep.isHeaderRow(row) {
				lowerHeader := ep.buildLowerHeader(row)
				startingCell = ep.calculateStartingCell(row)
				lay, err = ep.findFittingLayout(lowerHeader)
				if err != nil {
					return err
				}
			}
			continue
		}

		if ep.isEmptyRow(row) {
			break
		}

		if err := rowHandler(row, lay, startingCell); err != nil {
			return err
		}
	}

	if lay == nil {
		return exceptions.NewLayoutMatchException("No header row found in sheet: " + sheetName)
	}
	return nil
}

// ======================= LAYOUT & ROW UTILS ===============================

func (ep *ParserEngine) findFittingLayout(lowerHeader []string) (*Layout, error) {
	for i := range ep.layouts {
		if ep.layoutMatches(&ep.layouts[i], lowerHeader) {
			return &ep.layouts[i], nil
		}
	}
	return nil, exceptions.NewLayoutMatchException("No matching layout found for sheet")
}

func (ep *ParserEngine) layoutMatches(l *Layout, lower []string) bool {
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

func (ep *ParserEngine) isHeaderRow(row []string) bool {
	for _, val := range row {
		trimmed := strings.TrimSpace(val)
		if len(trimmed) == 0 {
			continue
		}
		lowerVal := strings.ToLower(trimmed)
		for _, keyword := range ep.headerKeywords {
			if strings.Contains(lowerVal, keyword) {
				return true
			}
		}
	}
	return false
}

func (ep *ParserEngine) isEmptyRow(row []string) bool {
	for _, val := range row {
		if len(strings.TrimSpace(val)) != 0 {
			return false
		}
	}
	return true
}

func (ep *ParserEngine) buildLowerHeader(row []string) []string {
	lower := make([]string, len(row))
	for i, val := range row {
		lower[i] = strings.ToLower(strings.TrimSpace(val))
	}
	return lower
}

func (ep *ParserEngine) calculateStartingCell(row []string) int {
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

// ======================= ENGINE UTILS ===============================

func (ep *ParserEngine) prepareParser(file io.ReadCloser) error {
	if ep.file != nil {
		ep.Close()
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

	ep.file = f
	ep.sheetNames = f.GetSheetList()
	ep.currentSheet = -1
	return nil
}

func memUsageStatus(label string, do func()) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	do()

	runtime.ReadMemStats(&m2)

	allocKB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024
	allocMB := allocKB / 1024

	logger.Debug(label+" - Alloc delta",
		"KB", allocKB,
		"MB", allocMB,
	)

	heapKB := float64(m2.HeapAlloc-m1.HeapAlloc) / 1024
	heapMB := heapKB / 1024

	logger.Debug(label+" - Heap delta",
		"KB", heapKB,
		"MB", heapMB,
	)
}
