package academicPlan

type SubjectID int64

type CareerID int64

type Career struct {
	ID   CareerID
	Name string
}

// Note: This entity does not represent a standalone subject (as defined in the
// normalized database schema).
//
// Instead, it models the subject as it exists within a specific academic plan.
// In relational terms, this corresponds to the subject associated with a particular
// career through the academic plan structure, which typically involves three tables:
//   - Career,
//   - Subject,
//   - and an intermediate AcademicPlan (or curriculum) table that links them and
//     stores additional attributes such as the semester in which the subject is scheduled.
type Subject struct {
	ID         SubjectID
	Name       string
	Department string
	Level      int
	Semester   int
}

type SemesterSubjects struct {
	Semester int
	Subjects []Subject
}

type AcademicPlan struct {
	CareerID  CareerID
	Semesters []SemesterSubjects
}
