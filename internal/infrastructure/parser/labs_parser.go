package parser

import (
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/commons"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/engine"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/logger"
)

// Laboratories are parsed differently from class schedules. In the schedules file each row is
// one course section with a single schedule per day, so a row maps directly to a DTO. Here a
// single day cell can hold several sections stacked as lines, each one with its own time range
// and an optional "(section)" tag:
//
//	18:00 - 20:00 (T1)
//	20:00 - 22:00 (T2)
//
// The same section may also reappear on later rows (other "turnos") contributing more days, so a
// row cannot be converted to a DTO in isolation. Instead we accumulate entries in a hashmap
// keyed by the offering identity (subject + plan + section + career) and merge each day slot
// while scanning.
//
// The name alone is not enough to identify a section: the same name can be reused with a
// different schedule, which means a different section. Therefore the schedule is part of the
// identity too: an occurrence only merges into a section with the same name when its day is
// free or already holds the same time, otherwise a new section is started. Sections that end up
// sharing a name get a numeric suffix at flatten. The career is part of the key because a single
// sheet carries rows from other careers; those keys get dropped during flatten. When a day cell
// has no tag it means the subject exposes a single section, so it falls back to defaultSection.

// defaultSection identifies a subject that only exposes a single laboratory section. In that
// case the day cell does not carry a "(seccion)" tag.
const defaultSection = "UNICA"

// labKey is the identity of a laboratory offering. A subject may be taught under several
// plans, and each plan may expose several sections (T1, T2, ...). The career is part of the
// key because a single sheet can carry rows from other careers; keeping them separate avoids
// contaminating valid entries, and they get discarded during flatten.
//
// The fields are normalized only to detect repeated entities; they are never exposed as the
// final result.
type labKey struct {
	subject string
	plan    string
	section string
	career  string
}

type LaboratoriesParser struct {
	engine *engine.ParserEngine
}

type ParsedLaboratories struct {
	Career string
	Labs   []LaboratoryDTO
}

