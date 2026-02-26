package parser

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const examOnlyCourse = 1
const normalCourse = 0

// Global compiled patterns
var (
	datePattern  = regexp.MustCompile(`(\d{1,2})[^\d]+(\d{1,2})[^\d]+(\d{2,4})`)
	nonTimeChars = regexp.MustCompile(`[^0-9:.]`)
	nonNumeric   = regexp.MustCompile(`[^0-9.,-]`)

	// Cache for parsed dates and times
	dateCache sync.Map
	timeCache sync.Map
)

// Buffer pool for time formatting
var timeBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 10)
	},
}

type TeacherDTO struct {
	Title     string
	FirstName string
	LastName  string
	Email     string
}

type TimeSlot struct {
	Start string
	End   string
}

type SubjectDTO struct {
	// General info
	Department string
	Semester   int
	Level      int
	Section    string

	// This is used as the name of the table "cursos", which is the final agreggate
	// with all schedule information of a subject in a specific time period.
	RawSubjectName string

	// this contains the plaussible real name of the subject. Example:
	// "Electiva I - Machine Learning" contains: "electiva I"
	// "Física 2" normalizes to: "fisica II"
	TentativeRealSubjectName string

	// Course type can be:
	// - Normal course
	// - Final exam only
	CourseType int

	// Teachers info
	Teachers []TeacherDTO

	// Exams
	// First partial
	Partial1Date *time.Time
	Partial1Time string
	Partial1Room string

	// Second partial
	Partial2Date *time.Time
	Partial2Time string
	Partial2Room string

	// First final
	Final1Date    *time.Time
	Final1Time    string
	Final1Room    string
	Final1RevDate *time.Time
	Final1RevTime string

	// Second final
	Final2Date    *time.Time
	Final2Time    string
	Final2Room    string
	Final2RevDate *time.Time
	Final2RevTime string

	// Weekly schedule
	MondayRoom string
	Monday     TimeSlot

	TuesdayRoom string
	Tuesday     TimeSlot

	WednesdayRoom string
	Wednesday     TimeSlot

	ThursdayRoom string
	Thursday     TimeSlot

	FridayRoom string
	Friday     TimeSlot

	SaturdayRoom  string
	Saturday      TimeSlot
	SaturdayDates string

	// Committee
	CommitteePresident string
	CommitteeMember1   string
	CommitteeMember2   string
}

// -----------------------------
// Setters with cleaning methods
// -----------------------------
func (s *SubjectDTO) SetDepartment(val string) {
	s.Department = strings.TrimSpace(val)
}

func (s *SubjectDTO) SetSubjectName(val string) {
	s.RawSubjectName = strings.TrimSpace(val)
	s.TentativeRealSubjectName = normalizeSubjectName(val)

	// Set course type based on the name
	s.CourseType = normalCourse
	// If contains a (*) it is a closed grade with only final exam
	for i := len(val) - 1; i >= 0; i-- {
		if rune(val[i]) == '*' {
			s.CourseType = examOnlyCourse
			break
		}
	}
}

func (s *SubjectDTO) SetSemester(val string) {
	s.Semester = convertStringToNumber(val)
}

func (s *SubjectDTO) SetLevel(val string) {
	s.Level = convertStringToNumber(val)
}

func (s *SubjectDTO) SetSection(val string) {
	s.Section = strings.TrimSpace(val)
}

func (s *SubjectDTO) SetTeachersFirtNames(firstNames string) {
	list := strings.Split(firstNames, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].FirstName = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) SetTeachersLastNames(secondNames string) {
	list := strings.Split(secondNames, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].LastName = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) SetTeachersTitles(titles string) {
	list := strings.Split(titles, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].Title = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) SetTeachersEmails(emails string) {
	list := strings.Split(emails, "\n")
	s.ensureTeachersLen(max(len(s.Teachers), len(list)))

	for i, v := range list {
		s.Teachers[i].Email = strings.TrimSpace(v)
	}
}

func (s *SubjectDTO) ensureTeachersLen(n int) {
	if len(s.Teachers) < n {
		s.Teachers = append(s.Teachers, make([]TeacherDTO, n-len(s.Teachers))...)
	}
}

func (s *SubjectDTO) SetPartial1Date(val string) {
	s.Partial1Date = parseDate(val)
}

func (s *SubjectDTO) SetPartial1Time(val string) {
	s.Partial1Time = cleanTime(val)
}

func (s *SubjectDTO) SetPartial1Room(val string) {
	s.Partial1Room = val
}

func (s *SubjectDTO) SetPartial2Date(val string) {
	s.Partial2Date = parseDate(val)
}

func (s *SubjectDTO) SetPartial2Time(val string) {
	s.Partial2Time = cleanTime(val)
}

