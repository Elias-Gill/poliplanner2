package commons

import (
	"strconv"
	"strings"
)

type Hour struct {
	Hour   int
	Minute int
	Valid  bool
}

type Date struct {
	Year  int
	Month int
	Day   int
	Valid bool
}

type TimeSlot struct {
	Start Hour
	End   Hour
}

// ParseTime convierte texto a Hour sin alocaciones de Regex.
func ParseTime(timeStr string) Hour {
	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.TrimRight(timeStr, "hs h.")
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return Hour{}
	}

	if before, after, ok := strings.Cut(timeStr, ":"); ok {
		hours := ParseDigits(before)
		minutes := ParseDigits(after)

		if hours > 23 {
			hours = 23
		}
		if minutes > 59 {
			minutes = 59
		}
		return Hour{Hour: hours, Minute: minutes, Valid: true}
	}

	if decimalValue, err := strconv.ParseFloat(timeStr, 64); err == nil {
		totalMinutes := int(decimalValue * 24 * 60)
		return Hour{Hour: (totalMinutes / 60) % 24, Minute: totalMinutes % 60, Valid: true}
	}

	return Hour{}
}

// ParseDate analiza numéricamente una cadena de fecha.
func ParseDate(value string) Date {
	value = strings.TrimSpace(value)
	if value == "" {
		return Date{}
	}

	var parts [3]int
	partIdx, currentVal := 0, 0
	hasDigits := false

	for i := 0; i < len(value); i++ {
		c := value[i]
		if c >= '0' && c <= '9' {
			currentVal = currentVal*10 + int(c-'0')
			hasDigits = true
		} else if hasDigits {
			if partIdx < 3 {
				parts[partIdx] = currentVal
				partIdx++
			}
			currentVal = 0
			hasDigits = false
		}
	}
	if hasDigits && partIdx < 3 {
		parts[partIdx] = currentVal
		partIdx++
	}

	if partIdx < 3 {
		return Date{}
	}

	day, month, year := parts[0], parts[1], parts[2]
	if year < 100 {
		year += 2000
	}

	if !IsValidDate(day, month, year) {
		return Date{}
	}

	return Date{Year: year, Month: month, Day: day, Valid: true}
}

// ParseTimeSlot parsea un rango de horarios (ej: "07:30 - 09:00").
func ParseTimeSlot(val string) TimeSlot {
	val = strings.TrimSpace(val)
	val = strings.TrimRight(strings.ToLower(val), "hs h.")
	val = strings.TrimSpace(val)

	before, after, ok := strings.Cut(val, "-")
	if !ok {
		return TimeSlot{}
	}

	return TimeSlot{
		Start: ParseTime(before),
		End:   ParseTime(after),
	}
}

// ConvertStringToNumber limpia y convierte una cadena a número entero.
func ConvertStringToNumber(str string) int {
	str = strings.TrimSpace(str)
	if str == "" || str == "-" {
		return 0
	}

	var keep [16]byte
	kIdx := 0
	for i := 0; i < len(str); i++ {
		c := str[i]
		if (c >= '0' && c <= '9') || c == '-' {
			if kIdx < len(keep) {
				keep[kIdx] = c
				kIdx++
			}
		} else if c == ',' || c == '.' {
			if kIdx < len(keep) {
				keep[kIdx] = '.'
				kIdx++
			}
		}
	}
	if kIdx == 0 {
		return 0
	}

	val, err := strconv.ParseFloat(string(keep[:kIdx]), 64)
	if err != nil {
		return 0
	}
	return int(val + 0.5)
}

func ParseDigits(s string) int {
	res := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			res = res*10 + int(s[i]-'0')
		}
	}
	return res
}

func IsValidDate(day, month, year int) bool {
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

// NormalizeKey recorta, colapsa espacios internos y pasa a mayúsculas. Se usa para
// construir claves de agrupación (por ejemplo materia + plan + sección) de forma que
// diferencias de espaciado o mayúsculas no impidan detectar entidades repetidas.
//
// NOTA: es una normalización exclusiva para comparar/agrupar, no un valor de negocio.
func NormalizeKey(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToUpper(strings.Join(fields, " "))
}

// ScanLines escanea texto multilínea sin alocar slices intermedios. Mantiene un tope de 4
// líneas (pensado para los profesores de la planilla de horarios).
func ScanLines(input string, assign func(idx int, line string)) int {
	return ScanLinesN(input, 4, assign)
}

// ScanLinesN escanea hasta `max` líneas no vacías, asignando cada una mediante assign.
// Conserva el mismo comportamiento que ScanLines (trim y salto de líneas vacías) pero
// permite a otros parsers extender el límite.
func ScanLinesN(input string, max int, assign func(idx int, line string)) int {
	if max <= 0 {
		return 0
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return 0
	}
	idx := 0
	for idx < max {
		next := strings.IndexByte(input, '\n')
		var line string
		if next == -1 {
			line = strings.TrimSpace(input)
		} else {
			line = strings.TrimSpace(input[:next])
		}

		if line != "" {
			assign(idx, line)
			idx++
		}
		if next == -1 {
			break
		}
		input = input[next+1:]
	}
	return idx
}
