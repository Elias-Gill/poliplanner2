package academic

type LaboratoryID int64

// Laboratory is a laboratory offering of a subject within a curriculum and
// academic period. Its schedule is independent from the course schedule.
type Laboratory struct {
	ID       LaboratoryID
	Section  string
	Schedule []ClassSession
}
