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
	EmailService        *EmailService
}

// Convenience function to instantiate all the services in one call
func NewServices(
	conn *sql.DB,
	userStore store.UserStorer,
	sheetVersionStore store.SheetVersionStorer,
	sheetVersionCheckStore store.SheetVersionCheckStorer,
	careerStore store.CareerStorer,
	subjectStore store.SubjectStorer,
	scheduleStore store.ScheduleStorer,
	scheduleDetailStore store.ScheduleDetailStorer,
	emailApiKey string,
) *Services {
	return &Services{
		UserService:         NewUserService(conn, userStore),
		SheetVersionService: NewSheetVersionService(conn, sheetVersionStore),
		CareerService:       NewCareerService(conn, careerStore),
		SubjectService:      NewSubjectService(conn, subjectStore),
		ScheduleService:     NewScheduleService(conn, scheduleStore, scheduleDetailStore, sheetVersionStore, subjectStore),
		ExcelService:        NewExcelService(conn, sheetVersionStore, sheetVersionCheckStore, careerStore, subjectStore),
		EmailService:        NewEmailService(emailApiKey),
	}
}
