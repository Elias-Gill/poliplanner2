package scrapper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func extractDateFromFilename(filename string) (time.Time, error) {
	nameWithoutExt := strings.TrimSuffix(filename, ".xlsx")
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{4})`),
		regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`),
		regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`),
		regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`),
		regexp.MustCompile(`(\d{2})/(\d{2})/(\d{4})`),
		regexp.MustCompile(`(\d{2})(\d{2})(\d{4})`),
	}
	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(nameWithoutExt); matches != nil {
			var day, month, year int
			var err error
			if len(matches[1]) == 4 {
				year, err = strconv.Atoi(matches[1])
				if err != nil { continue }
				month, err = strconv.Atoi(matches[2])
				if err != nil { continue }
				day, err = strconv.Atoi(matches[3])
				if err != nil { continue }
			} else {
				day, err = strconv.Atoi(matches[1])
				if err != nil { continue }
				month, err = strconv.Atoi(matches[2])
				if err != nil { continue }
				year, err = strconv.Atoi(matches[3])
				if err != nil { continue }
			}
			if year >= 2000 && year <= 2100 && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
				return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("no date found in filename: %s", filename)
}
