package commons

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

type TeacherDTO struct {
	Title     string
	FirstName string
	LastName  string
	Email     string
}
