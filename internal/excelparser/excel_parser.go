package parser

import (
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/excelparser/dto"
	"github.com/elias-gill/poliplanner2/internal/excelparser/exceptions"
	"github.com/elias-gill/poliplanner2/internal/utils"
	"github.com/xuri/excelize/v2"
)

// For reusing SubjectDTO objects to reduce allocations
var dtoPool = sync.Pool{
	New: func() any {
		d := new(dto.SubjectDTO)
		d.Reset() // Ensure clean state
		return d
	},
}

// ExcelParser handles parsing of Excel files containing subject schedules
type ExcelParser struct {
	layouts        []Layout
	file           *excelize.File
	sheetNames     []string
	currentSheet   int
	headerKeywords []string // Keywords to identify header rows
	fieldSetters   map[string]func(*dto.SubjectDTO, string)
}

// ParsingResult contains the parsed subjects for a specific career/sheet
type ParsingResult struct {
	Career   string
	Subjects []dto.SubjectDTO
}

// NewExcelParser creates a new Excel parser instance
func NewExcelParser(layoutsDir string, filePath string) (*ExcelParser, error) {
	// Loads the possible known layouts of the excel file
	loader := NewJsonLayoutLoader(layoutsDir)
	layouts, err := loader.LoadJsonLayouts()
	if err != nil {
		return nil, exceptions.NewExcelParserConfigurationException("Failed to load layouts", err)
	}

	setters := buildFieldSetters()

	p := &ExcelParser{
		layouts:        layouts,
		headerKeywords: []string{"item", "ítem"},
		currentSheet:   -1,
		fieldSetters:   setters,
	}

	utils.MemUsageStatus("Excel parser loading", func() {
		err = p.prepareParser(filePath)
	})

	if err != nil {
		return nil, err
	}

	return p, nil
}

// buildFieldSetters creates a map of field setters for SubjectDTO
func buildFieldSetters() map[string]func(*dto.SubjectDTO, string) {
	return map[string]func(*dto.SubjectDTO, string){
		"departamento":       func(d *dto.SubjectDTO, v string) { d.SetDepartment(v) },
		"asignatura":         func(d *dto.SubjectDTO, v string) { d.SetSubjectName(v) },
		"nivel":              func(d *dto.SubjectDTO, v string) { d.SetSemester(v) },
		"semestre":           func(d *dto.SubjectDTO, v string) { d.SetSemester(v) },
		"seccion":            func(d *dto.SubjectDTO, v string) { d.SetSection(v) },
		"titulo":             func(d *dto.SubjectDTO, v string) { d.SetTeacherTitle(v) },
		"apellido":           func(d *dto.SubjectDTO, v string) { d.SetTeacherLastName(v) },
		"nombre":             func(d *dto.SubjectDTO, v string) { d.SetTeacherName(v) },
		"correo":             func(d *dto.SubjectDTO, v string) { d.SetTeacherEmail(v) },
		"diaParcial1":        func(d *dto.SubjectDTO, v string) { d.SetPartial1Date(v) },
		"horaParcial1":       func(d *dto.SubjectDTO, v string) { d.SetPartial1Time(v) },
		"aulaParcial1":       func(d *dto.SubjectDTO, v string) { d.SetPartial1Room(v) },
		"diaParcial2":        func(d *dto.SubjectDTO, v string) { d.SetPartial2Date(v) },
		"horaParcial2":       func(d *dto.SubjectDTO, v string) { d.SetPartial2Time(v) },
		"aulaParcial2":       func(d *dto.SubjectDTO, v string) { d.SetPartial2Room(v) },
		"diaFinal1":          func(d *dto.SubjectDTO, v string) { d.SetFinal1Date(v) },
		"horaFinal1":         func(d *dto.SubjectDTO, v string) { d.SetFinal1Time(v) },
		"aulaFinal1":         func(d *dto.SubjectDTO, v string) { d.SetFinal1Room(v) },
		"diaFinal2":          func(d *dto.SubjectDTO, v string) { d.SetFinal2Date(v) },
		"horaFinal2":         func(d *dto.SubjectDTO, v string) { d.SetFinal2Time(v) },
		"aulaFinal2":         func(d *dto.SubjectDTO, v string) { d.SetFinal2Room(v) },
		"revisionFinal1Dia":  func(d *dto.SubjectDTO, v string) { d.SetFinal1RevDate(v) },
		"revisionFinal2Dia":  func(d *dto.SubjectDTO, v string) { d.SetFinal2RevDate(v) },
		"revisionFinal1Hora": func(d *dto.SubjectDTO, v string) { d.SetFinal1RevTime(v) },
		"revisionFinal2Hora": func(d *dto.SubjectDTO, v string) { d.SetFinal2RevTime(v) },
		"mesaPresidente":     func(d *dto.SubjectDTO, v string) { d.SetCommitteePresident(v) },
		"mesaMiembro1":       func(d *dto.SubjectDTO, v string) { d.SetCommitteeMember1(v) },
		"mesaMiembro2":       func(d *dto.SubjectDTO, v string) { d.SetCommitteeMember2(v) },
		"aulaLunes":          func(d *dto.SubjectDTO, v string) { d.SetMondayRoom(v) },
		"horaLunes":          func(d *dto.SubjectDTO, v string) { d.SetMonday(v) },
		"aulaMartes":         func(d *dto.SubjectDTO, v string) { d.SetTuesdayRoom(v) },
		"horaMartes":         func(d *dto.SubjectDTO, v string) { d.SetTuesday(v) },
		"aulaMiercoles":      func(d *dto.SubjectDTO, v string) { d.SetWednesdayRoom(v) },
		"horaMiercoles":      func(d *dto.SubjectDTO, v string) { d.SetWednesday(v) },
		"aulaJueves":         func(d *dto.SubjectDTO, v string) { d.SetThursdayRoom(v) },
		"horaJueves":         func(d *dto.SubjectDTO, v string) { d.SetThursday(v) },
		"aulaViernes":        func(d *dto.SubjectDTO, v string) { d.SetFridayRoom(v) },
		"horaViernes":        func(d *dto.SubjectDTO, v string) { d.SetFriday(v) },
		"aulaSabado":         func(d *dto.SubjectDTO, v string) { d.SetSaturdayRoom(v) },
		"horaSabado":         func(d *dto.SubjectDTO, v string) { d.SetSaturday(v) },
		"fechasSabado":       func(d *dto.SubjectDTO, v string) { d.SetSaturdayDates(v) },
	}
}

