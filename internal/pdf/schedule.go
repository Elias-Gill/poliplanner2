package pdf

import (
	"fmt"
	"io"
	"strings"

	_ "embed"
	model "github.com/elias-gill/poliplanner2/internal/model/schedule"
	"github.com/signintech/gopdf"
)

// NOTE: las fuentes para la generacion de pdf ya van embebidas en el compilado final

//go:embed fonts/Roboto-Regular.ttf
var defaultFont []byte

type SchedulePDFExporter struct{}

func NewSchedulePDFExporter() *SchedulePDFExporter {
	return &SchedulePDFExporter{}
}

func (e *SchedulePDFExporter) Export(view *model.StudentScheduleView, w io.Writer) (int64, error) {
	pdf := gopdf.GoPdf{}

	// A4 Dimensions: 595.28 x 841.89 points
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})
	pdf.AddPage()

	if err := pdf.AddTTFFontData("custom-font", defaultFont); err != nil {
		return 0, fmt.Errorf("failed to register TTF font: %w", err)
	}

	const (
		startX    float64 = 30
		pageWidth float64 = 535 // 595.28 - 60 (márgenes)
	)

	var currentY float64 = 0

	// ---------------------------------------------------------
	// 1. Header Principal / Brand Banner
	// ---------------------------------------------------------
	pdf.SetFillColor(15, 23, 42) // Slate 900
	_ = pdf.Rectangle(0, 0, 595.28, 65, "F", 0, 2)

	_ = pdf.SetFont("custom-font", "", 18)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(startX, 18)
	_ = pdf.Cell(nil, "POLIPLANNER")

	_ = pdf.SetFont("custom-font", "", 10)
	pdf.SetTextColor(148, 163, 184) // Slate 400
	pdf.SetXY(startX, 40)
	_ = pdf.Cell(nil, "Horario de Clases y Calendario de Exámenes")

	currentY = 85

	// ---------------------------------------------------------
	// 2. Horario Semanal
	// ---------------------------------------------------------
	currentY = drawSectionTitle(&pdf, startX, currentY, "1. HORARIO SEMANAL")

	days := []struct {
		Name  string
		Slots []model.ClassSlotView
	}{
		{"Lunes", view.Weekly.Monday},
		{"Martes", view.Weekly.Tuesday},
		{"Miércoles", view.Weekly.Wednesday},
		{"Jueves", view.Weekly.Thursday},
		{"Viernes", view.Weekly.Friday},
		{"Sábado", view.Weekly.Saturday},
	}

	for _, day := range days {
		if len(day.Slots) == 0 {
			continue
		}

		neededHeight := 22.0 + float64(len(day.Slots))*20.0
		currentY = checkPageBreak(&pdf, currentY, neededHeight)

		// Cabecera del Día
		pdf.SetFillColor(241, 245, 249) // Slate 100
		_ = pdf.Rectangle(startX, currentY, startX+pageWidth, currentY+22, "F", 0, 2)

		pdf.SetFillColor(37, 99, 235) // Blue 600 (Barra acentuada)
		_ = pdf.Rectangle(startX, currentY, startX+3, currentY+22, "F", 0, 2)

		_ = pdf.SetFont("custom-font", "", 10)
		pdf.SetTextColor(30, 41, 59)
		pdf.SetXY(startX+10, currentY+5)
		_ = pdf.Cell(nil, strings.ToUpper(day.Name))
		currentY += 22

		// Filas de Clases
		_ = pdf.SetFont("custom-font", "", 9)
		for i, slot := range day.Slots {
			if i%2 == 0 {
				pdf.SetFillColor(255, 255, 255)
			} else {
				pdf.SetFillColor(248, 250, 252)
			}
			_ = pdf.Rectangle(startX, currentY, startX+pageWidth, currentY+20, "F", 0, 2)

			timeStr := fmt.Sprintf("%02d:%02d - %02d:%02d",
				slot.Time.Start.Hour(), slot.Time.Start.Minute(),
				slot.Time.End.Hour(), slot.Time.End.Minute(),
			)

			pdf.SetTextColor(100, 116, 139)
			pdf.SetXY(startX+10, currentY+5)
			_ = pdf.Cell(nil, timeStr)

			pdf.SetTextColor(15, 23, 42)
			pdf.SetXY(startX+110, currentY+5)
			_ = pdf.Cell(nil, slot.Course)

			pdf.SetTextColor(71, 85, 105)
			pdf.SetXY(startX+430, currentY+5)
			roomText := slot.Room
			if roomText != "" {
				roomText = "Aula: " + roomText
			}
			_ = pdf.Cell(nil, roomText)

			pdf.SetStrokeColor(226, 232, 240)
			pdf.SetLineWidth(0.5)
			pdf.Line(startX, currentY+20, startX+pageWidth, currentY+20)

			currentY += 20
		}
		currentY += 10
	}

	// ---------------------------------------------------------
	// 3. Sección Exámenes (Con Separación Clara por Instancia)
	// ---------------------------------------------------------
	currentY += 10
	currentY = checkPageBreak(&pdf, currentY, 40)
	currentY = drawSectionTitle(&pdf, startX, currentY, "2. CALENDARIO DE EXÁMENES")

	examGroups := []struct {
		Title string
		Exams []model.ExamSlotView
	}{
		{"1° Parcial", view.Exams.Partial1},
		{"2° Parcial", view.Exams.Partial2},
		{"1° Final", view.Exams.Final1},
		{"2° Final", view.Exams.Final2},
	}

	hasExams := false
	for _, group := range examGroups {
		if len(group.Exams) > 0 {
			hasExams = true
			break
		}
	}

	if hasExams {
		for _, group := range examGroups {
			if len(group.Exams) == 0 {
				continue
			}

			// Validar si el bloque de la instancia entra (Header + Cabecera Tabla + al menos 1 fila)
			currentY = checkPageBreak(&pdf, currentY, 56) // Al menos header + 1 fila

			// --- Sub-Header de Instancia (ej. "1° PARCIAL") ---
			pdf.SetFillColor(238, 242, 255) // Indigo 50
			_ = pdf.Rectangle(startX, currentY, startX+pageWidth, currentY+20, "F", 0, 2)

			pdf.SetFillColor(79, 70, 229) // Indigo 600 (Barra lateral)
			_ = pdf.Rectangle(startX, currentY, startX+3, currentY+20, "F", 0, 2)

			_ = pdf.SetFont("custom-font", "", 9)
			pdf.SetTextColor(67, 56, 202) // Indigo 700
			pdf.SetXY(startX+10, currentY+5)
			_ = pdf.Cell(nil, fmt.Sprintf("%s", strings.ToUpper(group.Title)))
			currentY += 20

			// --- Cabecera de Columnas del Sub-grupo ---
			pdf.SetFillColor(248, 250, 252) // Slate 50
			_ = pdf.Rectangle(startX, currentY, startX+pageWidth, currentY+16, "F", 0, 2)

			_ = pdf.SetFont("custom-font", "", 8)
			pdf.SetTextColor(100, 116, 139)
			pdf.SetXY(startX+10, currentY+4)
			_ = pdf.Cell(nil, "ASIGNATURA")
			pdf.SetXY(startX+210, currentY+4)
			_ = pdf.Cell(nil, "FECHA / HORA")
			pdf.SetXY(startX+340, currentY+4)
			_ = pdf.Cell(nil, "AULA")
			pdf.SetXY(startX+405, currentY+4)
			_ = pdf.Cell(nil, "REVISIÓN")
			currentY += 16

			// --- Filas de Exámenes ---
			_ = pdf.SetFont("custom-font", "", 8)
			for i, exam := range group.Exams {
				currentY = checkPageBreak(&pdf, currentY, 20)

				if i%2 == 0 {
					pdf.SetFillColor(255, 255, 255)
				} else {
					pdf.SetFillColor(248, 250, 252)
				}
				_ = pdf.Rectangle(startX, currentY, startX+pageWidth, currentY+20, "F", 0, 2)

				// Materia
				pdf.SetTextColor(15, 23, 42)
				pdf.SetXY(startX+10, currentY+5)
				_ = pdf.Cell(nil, truncateText(exam.CourseName, 32))

				// Fecha
				pdf.SetTextColor(51, 65, 85)
				pdf.SetXY(startX+210, currentY+5)
				_ = pdf.Cell(nil, exam.Date)

				// Aula
				pdf.SetTextColor(71, 85, 105)
				pdf.SetXY(startX+340, currentY+5)
				_ = pdf.Cell(nil, exam.Room)

				// Fecha Revisión
				pdf.SetTextColor(100, 116, 139)
				pdf.SetXY(startX+405, currentY+5)
				revText := exam.Revision
				if revText == "" {
					revText = "-"
				}
				_ = pdf.Cell(nil, truncateText(revText, 20))

				// Línea Divisora
				pdf.SetStrokeColor(226, 232, 240)
				pdf.SetLineWidth(0.5)
				pdf.Line(startX, currentY+20, startX+pageWidth, currentY+20)

				currentY += 20
			}
			currentY += 12 // Espaciado entre sub-grupos de exámenes
		}
	} else {
		pdf.SetTextColor(100, 116, 139)
		pdf.SetXY(startX+10, currentY)
		_ = pdf.Cell(nil, "No hay exámenes programados registrados.")
		currentY += 15
	}

	// ---------------------------------------------------------
	// 4. Detalle de Asignaturas y Docentes
	// ---------------------------------------------------------
	currentY += 15
	currentY = checkPageBreak(&pdf, currentY, 40)
	currentY = drawSectionTitle(&pdf, startX, currentY, "3. SECCIONES")

	for _, course := range view.Info {
		cardHeight := 30.0 + float64(len(course.Teachers))*16.0
		currentY = checkPageBreak(&pdf, currentY, cardHeight)

		// Tarjeta Curso
		pdf.SetFillColor(248, 250, 252)
		pdf.SetStrokeColor(226, 232, 240)
		pdf.SetLineWidth(0.8)
		_ = pdf.Rectangle(startX, currentY, startX+pageWidth, currentY+cardHeight, "FD", 0, 2)

		_ = pdf.SetFont("custom-font", "", 10)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(startX+12, currentY+8)
		_ = pdf.Cell(nil, course.Name)

		_ = pdf.SetFont("custom-font", "", 8)
		pdf.SetTextColor(100, 116, 139)
		tagText := fmt.Sprintf("Sección %s  |  Turno %s", course.Section, course.Shift)
		pdf.SetXY(startX+380, currentY+8)
		_ = pdf.Cell(nil, tagText)

		// Docentes
		_ = pdf.SetFont("custom-font", "", 8)
		teacherY := currentY + 24
		for _, teacher := range course.Teachers {
			pdf.SetTextColor(71, 85, 105)
			pdf.SetXY(startX+12, teacherY)

			tInfo := teacher.Name
			if teacher.Email != "" {
				tInfo += fmt.Sprintf(" (%s)", teacher.Email)
			}
			_ = pdf.Cell(nil, "• Docente: "+tInfo)
			teacherY += 14
		}

		currentY += cardHeight + 10
	}

	return pdf.WriteTo(w)
}

// ---------------------------------------------------------
// Helpers
// ---------------------------------------------------------

func drawSectionTitle(pdf *gopdf.GoPdf, x, y float64, title string) float64 {
	pdf.SetFillColor(37, 99, 235)
	_ = pdf.Rectangle(x, y, x+4, y+14, "F", 0, 2)

	_ = pdf.SetFont("custom-font", "", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetXY(x+10, y)
	_ = pdf.Cell(nil, title)

	return y + 22
}

func checkPageBreak(pdf *gopdf.GoPdf, currentY float64, neededHeight float64) float64 {
	const pageBottomLimit float64 = 770
	if currentY+neededHeight > pageBottomLimit {
		pdf.AddPage()
		return 40
	}
	return currentY
}

func truncateText(s string, maxChars int) string {
	if len(s) > maxChars {
		return s[:maxChars-3] + "..."
	}
	return s
}
