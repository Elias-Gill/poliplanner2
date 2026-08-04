package parser

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/exceptions"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/layout"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/xuri/excelize/v2"
)

var dtoPool = sync.Pool{
	New: func() any {
		return new(SubjectDTO)
	},
}

type ExcelParser struct {
	layouts        []layout.Layout
	file           *excelize.File
	sheetNames     []string
	currentSheet   int
	headerKeywords []string
	fieldSetters   map[string]func(*SubjectDTO, string)
}

type ParsedSheet struct {
	Name     string
	Subjects []SubjectDTO
}

func NewParser(file io.ReadCloser) (*ExcelParser, error) {
	// Initialize layout loader
	layoutsDir := filepath.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layout", "schedules")
	loader := layout.NewJsonLayoutLoader(layoutsDir)
	layouts, err := loader.LoadJsonLayouts()
	if err != nil {
		return nil, exceptions.NewExcelParserConfigurationException("Failed to load layouts", err)
	}

	p := &ExcelParser{
		layouts:        layouts,
		headerKeywords: []string{"item", "ítem", "DPTO.", "dpto"},
		currentSheet:   -1,
		fieldSetters:   buildFieldSetters(),
	}

	memUsageStatus("Excel parser loading", func() {
		err = p.prepareParser(file)
	})

	if err != nil {
		return nil, err
	}
	return p, nil
}

func (ep *ExcelParser) Close() {
	if ep.file != nil {
		ep.file.Close()
		ep.file = nil
	}
}

func (ep *ExcelParser) NextSheet() bool {
	ep.currentSheet++
	for ep.currentSheet < len(ep.sheetNames) {
		name := ep.sheetNames[ep.currentSheet]
		if !ep.shouldParseSheet(name) {
			ep.currentSheet++
			continue
		}
		return true
	}
	return false
}

func (ep *ExcelParser) ParseCurrentSheet() (*ParsedSheet, error) {
	if ep.currentSheet < 0 || ep.currentSheet >= len(ep.sheetNames) {
		return nil, exceptions.NewExcelParserException("No current sheet selected", nil)
	}

	sheetName := ep.sheetNames[ep.currentSheet]
	logger.Info("Parsing", "sheet_name", sheetName)

	subjects, err := ep.parseSheet(sheetName)
	return &ParsedSheet{
		Name:     strings.ToUpper(strings.ReplaceAll(sheetName, " ", "")),
		Subjects: subjects,
	}, err
}

func (ep *ExcelParser) prepareParser(file io.ReadCloser) error {
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

func (ep *ExcelParser) parseSheet(sheetName string) ([]SubjectDTO, error) {
	subjects := make([]SubjectDTO, 0, 250)

	stream, err := ep.file.Rows(sheetName)
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

		if len(row) == 0 || ep.isEmptyRow(row) {
			continue
		}

		if lay == nil {
			if ep.isHeaderRow(row) {
				lowerHeader = ep.buildLowerHeader(row)
				startingCell = ep.calculateStartingCell(row)
				l, err := ep.findFittingLayout(lowerHeader)
				if err != nil {
					return nil, err
				}
				lay = l
			}
			continue
		}

		if ep.isEmptyRow(row) {
			break
		}

		d := dtoPool.Get().(*SubjectDTO)
		d.Reset()
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
			if setter, ok := ep.fieldSetters[field]; ok {
				setter(d, val)
			}
		}
		// Slices fijos + Structs planos = Copia segura y aislada por valor en memoria contigua
		subjects = append(subjects, *d)
		dtoPool.Put(d)
	}

	if lay == nil {
		return nil, exceptions.NewLayoutMatchException("No header row found in sheet: " + sheetName)
	}
	return subjects, nil
}

func (ep *ExcelParser) findFittingLayout(lowerHeader []string) (*layout.Layout, error) {
	for i := range ep.layouts {
		if ep.layoutMatches(&ep.layouts[i], lowerHeader) {
			return &ep.layouts[i], nil
		}
	}
	return nil, exceptions.NewLayoutMatchException("No matching layout found for sheet")
}

func (ep *ExcelParser) layoutMatches(l *layout.Layout, lower []string) bool {
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
			if strings.Contains(val, p) { // Eliminada la alocación oculta de strings.ToLower(p)
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

func (ep *ExcelParser) buildLowerHeader(row []string) []string {
	lower := make([]string, len(row))
	for i, val := range row {
		lower[i] = strings.ToLower(strings.TrimSpace(val))
	}
	return lower
}

func (ep *ExcelParser) isHeaderRow(row []string) bool {
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
			for _, r := range val {
				if !unicode.IsSpace(r) {
					return i
				}
			}
		}
	}
	return 0
}

func (ep *ExcelParser) shouldParseSheet(name string) bool {
	if len(name) == 0 {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	valid := map[string]struct{}{
		"iae": {}, "icm": {}, "iek": {}, "iel": {}, "ien": {}, "iin": {}, "imk": {}, "isp": {},
		"lca": {}, "lci": {}, "lcik": {}, "lel": {}, "lgh": {}, "tse": {}, "villarrica": {},
	}
	if _, ok := valid[lower]; ok {
		return true
	}
	return strings.Contains(lower, "oviedo")
}

func buildFieldSetters() map[string]func(*SubjectDTO, string) {
	return map[string]func(*SubjectDTO, string){
		"departamento":       func(d *SubjectDTO, v string) { d.SetDepartment(v) },
		"enfasis":            func(d *SubjectDTO, v string) { d.SetEmphases(v) },
		"plan":               func(d *SubjectDTO, v string) { d.SetPlan(v) },
		"asignatura":         func(d *SubjectDTO, v string) { d.SetSubjectName(v) },
		"nivel":              func(d *SubjectDTO, v string) { d.SetLevel(v) },
		"semestre":           func(d *SubjectDTO, v string) { d.SetSemester(v) },
		"turno":              func(d *SubjectDTO, v string) { d.SetShift(v) },
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
		"aulaLunes":          func(d *SubjectDTO, v string) { d.SetDayRoom(academic.Monday, v) },
		"horaLunes":          func(d *SubjectDTO, v string) { d.SetDayTime(academic.Monday, v) },
		"aulaMartes":         func(d *SubjectDTO, v string) { d.SetDayRoom(academic.Tuesday, v) },
		"horaMartes":         func(d *SubjectDTO, v string) { d.SetDayTime(academic.Tuesday, v) },
		"aulaMiercoles":      func(d *SubjectDTO, v string) { d.SetDayRoom(academic.Wednesday, v) },
		"horaMiercoles":      func(d *SubjectDTO, v string) { d.SetDayTime(academic.Wednesday, v) },
		"aulaJueves":         func(d *SubjectDTO, v string) { d.SetDayRoom(academic.Thursday, v) },
		"horaJueves":         func(d *SubjectDTO, v string) { d.SetDayTime(academic.Thursday, v) },
		"aulaViernes":        func(d *SubjectDTO, v string) { d.SetDayRoom(academic.Friday, v) },
		"horaViernes":        func(d *SubjectDTO, v string) { d.SetDayTime(academic.Friday, v) },
		"aulaSabado":         func(d *SubjectDTO, v string) { d.SetDayRoom(academic.Saturday, v) },
		"horaSabado":         func(d *SubjectDTO, v string) { d.SetDayTime(academic.Saturday, v) },
		"fechasSabado":       func(d *SubjectDTO, v string) { d.SetSaturdayDates(v) },
	}
}