func (ep *ExcelParser) prepareParser(filePath string) error {
	if ep.file != nil {
		ep.Close()
	}

	// Use optimized options for better performance
	f, err := excelize.OpenFile(filePath, excelize.Options{
		// Limit memory usage by restricting unzip size
		UnzipSizeLimit: 256 << 20, // 256MB limit
		// Skip loading cell styles we don't need
		UnzipXMLSizeLimit: 64 << 20, // 64MB per XML file
	})
	if err != nil {
		if os.IsNotExist(err) {
			return exceptions.NewExcelParserConfigurationException("File not found: "+filePath, err)
		}
		return exceptions.NewExcelParserInputException("Error reading file: "+filePath, err)
	}

	ep.file = f
	ep.sheetNames = f.GetSheetList()
	ep.currentSheet = -1

	return nil
}

// Close releases resources used by the parser
func (ep *ExcelParser) Close() {
	if ep.file != nil {
		ep.file.Close()
		ep.file = nil
	}
}

// NextSheet moves to the next valid sheet for parsing
func (ep *ExcelParser) NextSheet() bool {
	ep.currentSheet++
	for ep.currentSheet < len(ep.sheetNames) {
		name := ep.sheetNames[ep.currentSheet]
		// Fast check for ignored sheets
		if ep.shouldIgnoreSheet(name) {
			ep.currentSheet++
			continue
		}
		return true
	}
	return false
}

func (ep *ExcelParser) shouldIgnoreSheet(name string) bool {
	// Quick length check first
	if len(name) == 0 {
		return false
	}

	// Common ignored sheet names
	if name == "Códigos" || name == "códigos" {
		return true
	}

	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "odigos") ||
		strings.Contains(lowerName, "asignaturas") ||
		strings.Contains(lowerName, "homologadas") ||
		strings.Contains(lowerName, "homólogas") {
		return true
	}

	return false
}

// Parses the currently selected sheet
func (ep *ExcelParser) ParseCurrentSheet() (*ParsingResult, error) {
	if ep.currentSheet < 0 || ep.currentSheet >= len(ep.sheetNames) {
		return nil, exceptions.NewExcelParserException("No current sheet selected", nil)
	}
	sheetName := ep.sheetNames[ep.currentSheet]
	subjects, err := ep.parseSheet(sheetName)
	if err != nil {
		return nil, err
	}
	return &ParsingResult{Career: sheetName, Subjects: subjects}, nil
}

