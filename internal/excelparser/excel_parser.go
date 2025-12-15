package parser

import (
	"os"
	"strings"
	"sync"

	"github.com/elias-gill/poliplanner2/internal/excelparser/dto"
	"github.com/elias-gill/poliplanner2/internal/excelparser/exceptions"
	"github.com/tealeg/xlsx/v3"
)

var dtoPool = sync.Pool{
	New: func() any { return new(dto.SubjectDTO) },
}

type ExcelParser struct {
	layouts        []Layout
	file           *xlsx.File
	sheets         []string
	currentSheet   int
	headerKeywords map[string]bool
	fieldSetters   map[string]func(*dto.SubjectDTO, string)
}

type ParsingResult struct {
	Career   string
	Subjects []dto.SubjectDTO
}

func NewExcelParser(layoutsDir string, file string) (*ExcelParser, error) {
	loader := NewJsonLayoutLoader(layoutsDir)
	layouts, err := loader.LoadJsonLayouts()
	if err != nil {
		return nil, exceptions.NewExcelParserConfigurationException("Failed to load layouts", err)
	}
	setters := buildFieldSetters()

	p := &ExcelParser{
		layouts: layouts,
		headerKeywords: map[string]bool{
			"item": true,
			"ítem": true,
		},
		currentSheet: -1,
		fieldSetters: setters,
	}

	return p, p.prepareParser(file)
}

func buildFieldSetters() map[string]func(*dto.SubjectDTO, string) {
	return map[string]func(*dto.SubjectDTO, string){
		"departamento": func(d *dto.SubjectDTO, v string) { d.SetDepartment(v) },
		"asignatura":   func(d *dto.SubjectDTO, v string) { d.SetSubjectName(v) },
		"nivel":        func(d *dto.SubjectDTO, v string) { d.SetSemester(v) },
		"semestre":     func(d *dto.SubjectDTO, v string) { d.SetSemester(v) },
		"seccion":      func(d *dto.SubjectDTO, v string) { d.SetSection(v) },
		"titulo":       func(d *dto.SubjectDTO, v string) { d.SetTeacherTitle(v) },
		"apellido":     func(d *dto.SubjectDTO, v string) { d.SetTeacherLastName(v) },
		"nombre":       func(d *dto.SubjectDTO, v string) { d.SetTeacherName(v) },
		"correo":       func(d *dto.SubjectDTO, v string) { d.SetTeacherEmail(v) },
		"diaParcial1":  func(d *dto.SubjectDTO, v string) { d.SetPartial1Date(v) },
		"horaParcial1": func(d *dto.SubjectDTO, v string) { d.SetPartial1Time(v) },
		"aulaParcial1": func(d *dto.SubjectDTO, v string) { d.SetPartial1Room(v) },
		"diaParcial2":  func(d *dto.SubjectDTO, v string) { d.SetPartial2Date(v) },
		"horaParcial2": func(d *dto.SubjectDTO, v string) { d.SetPartial2Time(v) },
		"aulaParcial2": func(d *dto.SubjectDTO, v string) { d.SetPartial2Room(v) },
		"diaFinal1":    func(d *dto.SubjectDTO, v string) { d.SetFinal1Date(v) },
		"horaFinal1":   func(d *dto.SubjectDTO, v string) { d.SetFinal1Time(v) },
		"aulaFinal1":   func(d *dto.SubjectDTO, v string) { d.SetFinal1Room(v) },
		"diaFinal2":    func(d *dto.SubjectDTO, v string) { d.SetFinal2Date(v) },
		"horaFinal2":   func(d *dto.SubjectDTO, v string) { d.SetFinal2Time(v) },
		"aulaFinal2":   func(d *dto.SubjectDTO, v string) { d.SetFinal2Room(v) },
		"revisionDia": func(d *dto.SubjectDTO, v string) {
			d.SetFinal1RevDate(v)
			d.SetFinal2RevDate(v)
		},
		"revisionHora": func(d *dto.SubjectDTO, v string) {
			d.SetFinal1RevTime(v)
			d.SetFinal2RevTime(v)
		},
		"mesaPresidente": func(d *dto.SubjectDTO, v string) { d.SetCommitteePresident(v) },
		"mesaMiembro1":   func(d *dto.SubjectDTO, v string) { d.SetCommitteeMember1(v) },
		"mesaMiembro2":   func(d *dto.SubjectDTO, v string) { d.SetCommitteeMember2(v) },
		"aulaLunes":      func(d *dto.SubjectDTO, v string) { d.SetMondayRoom(v) },
		"horaLunes":      func(d *dto.SubjectDTO, v string) { d.SetMonday(v) },
		"aulaMartes":     func(d *dto.SubjectDTO, v string) { d.SetTuesdayRoom(v) },
		"horaMartes":     func(d *dto.SubjectDTO, v string) { d.SetTuesday(v) },
		"aulaMiercoles":  func(d *dto.SubjectDTO, v string) { d.SetWednesdayRoom(v) },
		"horaMiercoles":  func(d *dto.SubjectDTO, v string) { d.SetWednesday(v) },
		"aulaJueves":     func(d *dto.SubjectDTO, v string) { d.SetThursdayRoom(v) },
		"horaJueves":     func(d *dto.SubjectDTO, v string) { d.SetThursday(v) },
		"aulaViernes":    func(d *dto.SubjectDTO, v string) { d.SetFridayRoom(v) },
		"horaViernes":    func(d *dto.SubjectDTO, v string) { d.SetFriday(v) },
		"aulaSabado":     func(d *dto.SubjectDTO, v string) { d.SetSaturdayRoom(v) },
		"horaSabado":     func(d *dto.SubjectDTO, v string) { d.SetSaturday(v) },
		"fechasSabado":   func(d *dto.SubjectDTO, v string) { d.SetSaturdayDates(v) },
	}
}

