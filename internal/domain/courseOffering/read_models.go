package courseOffering

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

// Light weight grade info, used to optimize database and network usage when listing a lot of
// grades
type CourseListItem struct {
	ID          int64
	CourseName  string
	SubjectName string
	Section     string
	Semester    int
	Teachers    []int64 // references teachers
}

// ==========================================================
// =                        UTILS                           =
// ==========================================================

type SubjectGroup struct {
	SubjectName string
	Courses     []*CourseListItem
}

type SemesterGroup struct {
	Semester int
	Subjects []SubjectGroup
}