func (ep *ExcelParser) parseSheet(sheetName string) ([]dto.SubjectDTO, error) {
	subjects := make([]dto.SubjectDTO, 0, 250)

	// Use streaming API for better memory efficiency
	stream, err := ep.file.Rows(sheetName)
	if err != nil {
		return nil, exceptions.NewExcelParserInputException("Sheet not found: "+sheetName, err)
	}
	defer stream.Close()

	var lowerHeader []string
	var layout *Layout
	var startingCell int
	rowIdx := 0

	for stream.Next() {
		row, err := stream.Columns()
		if err != nil {
			return nil, exceptions.NewExcelParserInputException("Error reading row", err)
		}

		// Skip completely empty rows early
		if len(row) == 0 || ep.isEmptyRow(row) {
			rowIdx++
			continue
		}

		if layout == nil {
			if ep.isHeaderRow(row) {
				lowerHeader = ep.buildLowerHeader(row)
				startingCell = ep.calculateStartingCell(row)
				l, err := ep.findFittingLayout(lowerHeader)
				if err != nil {
					return nil, err
				}
				layout = l
				rowIdx++
				continue
			}
			rowIdx++
			continue
		}

		// Stop on empty row
		if ep.isEmptyRow(row) {
			break
		}

		// Parse data row
		d := dtoPool.Get().(*dto.SubjectDTO)
		d.Reset()
		current := startingCell - 1

		for _, field := range layout.Headers {
			current++
			if current >= len(row) {
				break
			}
			val := row[current]
			if len(val) == 0 {
				continue
			}
			// Fast trim inline (only if needed)
			if val[0] == ' ' || val[len(val)-1] == ' ' {
				val = strings.TrimSpace(val)
				if len(val) == 0 {
					continue
				}
			}
			if setter, ok := ep.fieldSetters[field]; ok {
				setter(d, val)
			}
		}
		subjects = append(subjects, *d)
		dtoPool.Put(d)
		rowIdx++
	}

	if layout == nil {
		return nil, exceptions.NewLayoutMatchException("No header row found in sheet: " + sheetName)
	}

	return subjects, nil
}

func (ep *ExcelParser) buildLowerHeader(row []string) []string {
	lower := make([]string, len(row))
	for i, val := range row {
		lower[i] = strings.ToLower(strings.TrimSpace(val))
	}
	return lower
}

func (ep *ExcelParser) isHeaderRow(row []string) bool {
	for _, val := range row {
		if len(val) == 0 {
			continue
		}

		// Normalize and check
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

func (ep *ExcelParser) isEmptyRow(row []string) bool {
	for _, val := range row {
		// if any non-whitespace character exists
		for _, r := range val {
			if !unicode.IsSpace(r) {
				return false
			}
		}
	}
	return true
}

func (ep *ExcelParser) calculateStartingCell(row []string) int {
	for i, val := range row {
		if len(val) > 0 {
			// Check if it's not just whitespace
			for _, r := range val {
				if !unicode.IsSpace(r) {
					return i
				}
			}
		}
	}
	return 0
}

// Finds a layout that matches the header row
func (ep *ExcelParser) findFittingLayout(lowerHeader []string) (*Layout, error) {
	for i := range ep.layouts {
		if ep.layoutMatches(&ep.layouts[i], lowerHeader) {
			return &ep.layouts[i], nil
		}
	}
	return nil, exceptions.NewLayoutMatchException("No matching layout found")
}

// Check if a layout matches a given header row
func (ep *ExcelParser) layoutMatches(layout *Layout, lower []string) bool {
	cellIdx := 0
	hdrIdx := 0
	for hdrIdx < len(layout.Headers) && cellIdx < len(lower) {
		val := lower[cellIdx]
		cellIdx++
		if val == "" {
			continue
		}
		patterns, ok := layout.Patterns[layout.Headers[hdrIdx]]
		if !ok {
			return false
		}
		match := false
		for _, p := range patterns {
			if strings.Contains(val, strings.ToLower(p)) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
		hdrIdx++
	}
	return hdrIdx == len(layout.Headers)
}
