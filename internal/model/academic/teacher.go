package academic

type TeacherID int64

type Teacher struct {
	ID        TeacherID
	Title     string
	FirstName string
	LastName  string
	Email     string
}
