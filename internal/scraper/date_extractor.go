package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func extractDateFromFilename(filename string) (time.Time, error) {
	// Remove .xlsx extension if present
	nameWithoutExt := strings.TrimSuffix(filename, ".xlsx")

	// list of common date patterns in filenames
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{4})`), // DD.MM.YYYY
		regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`),   // DD-MM-YYYY
		regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`), // YYYY.MM.DD
		regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`),   // YYYY-MM-DD
		regexp.MustCompile(`(\d{2})/(\d{2})/(\d{4})`),   // DD/MM/YYYY
		regexp.MustCompile(`(\d{2})(\d{2})(\d{4})`),     // DDMMYYYY (no separator)
	}

	// try each pattern in order
	for _, pattern := range patterns {
		// found a match
		if matches := pattern.FindStringSubmatch(nameWithoutExt); matches != nil {
			var day, month, year int
			var err error
			// if year is first in the pattern (YYYY...)
			if len(matches[1]) == 4 {
				year, err = strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				month, err = strconv.Atoi(matches[2])
				if err != nil {
					continue
				}
				day, err = strconv.Atoi(matches[3])
				if err != nil {
					continue
				}
			} else { // Else the DAY is first (DD...)
				day, err = strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				month, err = strconv.Atoi(matches[2])
				if err != nil {
					continue
				}
				year, err = strconv.Atoi(matches[3])
				if err != nil {
					continue
				}
			}

			// Basic date validation
			if year >= 2000 && year <= 2100 && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
				// return parsed date in UTC
				return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("no date found in filename: %s", filename)
}