func (ep *ExcelParser) prepareParser(filePath string) error {
	if ep.file != nil {
		ep.file = nil // no Close, GC manages its deallocation
	}
	f, err := xlsx.OpenFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return exceptions.NewExcelParserConfigurationException("File not found: "+filePath, err)
		}
		return exceptions.NewExcelParserInputException("Error reading file: "+filePath, err)
	}
	ep.file = f
	ep.sheets = make([]string, len(f.Sheets))
	for i, s := range f.Sheets {
		ep.sheets[i] = s.Name
	}
	ep.currentSheet = -1
	return nil
}

func (ep *ExcelParser) Close() {
	ep.file = nil // mark for GC
}

func (ep *ExcelParser) NextSheet() bool {
	ep.currentSheet++
	for ep.currentSheet < len(ep.sheets) {
		name := ep.sheets[ep.currentSheet]
		lower := strings.ToLower(name)
		// Ignore this garbage
		if strings.Contains(lower, "odigos") ||
			strings.Contains(lower, "asignaturas") ||
			strings.Contains(lower, "homologadas") ||
			strings.Contains(lower, "homólogas") ||
			lower == "códigos" {
			ep.currentSheet++
			continue
		}
		return true
	}
	return false
}

func (ep *ExcelParser) ParseCurrentSheet() (*ParsingResult, error) {
	if ep.currentSheet < 0 || ep.currentSheet >= len(ep.sheets) {
		return nil, exceptions.NewExcelParserException("No current sheet selected", nil)
	}
	sheetName := ep.sheets[ep.currentSheet]
	subjects, err := ep.parseSheet(sheetName)
	if err != nil {
		return nil, err
	}
	return &ParsingResult{Career: sheetName, Subjects: subjects}, nil
}

func (ep *ExcelParser) parseSheet(sheetName string) ([]dto.SubjectDTO, error) {
	sh, ok := ep.file.Sheet[sheetName]
	if !ok {
		return nil, exceptions.NewExcelParserInputException("Sheet not found: "+sheetName, nil)
	}
	subjects := make([]dto.SubjectDTO, 0, 250)
	var lowerHeader []string
	var layout *Layout
	var startingCell int
	rowIdx := 0
	maxRow := sh.MaxRow

	// Helper to collect cells into slice
	collectCells := func() ([]*xlsx.Cell, bool) {
		row, _ := sh.Row(rowIdx)
		if row == nil {
			return nil, false
		}
		cells := make([]*xlsx.Cell, 0, 50)
		err := row.ForEachCell(func(c *xlsx.Cell) error {
			cells = append(cells, c)
			return nil
		})
		if err != nil {
			return nil, false
		}
		return cells, true
	}

	for rowIdx < maxRow {
		cells, ok := collectCells()
		if !ok || len(cells) == 0 {
			rowIdx++
			continue
		}
		if layout == nil {
			if ep.isHeaderRow(cells) {
				lowerHeader = ep.buildLowerHeader(cells)
				startingCell = ep.calculateStartingCell(cells)
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
		if ep.isEmptyRow(cells) {
			break
		}
		d := dtoPool.Get().(*dto.SubjectDTO)
		*d = dto.SubjectDTO{}
		current := startingCell - 1
		for _, field := range layout.Headers {
			current++
			if current >= len(cells) {
				break
			}
			val := strings.TrimSpace(cells[current].Value)
			if val == "" {
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

func (ep *ExcelParser) buildLowerHeader(cells []*xlsx.Cell) []string {
	lower := make([]string, len(cells))
	for i, c := range cells {
		lower[i] = strings.ToLower(strings.TrimSpace(c.Value))
	}
	return lower
}

func (ep *ExcelParser) isHeaderRow(cells []*xlsx.Cell) bool {
	for _, c := range cells {
		if c.Value != "" {
			if ep.headerKeywords[strings.ToLower(strings.TrimSpace(c.Value))] {
				return true
			}
		}
	}
	return false
}

func (ep *ExcelParser) isEmptyRow(cells []*xlsx.Cell) bool {
	for _, c := range cells {
		if strings.TrimSpace(c.Value) != "" {
			return false
		}
	}
	return true
}

func (ep *ExcelParser) calculateStartingCell(cells []*xlsx.Cell) int {
	for i, c := range cells {
		if strings.TrimSpace(c.Value) != "" {
			return i
		}
	}
	return 0
}

func (ep *ExcelParser) findFittingLayout(lowerHeader []string) (*Layout, error) {
	for i := range ep.layouts {
		if ep.layoutMatches(&ep.layouts[i], lowerHeader) {
			return &ep.layouts[i], nil
		}
	}
	return nil, exceptions.NewLayoutMatchException("No matching layout found")
}

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
