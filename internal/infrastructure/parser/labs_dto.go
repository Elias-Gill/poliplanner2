package parser

type LaboratoryDTO struct {
	// Curriculum identification
	RawName string
	Plan    string

	// E.g.: T1, T2, T3, etc. When no section is found, "UNICA" is used as default.
	Section string

	// Academic period the laboratory belongs to. Used later by the service to exclude trash
	// data from other periods.
	Semester int

	WeekSchedule [7]WeekDayData
}