func (s *SubjectDTO) SetPartial2Room(val string) {
	s.Partial2Room = val
}

func (s *SubjectDTO) SetFinal1Date(val string) {
	s.Final1Date = parseDate(val)
}

func (s *SubjectDTO) SetFinal1Time(val string) {
	s.Final1Time = cleanTime(val)
}

func (s *SubjectDTO) SetFinal1Room(val string) {
	s.Final1Room = val
}

func (s *SubjectDTO) SetFinal1RevDate(val string) {
	s.Final1RevDate = parseDate(val)
}

func (s *SubjectDTO) SetFinal1RevTime(val string) {
	s.Final1RevTime = cleanTime(val)
}

func (s *SubjectDTO) SetFinal2Date(val string) {
	s.Final2Date = parseDate(val)
}

func (s *SubjectDTO) SetFinal2Time(val string) {
	s.Final2Time = cleanTime(val)
}

func (s *SubjectDTO) SetFinal2Room(val string) {
	s.Final2Room = val
}

func (s *SubjectDTO) SetFinal2RevDate(val string) {
	s.Final2RevDate = parseDate(val)
}

func (s *SubjectDTO) SetFinal2RevTime(val string) {
	s.Final2RevTime = cleanTime(val)
}

func (s *SubjectDTO) SetCommitteePresident(val string) {
	s.CommitteePresident = val
}

func (s *SubjectDTO) SetCommitteeMember1(val string) {
	s.CommitteeMember1 = val
}

func (s *SubjectDTO) SetCommitteeMember2(val string) {
	s.CommitteeMember2 = val
}

func (s *SubjectDTO) SetMondayRoom(val string) {
	s.MondayRoom = val
}

func (s *SubjectDTO) SetMonday(val string) {
	s.Monday = convertIntoTimeSlot(val)
}

func (s *SubjectDTO) SetTuesdayRoom(val string) {
	s.TuesdayRoom = val
}

func (s *SubjectDTO) SetTuesday(val string) {
	s.Tuesday = convertIntoTimeSlot(val)
}

func (s *SubjectDTO) SetWednesdayRoom(val string) {
	s.WednesdayRoom = val
}

func (s *SubjectDTO) SetWednesday(val string) {
	s.Wednesday = convertIntoTimeSlot(val)
}

func (s *SubjectDTO) SetThursdayRoom(val string) {
	s.ThursdayRoom = val
}

func (s *SubjectDTO) SetThursday(val string) {
	s.Thursday = convertIntoTimeSlot(val)
}

func (s *SubjectDTO) SetFridayRoom(val string) {
	s.FridayRoom = val
}

func (s *SubjectDTO) SetFriday(val string) {
	s.Friday = convertIntoTimeSlot(val)
}

func (s *SubjectDTO) SetSaturdayRoom(val string) {
	s.SaturdayRoom = val
}

func (s *SubjectDTO) SetSaturday(val string) {
	s.Saturday = convertIntoTimeSlot(val)
}

func (s *SubjectDTO) SetSaturdayDates(val string) {
	s.SaturdayDates = val
}

// ----------------------------------------
// Cleaning helper functions
// ----------------------------------------

// Updated cleanTime function with performance improvements
func cleanTime(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	// Check cache first
	if cached, found := timeCache.Load(timeStr); found {
		return cached.(string)
	}

	// Remove non-numeric characters (except : and .)
	cleaned := nonTimeChars.ReplaceAllString(timeStr, "")
	cleaned = strings.TrimSpace(cleaned)

	var result string
	if strings.Contains(cleaned, ":") {
		// Parse as hh:mm format
		segments := strings.Split(cleaned, ":")
		if len(segments) >= 2 {
			hours, err1 := strconv.Atoi(segments[0])
			minutes, err2 := strconv.Atoi(segments[1])

			if err1 == nil && err2 == nil {
				// Validate ranges
				if hours < 0 {
					hours = 0
				}
				if hours > 23 {
					hours = 23
				}
				if minutes < 0 {
					minutes = 0
				}
				if minutes > 59 {
					minutes = 59
				}

				// Use buffer for string building
				buf := timeBufferPool.Get().([]byte)
				buf = buf[:0]

				if hours < 10 {
					buf = append(buf, '0')
				}
				buf = strconv.AppendInt(buf, int64(hours), 10)
				buf = append(buf, ':')
				if minutes < 10 {
					buf = append(buf, '0')
				}
				buf = strconv.AppendInt(buf, int64(minutes), 10)

				result = string(buf)
				timeBufferPool.Put(buf)
			}
		}
	} else {
		// Parse as Excel decimal format
		decimalValue, err := strconv.ParseFloat(cleaned, 64)
		if err == nil {
			totalMinutes := int(decimalValue * 24 * 60)
			hours := (totalMinutes / 60) % 24
			minutes := totalMinutes % 60

			// Use buffer for string building
			buf := timeBufferPool.Get().([]byte)
			buf = buf[:0]

			if hours < 10 {
				buf = append(buf, '0')
			}
			buf = strconv.AppendInt(buf, int64(hours), 10)
			buf = append(buf, ':')
			if minutes < 10 {
				buf = append(buf, '0')
			}
			buf = strconv.AppendInt(buf, int64(minutes), 10)

			result = string(buf)
			timeBufferPool.Put(buf)
		}
	}

	// Cache the result if valid
	if result != "" {
		timeCache.Store(timeStr, result)
	}

	return result
}

