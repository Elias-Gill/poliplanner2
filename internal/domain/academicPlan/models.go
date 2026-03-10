package academicPlan

type AcademicPlanID int64

type Subject struct {
	Name       string
	Department string
	Semester   int
	Level      int
}

type career struct {
	// acronym of the career, should be all capitalized
	Name string
}

// Our root aggreate. Represents the academic plan of a career.
type AcademicPlan struct {
	Subjects []Subject
	Career   career
}

func NewAcademicPlan(careerCode string) AcademicPlan {
	return AcademicPlan{
		Career: career{
			Name: careerCode,
		},
	}
}

func (a *AcademicPlan) AddSubject(name string, department string, semester int, level int) {
	a.Subjects = append(a.Subjects, Subject{
		Name:       name,
		Department: department,
		Semester:   semester,
		Level:      level,
	})
}
