package scraper

import "regexp"

// Name patterns used to classify the Excel files published by the university.
// They are intentionally kept private to this package: the engine only receives
// the resulting filter function.
var (
	// Permite espacios y cualquier caracter entre la palabra clave y la extension .xlsx
	schedulePattern = regexp.MustCompile(
		`(?i).*(horario|clases|examen(?:es)?|exame|exam).*\.xlsx$`)
	laboratoryPattern = regexp.MustCompile(
		`(?i).*(laboratorio(?:s)?|lab|asignacior|asignacion).*\.xlsx$`)
)

// scheduleFilter accepts schedule files but explicitly rejects laboratory ones.
// Laboratory files are often named "Horario de practicas del laboratorio ...",
// which would otherwise match the schedule pattern too, so this exclusion is
// what keeps the two scrapers truly independent.
func scheduleFilter(name string) bool {
	return schedulePattern.MatchString(name) && !laboratoryPattern.MatchString(name)
}

func laboratoryFilter(name string) bool {
	return laboratoryPattern.MatchString(name)
}
