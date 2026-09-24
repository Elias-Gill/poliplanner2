package academic

import (
	"fmt"
	"strings"
)

type LaboratoryID int64

// Laboratory is a laboratory offering of a subject within a curriculum and
// academic period. Its schedule is independent from the course schedule.
type Laboratory struct {
	ID       LaboratoryID
	Section  string
	Schedule []ClassSession
}

// FormattedSchedule returns a compact human readable weekly schedule for templates.
func (l Laboratory) FormattedSchedule() string {
	if len(l.Schedule) == 0 {
		return "Sin horario asignado"
	}

	parts := make([]string, 0, len(l.Schedule))
	for _, s := range l.Schedule {
		parts = append(parts, fmt.Sprintf("%s %s", s.Day, s.Time))
	}
	return strings.Join(parts, " | ")
}
