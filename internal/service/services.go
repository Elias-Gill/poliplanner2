package service

import (
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type Services struct {
	UserService         *UserService
	SheetVersionService *SheetVersionService
	CareerService       *CareerService
	SubjectService      *SubjectService
	ScheduleService     *ScheduleService
	ExcelService        *ExcelService
}

// Convenience function to instantiate all the services in one call
func NewServices(
	conn *sql.DB,
	userStore store.UserStorer,
	sheetVersionStore store.SheetVersionStorer,
	careerStore store.CareerStorer,
	subjectStore store.SubjectStorer,
	scheduleStore store.ScheduleStorer,
	scheduleDetailStore store.ScheduleDetailStorer,
) *Services {
	return &Services{
		UserService:         NewUserService(conn, userStore),
		SheetVersionService: NewSheetVersionService(conn, sheetVersionStore),
		CareerService:       NewCareerService(conn, careerStore),
		SubjectService:      NewSubjectService(conn, subjectStore),
		ScheduleService:     NewScheduleService(conn, scheduleStore, scheduleDetailStore),
		ExcelService:        NewExcelService(conn, sheetVersionStore, careerStore, subjectStore),
	}
}
