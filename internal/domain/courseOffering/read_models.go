package courseOffering

import "github.com/elias-gill/poliplanner2/internal/domain/teacher"

type SectionID int64

type Section struct {
	ID         SectionID
	Section    string
	CourseName string
	Teachers   []teacher.Teacher
	Type       CourseType
}

type OfferList struct {
	Subject string
	Offer   []Section
}
