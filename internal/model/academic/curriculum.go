package academic

// Plan represents a specific academic study plan offered by the university.
// Multiple plans can coexist for a single career, allowing the same subject
// to have different curriculum configurations across plans.
type Plan struct {
	Code string
}

// Note: The university's official data sources only provide the emphasis code, leaving
// the full name currently loaded via metadata
type Emphasis struct {
	Career CareerID
	Code   string
	Name   string
}

type CurriculumID int64

// Curriculum represents a specific instance of a subject taught within a career. It maps the
// subject to a particular plan and semester/level (depending on the career's structure).
type Curriculum struct {
	ID       CurriculumID
	Semester int
	Level    int
	Plan     Plan
	Emphases []Emphasis
}

type CareerID int64

// Career represents a university degree or program.
//
// Note: The university's official data sources only provide the career code, leaving
// the full name currently loaded via metadata
type Career struct {
	ID   CareerID
	Code string // Unique identifier code used by the university (e.g., "IIN", "LCI").
	Name string // Full career name, retrieved from our metadata files
}

// Subject represents an indivisible academic unit that falls under the responsibility
// of a specific university department.
type SubjectID int64

type Subject struct {
	ID         SubjectID
	Name       string
	Department Department
}

// Department represents an administrative division within the university
// responsible for managing specific subjects.
//
// Note: Much like with Careers, the university's official data sources only provide
// the department code, leaving the full name known via our own metadata.
type DepartmentID int64

type Department struct {
	ID   DepartmentID
	Code string // Unique department identifier code used by the university.
	Name string // Full department name, retrieved from our metadata files
}
