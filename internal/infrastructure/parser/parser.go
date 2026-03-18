package parser

import (
	"io"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/parser/exceptions"
	"github.com/elias-gill/poliplanner2/internal/parser/layout"
	"github.com/elias-gill/poliplanner2/internal/utils"
	"github.com/xuri/excelize/v2"
)
// For reusing SubjectDTO objects to reduce allocations
var dtoPool = sync.Pool{
	New: func() any {
		d := new(SubjectDTO)
		d.Reset() // Ensure clean state
		return d
	},
}

// ExcelParser handles parsing of Excel files containing subject schedules
type ExcelParser struct {
	layouts        []layout.Layout
	file           *excelize.File
	sheetNames     []string
	currentSheet   int
	headerKeywords []string // Keywords to identify header rows
	fieldSetters   map[string]func(*SubjectDTO, string)
}

// ParsedSheet contains the parsed subjects for a specific career/sheet
type ParsedSheet struct {
	Name     string
	Subjects []SubjectDTO
}

// NewExcelParser creates a new Excel parser instance
func NewExcelParser(file io.ReadCloser) (*ExcelParser, error) {
	// Loads the possible known layouts of the excel file
	loader := layout.NewJsonLayoutLoader()
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
		err = p.prepareParser(file)
	})

	if err != nil {
		return nil, err
	}

	return p, nil
}

// ================================
// =        Public API            =
// ================================

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

// Parses the currently selected sheet
func (ep *ExcelParser) ParseCurrentSheet() (*ParsedSheet, error) {
	if ep.currentSheet < 0 || ep.currentSheet >= len(ep.sheetNames) {
		return nil, exceptions.NewExcelParserException("No current sheet selected", nil)
	}
	sheetName := ep.sheetNames[ep.currentSheet]
	subjects, err := ep.parseSheet(sheetName)
	if err != nil {
		return nil, err
	}
	return &ParsedSheet{Name: strings.ToUpper(sheetName), Subjects: subjects}, nil
}

// =====================================
// =        Private methods            =
// =====================================

