package academicPlan

type PlanID int64

// Note: This entity does not represent a standalone subject as defined in the
// normalized database schema. Instead, it models the subject as it exists within
// a specific academic plan. In relational terms, this corresponds to the subject
// associated with a particular career through the academic plan structure,
// which typically involves three tables: Career, Subject, and an intermediate
// AcademicPlan (or curriculum) table that links them and stores additional
// attributes such as the semester in which the subject is scheduled.
type Subject struct {
	PlanID     PlanID
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
