package service

import (
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type Services struct {
	UserService         *UserService
	SheetVersionService *SheetVersionService
	ScheduleService     *ScheduleService
	ExcelService        *ExcelService
	EmailService        *EmailService
	SubjectService      *GradeService
	CareerService       *CareerService
}

// Convenience function to instantiate all the services in one call
func NewServices(
	userStore store.UserStorer,
	sheetVersionStore store.SheetVersionStorer,
	gradeStore store.GradeStorer,
	scheduleStore store.ScheduleStorer,
	careerStore store.CareerStorer,
	emailApiKey string,
) *Services {
	return &Services{
		UserService:         NewUserService(userStore),
		SheetVersionService: NewSheetVersionService(sheetVersionStore),
		ExcelService:        NewExcelService(sheetVersionStore, gradeStore),
		ScheduleService:     NewScheduleService(scheduleStore),
		CareerService:       NewCareerService(careerStore),
		EmailService:        NewEmailService(emailApiKey),
		SubjectService:      NewSubjectService(gradeStore),
	}
}
