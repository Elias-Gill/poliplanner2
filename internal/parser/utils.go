package parser

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Global compiled patterns
var (
	datePattern  = regexp.MustCompile(`(\d{1,2})[^\d]+(\d{1,2})[^\d]+(\d{2,4})`)
	nonTimeChars = regexp.MustCompile(`[^0-9:.]`)
	nonNumeric   = regexp.MustCompile(`[^0-9.,-]`)

	// Cache for parsed dates and times
	dateCache sync.Map
	timeCache sync.Map
)

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
