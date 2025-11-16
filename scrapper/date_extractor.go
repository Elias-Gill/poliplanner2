package scrapper

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

func extractDateFromFilename(filename string) time.Time {
	// Remove file extension
	nameWithoutExt := strings.TrimSuffix(filename, ".xlsx")
	
	// Common date patterns in filenames
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{4})`),     // dd.mm.yyyy
		regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`),       // dd-mm-yyyy
		regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`),     // yyyy.mm.dd
		regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`),       // yyyy-mm-dd
		regexp.MustCompile(`(\d{2})/(\d{2})/(\d{4})`),       // dd/mm/yyyy
		regexp.MustCompile(`(\d{2})(\d{2})(\d{4})`),         // ddmmyyyy
	}
	
	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(nameWithoutExt); matches != nil {
			var day, month, year int
			if len(matches[1]) == 4 { // yyyy-mm-dd format
				year, _ = strconv.Atoi(matches[1])
				month, _ = strconv.Atoi(matches[2])
				day, _ = strconv.Atoi(matches[3])
			} else { // dd-mm-yyyy O ddmmyyyy format
				day, _ = strconv.Atoi(matches[1])
				month, _ = strconv.Atoi(matches[2])
				year, _ = strconv.Atoi(matches[3])
			}
			
			// Validate date
			if year >= 2000 && year <= 2100 && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
				return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			}
		}
	}
	return time.Time{}
}
