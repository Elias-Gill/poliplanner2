package parser

import (
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/engine"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

var dtoPool = sync.Pool{
	New: func() any {
		return new(SubjectDTO)
	},
}

type SchedulesParser struct {
	engine       *engine.ParserEngine
	fieldSetters map[string]func(*SubjectDTO, string)
}

type ParsedSchedules struct {
	Career   string
	Subjects []SubjectDTO
}

func NewScheduleParser(file io.ReadCloser) (*SchedulesParser, error) {
	layoutsDir := filepath.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layouts", "schedules")

	en, err := engine.NewParser(file, layoutsDir)
	if err != nil {
		return nil, err
	}
	en.SheetFilter = shouldParseSheet

	return &SchedulesParser{
		engine:       en,
		fieldSetters: buildFieldSetters(),
	}, nil
}

func (ep *SchedulesParser) Close() {
	ep.engine.Close()
}

func (ep *SchedulesParser) ParseNextSheet() (*ParsedSchedules, error) {
	name, ok := ep.engine.NextSheet()
	if !ok {
		// There is no sheet to parse
		return nil, nil
	}

	subjects := make([]SubjectDTO, 0, 250)

	err := ep.engine.ParseSheet(name, func(row []string, lay *engine.Layout, startingCell int) error {
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
		subjects = append(subjects, *d)
		dtoPool.Put(d)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ParsedSchedules{
		Career:   strings.ToUpper(strings.ReplaceAll(name, " ", "")),
		Subjects: subjects,
	}, nil
}

func shouldParseSheet(name string) bool {
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