// Updated convertStringToNumber with regex performance
func convertStringToNumber(str string) int {
	if str == "" {
		return 0
	}

	// Remove non-numeric characters and replace commas with dots
	cleaned := nonNumeric.ReplaceAllString(str, "")
	cleaned = strings.ReplaceAll(cleaned, ",", ".")

	if cleaned == "" || cleaned == "-" || cleaned == "." {
		return 0
	}

	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}

	return int(value + 0.5)
}

// Updated parseDate function with caching and direct time construction
func parseDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	// Check cache first
	if cached, found := dateCache.Load(value); found {
		return cached.(*time.Time)
	}

	matches := datePattern.FindStringSubmatch(value)
	if matches == nil {
		return nil
	}

	day, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	yearStr := matches[3]

	// Normalize year
	if len(yearStr) == 2 {
		yearStr = "20" + yearStr
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil
	}

	// Validate date
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
	}

	// Create time directly
	var t *time.Time
	if isValidDate(day, month, year) {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		t = &date

		// Cache the result
		dateCache.Store(value, t)
	}

	return t
}

// Helper function to validate date components
func isValidDate(day, month, year int) bool {
	if month < 1 || month > 12 || day < 1 {
		return false
	}

	daysInMonth := 31
	switch month {
	case 2:
		if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
			daysInMonth = 29
		} else {
			daysInMonth = 28
		}
	case 4, 6, 9, 11:
		daysInMonth = 30
	}

	return day <= daysInMonth
}

func (d *SubjectDTO) Reset() {
	*d = SubjectDTO{}
}

var roman = [...]string{
	"", "I", "II", "III", "IV", "V",
	"VI", "VII", "VIII", "IX", "X",
	"XI", "XII", "XIII", "XIV", "XV",
	"XVI", "XVII", "XVIII", "XIX", "XX",
}

var romanToInt = map[string]int{
	"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5,
	"VI": 6, "VII": 7, "VIII": 8, "IX": 9, "X": 10,
	"XI": 11, "XII": 12, "XIII": 13, "XIV": 14, "XV": 15,
	"XVI": 16, "XVII": 17, "XVIII": 18, "XIX": 19, "XX": 20,
}

var accents = strings.NewReplacer(
	"á", "a", "Á", "A",
	"é", "e", "É", "E",
	"í", "i", "Í", "I",
	"ó", "o", "Ó", "O",
	"ú", "u", "Ú", "U",
)

func normalizeSubjectName(val string) string {
	// Parte antes del -
	if i := strings.IndexByte(val, '-'); i >= 0 {
		val = val[:i]
	}
	val = strings.TrimSpace(val)

	// Reemplazo de parentesis
	val = regexp.MustCompile(`\s*\([^()]*\)`).ReplaceAllString(val, "")

	fields := strings.Fields(val)
	if len(fields) == 0 {
		return ""
	}

	for i := range fields {
		f := fields[i]

		// Número arábigo → romano
		if n, err := strconv.Atoi(f); err == nil && n >= 1 && n <= 20 {
			fields[i] = roman[n]
			continue
		}

		// Romano ya en mayúscula
		up := strings.ToUpper(f)
		if _, ok := romanToInt[up]; ok {
			fields[i] = up
			continue
		}

		// Normalizar a minúsculas y quitar acentos
		fields[i] = strings.ToLower(accents.Replace(f))
	}

	return strings.Join(fields, " ")
}

func convertIntoTimeSlot(val string) TimeSlot {
	val = strings.TrimSpace(val)
	if val == "" {
		return TimeSlot{}
	}

	// Quitar sufijos comunes como "hs", "h", "."
	val = strings.TrimRight(strings.ToLower(val), "hs h.")
	val = strings.TrimSpace(val)

	parts := strings.Split(val, "-")
	if len(parts) != 2 {
		return TimeSlot{}
	}

	start := cleanTime(strings.TrimSpace(parts[0]))
	end := cleanTime(strings.TrimSpace(parts[1]))

	if start == "" || end == "" {
		return TimeSlot{}
	}

	return TimeSlot{
		Start: start,
		End:   end,
	}
}
