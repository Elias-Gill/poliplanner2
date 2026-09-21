package parser

import (
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/commons"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/engine"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/logger"
)

// As the parsing of this excel file is a lot messier (hence the strange format of the
// laboratory sections), but at the same time the file is not as big as the normal schedules
// file, then we can (and have) to be a lot more fancy when comparing names, sections and
// storing our results.

// defaultSection identifies a subject that only exposes a single laboratory section. In that
// case the day cell does not carry a "(seccion)" tag.
const defaultSection = "UNICA"

// labKey is the identity of a laboratory offering. A subject may be taught under several
// plans, and each plan may expose several sections (T1, T2, ...). The fields are normalized
// only to detect repeated entities; they are never exposed as the final result.
type labKey struct {
	subject string
	plan    string
	section string
}

type LaboratoriesParser struct {
	engine *engine.ParserEngine
}

type ParsedLaboratories struct {
	Career string
	Labs   []LaboratoryDTO
}

func NewLaboratoriesParser(file io.ReadCloser) (*LaboratoriesParser, error) {
	layoutsDir := filepath.Join(config.Get().Paths.BaseDir, "internal", "infrastructure", "parser", "layouts", "laboratories")

	en, err := engine.NewParser(file, layoutsDir)
	if err != nil {
		return nil, err
	}
	en.SheetFilter = shouldParseSheet

	return &LaboratoriesParser{
		engine: en,
	}, nil
}

func (ep *LaboratoriesParser) Close() {
	ep.engine.Close()
}

func (ep *LaboratoriesParser) ParseNextSheet() (*ParsedLaboratories, error) {
	name, ok := ep.engine.NextSheet()
	if !ok {
		// There is no sheet to parse
		return nil, nil
	}

	acc := newLabAccumulator()

	err := ep.engine.ParseSheet(name, func(row []string, lay *engine.Layout, startingCell int) error {
		acc.parseRow(row, lay, startingCell)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ParsedLaboratories{
		Career: strings.ToUpper(strings.ReplaceAll(name, " ", "")),
		Labs:   acc.flatten(),
	}, nil
}

// =================================================
// Per-sheet accumulator
// =================================================

// labAccumulator groups the parsed laboratories of a single sheet. It is recreated for every
// sheet so no state leaks between careers.
type labAccumulator struct {
	labs map[labKey]*LaboratoryDTO

	// Last non-empty subject/plan seen. Excel tends to merge these cells vertically, so a
	// row may not repeat them and we need to carry them forward.
	currentSubject string
	currentPlan    string
}

func newLabAccumulator() *labAccumulator {
	return &labAccumulator{
		labs: make(map[labKey]*LaboratoryDTO, 64),
	}
}

// parseRow extracts the subject, the plan and the schedule of every day from a single row.
// Headers not present in the switch are safely ignored, which lets the layout evolve without
// touching this code.
func (a *labAccumulator) parseRow(row []string, lay *engine.Layout, startingCell int) {
	var subject, plan string
	var days [7]string

	current := startingCell - 1
	for _, field := range lay.Headers {
		current++
		if current >= len(row) {
			break
		}

		val := strings.TrimSpace(row[current])
		if val == "" {
			continue
		}

		switch field {
		case "asignatura":
			subject = val
		case "plan":
			plan = val
		case "horaLunes":
			days[academic.Monday] = val
		case "horaMartes":
			days[academic.Tuesday] = val
		case "horaMiercoles":
			days[academic.Wednesday] = val
		case "horaJueves":
			days[academic.Thursday] = val
		case "horaViernes":
			days[academic.Friday] = val
		case "horaSabado":
			days[academic.Saturday] = val
		default:
			// Ignore any other column: this is our wildcard.
		}
	}

	// Carry forward the subject/plan when the cell is empty (merged cells).
	if subject != "" {
		a.currentSubject = subject
	}
	if plan != "" {
		a.currentPlan = plan
	}
	if a.currentSubject == "" || a.currentPlan == "" {
		return
	}

	for day := academic.Monday; day <= academic.Saturday; day++ {
		if days[day] == "" {
			continue
		}
		a.addDay(day, days[day])
	}
}

// addDay parses a day cell and stores every extracted section into the accumulator.
func (a *labAccumulator) addDay(day academic.WeekDay, cell string) {
	for _, entry := range parseDayCell(cell) {
		key := labKey{
			subject: commons.NormalizeKey(a.currentSubject),
			plan:    commons.NormalizeKey(a.currentPlan),
			section: commons.NormalizeKey(entry.Section),
		}

		dto, ok := a.labs[key]
		if !ok {
			dto = &LaboratoryDTO{
				RawName: a.currentSubject,
				Plan:    a.currentPlan,
				Section: entry.Section,
			}
			a.labs[key] = dto
		}

		// A section should only have one slot per day; when the same day is listed again
		// (usually because it appears on another turno row) we keep the latest value. The
		// unique-section case is the only genuinely ambiguous one, so it is the only one
		// that gets a warning to avoid flooding the logs.
		if dto.WeekSchedule[day].Time.Start.Valid && strings.EqualFold(entry.Section, defaultSection) {
			logger.Warn("laboratory schedule overwritten for unique section",
				"subject", a.currentSubject,
				"plan", a.currentPlan,
				"day", day.String(),
			)
		}

		dto.WeekSchedule[day] = WeekDayData{Time: entry.Time}
	}
}

// flatten converts the internal map into a deterministic slice of DTOs.
func (a *labAccumulator) flatten() []LaboratoryDTO {
	keys := make([]labKey, 0, len(a.labs))
	for k := range a.labs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].subject != keys[j].subject {
			return keys[i].subject < keys[j].subject
		}
		if keys[i].plan != keys[j].plan {
			return keys[i].plan < keys[j].plan
		}
		return keys[i].section < keys[j].section
	})

	labs := make([]LaboratoryDTO, 0, len(keys))
	for _, k := range keys {
		labs = append(labs, *a.labs[k])
	}
	return labs
}

