package teacher

type TeacherID int64

type Teacher struct {
	ID        TeacherID
	FirstName string
	LastName  string
	Email     string
	searchKey string
}

func (t Teacher) GetSearchKey() string {
	return t.searchKey
}
