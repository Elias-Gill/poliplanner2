package model

type Section struct {
	ID        int64  // Identificador único de la sección
	Section   string // Código del curso o materia
	Name      string // Nombre del curso
	Professor string // Nombre del profesor
	Type      int    // 0 = curso normal, 1 = examen final
}

type SectionsList struct {
	Assignment string
	Sections   []Section
}