// =================================================
// Cleaning / parsing helpers
// =================================================

// dayEntry is a single (section, time slot) pair extracted from a day cell.
type dayEntry struct {
	Section string
	Time    commons.TimeSlot
}

// parseDayCell parses the schedule cell of a single weekday. A cell may hold several
// sections, one per line, for example:
//
//	18:00 - 20:00 (T1)
//	20:00 - 22:00 (T2)
//
// When the subject exposes a single section there is no "(seccion)" tag, so the entry falls
// back to defaultSection. Lines without a valid time range are ignored.
func parseDayCell(cell string) []dayEntry {
	entries := make([]dayEntry, 0, 4)

	commons.ScanLinesN(cell, 64, func(_ int, line string) {
		section, hourPart := splitSection(line)

		slot := commons.ParseTimeSlot(hourPart)
		if !slot.Start.Valid || !slot.End.Valid {
			return
		}

		if section == "" {
			section = defaultSection
		}

		entries = append(entries, dayEntry{Section: section, Time: slot})
	})

	return entries
}

// splitSection separates the optional "(seccion)" tag from the time range. The tag is usually
// at the end "18:00 - 20:00 (T1)" but may also be at the beginning; the content between the
// parentheses is kept as-is. The returned hour part has the tag removed.
func splitSection(line string) (section, hour string) {
	start := strings.IndexByte(line, '(')
	if start == -1 {
		return "", strings.TrimSpace(line)
	}

	relEnd := strings.IndexByte(line[start:], ')')
	if relEnd == -1 {
		// Malformed tag without a closing parenthesis, treat the line as a plain hour range.
		return "", strings.TrimSpace(line)
	}

	end := start + relEnd

	section = strings.TrimSpace(line[start+1 : end])
	hour = strings.TrimSpace(line[:start] + " " + line[end+1:])
	return section, hour
}