func NewLaboratoriesParser(file io.ReadCloser, layoutsDir string) (*LaboratoriesParser, error) {
	laboratoriesDir := filepath.Join(layoutsDir, "laboratories")

	en, err := engine.NewParser(file, laboratoriesDir)
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

	acc := newLabAccumulator(commons.NormalizeCareer(name))

	err := ep.engine.ParseSheet(name, func(row []string, lay *engine.Layout, startingCell int) error {
		acc.parseRow(row, lay, startingCell)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ParsedLaboratories{
		Career: commons.NormalizeCareer(name),
		Labs:   acc.flatten(),
	}, nil
}

// =================================================
// Per-sheet accumulator
// =================================================

// labAccumulator groups the parsed laboratories of a single sheet. It is recreated for every
// sheet so no state leaks between careers.
type labAccumulator struct {
	// Multiple laboratories can share the same name (subject + plan + section + career) when
	// their schedule differs, so each key maps to a list of distinct schedules.
	labs map[labKey][]*LaboratoryDTO

	// Career of the sheet being parsed, normalized. Only entries whose career matches it are
	// kept during flatten.
	sheetCareer string

	// Last non-empty subject/plan/career/semester seen. Excel tends to merge these cells
	// vertically, so a row may not repeat them and we need to carry them forward.
	currentSubject string
	currentPlan    string
	currentCareer  string
	currentSemester string
}

func newLabAccumulator(sheetCareer string) *labAccumulator {
	return &labAccumulator{
		labs:        make(map[labKey][]*LaboratoryDTO, 64),
		sheetCareer: sheetCareer,
	}
}

// parseRow extracts the subject, the plan and the schedule of every day from a single row.
// Headers not present in the switch are safely ignored, which lets the layout evolve without
// touching this code.
func (a *labAccumulator) parseRow(row []string, lay *engine.Layout, startingCell int) {
	var subject, plan, career, semester string
	var days [7]string

	current := startingCell - 1
	// Parse the entire row first after creating the new object
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
		case "carrera":
			career = val
		case "semestre":
			semester = val
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

	// Carry forward the subject/plan/career/semester when the cell is empty (merged cells).
	if subject != "" {
		a.currentSubject = subject
	}
	if plan != "" {
		a.currentPlan = plan
	}
	if career != "" {
		a.currentCareer = commons.NormalizeCareer(career)
	}
	if semester != "" {
		a.currentSemester = semester
	}
	if a.currentSubject == "" || a.currentPlan == "" {
		return
	}

	// A row without career information is assumed to belong to the sheet being parsed.
	rowCareer := a.currentCareer
	if rowCareer == "" {
		rowCareer = a.sheetCareer
	}

	for day := academic.Monday; day <= academic.Saturday; day++ {
		if days[day] == "" {
			continue
		}
		a.addDay(day, days[day], rowCareer)
	}
}

// addDay parses a day cell and stores every extracted section into the accumulator.
func (a *labAccumulator) addDay(day academic.WeekDay, cell string, career string) {
	for _, entry := range parseDayCell(cell) {
		a.addEntry(day, entry, career)
	}
}

// addEntry stores a single day occurrence. The section is no longer identified only by its
// name: the schedule is part of the identity too. The occurrence is merged into an existing
// section with the same name when its day is free or already holds the same time; when every
// candidate already has a different time for that day, it is a different section and a new
// one is created instead of overwriting.
func (a *labAccumulator) addEntry(day academic.WeekDay, entry dayEntry, career string) {
	key := labKey{
		subject: commons.NormalizeKey(a.currentSubject),
		plan:    commons.NormalizeKey(a.currentPlan),
		section: commons.NormalizeKey(entry.Section),
		career:  career,
	}

	// Prefer an existing section that already holds this exact slot, so a free day in another
	// section with the same name does not steal the occurrence.
	for _, dto := range a.labs[key] {
		if dto.WeekSchedule[day].Time == entry.Time {
			return
		}
	}
	for _, dto := range a.labs[key] {
		if !dto.WeekSchedule[day].Time.Start.Valid {
			dto.WeekSchedule[day] = WeekDayData{Time: entry.Time}
			return
		}
	}

	semester, err := strconv.Atoi(a.currentSemester)
	if err != nil {
		semester = 0
	}

	dto := &LaboratoryDTO{
		RawName:  a.currentSubject,
		Plan:     a.currentPlan,
		Section:  entry.Section,
		Semester: int(semester),
	}
	dto.WeekSchedule[day] = WeekDayData{Time: entry.Time}
	a.labs[key] = append(a.labs[key], dto)

	if strings.EqualFold(entry.Section, defaultSection) && len(a.labs[key]) > 1 {
		logger.Warn("multiple schedules for an unnamed laboratory section",
			"subject", a.currentSubject,
			"plan", a.currentPlan,
			"day", day.String(),
		)
	}
}

// flatten converts the internal map into a deterministic slice of DTOs. Entries whose career
// does not match the sheet being parsed are dropped, removing the foreign rows that the same
// sheet can carry. Sections that ended up sharing a name (same name, different schedule) get a
// numeric suffix so they stay unique within their curriculum.
func (a *labAccumulator) flatten() []LaboratoryDTO {
	keys := make([]labKey, 0, len(a.labs))
	for k := range a.labs {
		if k.career != a.sheetCareer {
			continue
		}
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

	usedNames := make(map[string]map[string]bool)
	labs := make([]LaboratoryDTO, 0, len(keys))
	for _, k := range keys {
		malla := k.subject + "\x00" + k.plan + "\x00" + k.career
		names, ok := usedNames[malla]
		if !ok {
			names = make(map[string]bool)
			usedNames[malla] = names
		}

		for _, dto := range a.labs[k] {
			base := dto.Section
			if base == "" {
				base = defaultSection
			}
			dto.Section = uniqueSectionName(base, names)
			labs = append(labs, *dto)
		}
	}
	return labs
}

// uniqueSectionName returns base, or base_2, base_3, ... if it is already taken.
func uniqueSectionName(base string, taken map[string]bool) string {
	name := base
	for i := 2; taken[name]; i++ {
		name = base + "_" + strconv.Itoa(i)
	}
	taken[name] = true
	return name
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