func (ep *ExcelParser) prepareParser(file io.ReadCloser) error {
	if ep.file != nil {
		ep.Close()
	}

	// Use optimized options for better performance
	f, err := excelize.OpenReader(file, excelize.Options{
		// Limit memory usage by restricting unzip size
		UnzipSizeLimit: 8 << 20, // 12MB limit
		// Skip loading cell styles we don't need
		UnzipXMLSizeLimit: 16 << 20, // 32MB per XML file
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

func (ep *ExcelParser) parseSheet(sheetName string) ([]SubjectDTO, error) {
	subjects := make([]SubjectDTO, 0, 250)

	// Use streaming API for better memory efficiency
	stream, err := ep.file.Rows(sheetName)
	if err != nil {
		return nil, exceptions.NewExcelParserInputException("Sheet not found: "+sheetName, err)
	}
	defer stream.Close()

	var lowerHeader []string
	var layout *layout.Layout
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
		d := dtoPool.Get().(*SubjectDTO)
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

// Finds a layout that matches the header row
func (ep *ExcelParser) findFittingLayout(lowerHeader []string) (*layout.Layout, error) {
	for i := range ep.layouts {
		if ep.layoutMatches(&ep.layouts[i], lowerHeader) {
			return &ep.layouts[i], nil
		}
	}
	return nil, exceptions.NewLayoutMatchException("No matching layout found")
}

// Check if a layout matches a given header row
func (ep *ExcelParser) layoutMatches(layout *layout.Layout, lower []string) bool {
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

func (ep *ExcelParser) buildLowerHeader(row []string) []string {
	lower := make([]string, len(row))
	for i, val := range row {
		lower[i] = strings.ToLower(strings.TrimSpace(val))
	}
	return lower
}

// ================================
// =           Helpers            =
// ================================

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
		if len(strings.TrimSpace(val)) != 0 {
			return false
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

// buildFieldSetters creates a map of field setters for SubjectDTO
func buildFieldSetters() map[string]func(*SubjectDTO, string) {
	return map[string]func(*SubjectDTO, string){
		"departamento":       func(d *SubjectDTO, v string) { d.SetDepartment(v) },
		"asignatura":         func(d *SubjectDTO, v string) { d.SetSubjectName(v) },
		"nivel":              func(d *SubjectDTO, v string) { d.SetLevel(v) },
		"semestre":           func(d *SubjectDTO, v string) { d.SetSemester(v) },
		"seccion":            func(d *SubjectDTO, v string) { d.SetSection(v) },
		"titulo":             func(d *SubjectDTO, v string) { d.SetTeachersTitles(v) },
		"apellido":           func(d *SubjectDTO, v string) { d.SetTeachersLastNames(v) },
		"nombre":             func(d *SubjectDTO, v string) { d.SetTeachersFirtNames(v) },
		"correo":             func(d *SubjectDTO, v string) { d.SetTeachersEmails(v) },
		"diaParcial1":        func(d *SubjectDTO, v string) { d.SetPartial1Date(v) },
		"horaParcial1":       func(d *SubjectDTO, v string) { d.SetPartial1Time(v) },
		"aulaParcial1":       func(d *SubjectDTO, v string) { d.SetPartial1Room(v) },
		"diaParcial2":        func(d *SubjectDTO, v string) { d.SetPartial2Date(v) },
		"horaParcial2":       func(d *SubjectDTO, v string) { d.SetPartial2Time(v) },
		"aulaParcial2":       func(d *SubjectDTO, v string) { d.SetPartial2Room(v) },
		"diaFinal1":          func(d *SubjectDTO, v string) { d.SetFinal1Date(v) },
		"horaFinal1":         func(d *SubjectDTO, v string) { d.SetFinal1Time(v) },
		"aulaFinal1":         func(d *SubjectDTO, v string) { d.SetFinal1Room(v) },
		"diaFinal2":          func(d *SubjectDTO, v string) { d.SetFinal2Date(v) },
		"horaFinal2":         func(d *SubjectDTO, v string) { d.SetFinal2Time(v) },
		"aulaFinal2":         func(d *SubjectDTO, v string) { d.SetFinal2Room(v) },
		"revisionFinal1Dia":  func(d *SubjectDTO, v string) { d.SetFinal1RevDate(v) },
		"revisionFinal2Dia":  func(d *SubjectDTO, v string) { d.SetFinal2RevDate(v) },
		"revisionFinal1Hora": func(d *SubjectDTO, v string) { d.SetFinal1RevTime(v) },
		"revisionFinal2Hora": func(d *SubjectDTO, v string) { d.SetFinal2RevTime(v) },
		"mesaPresidente":     func(d *SubjectDTO, v string) { d.SetCommitteePresident(v) },
		"mesaMiembro1":       func(d *SubjectDTO, v string) { d.SetCommitteeMember1(v) },
		"mesaMiembro2":       func(d *SubjectDTO, v string) { d.SetCommitteeMember2(v) },
		"aulaLunes":          func(d *SubjectDTO, v string) { d.SetDayRoom(Monday, v) },
		"horaLunes":          func(d *SubjectDTO, v string) { d.SetDayTime(Monday, v) },
		"aulaMartes":         func(d *SubjectDTO, v string) { d.SetDayRoom(Tuesday, v) },
		"horaMartes":         func(d *SubjectDTO, v string) { d.SetDayTime(Tuesday, v) },
		"aulaMiercoles":      func(d *SubjectDTO, v string) { d.SetDayRoom(Wednesday, v) },
		"horaMiercoles":      func(d *SubjectDTO, v string) { d.SetDayTime(Wednesday, v) },
		"aulaJueves":         func(d *SubjectDTO, v string) { d.SetDayRoom(Thursday, v) },
		"horaJueves":         func(d *SubjectDTO, v string) { d.SetDayTime(Thursday, v) },
		"aulaViernes":        func(d *SubjectDTO, v string) { d.SetDayRoom(Friday, v) },
		"horaViernes":        func(d *SubjectDTO, v string) { d.SetDayTime(Friday, v) },
		"aulaSabado":         func(d *SubjectDTO, v string) { d.SetDayRoom(Saturday, v) },
		"horaSabado":         func(d *SubjectDTO, v string) { d.SetDayTime(Saturday, v) },
		"fechasSabado":       func(d *SubjectDTO, v string) { d.SetSaturdayDates(v) },
	}
}
