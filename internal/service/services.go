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
	AcademicPlanService *AcademicPlanService
}

// Convenience function to instantiate all the services in one call
func NewServices(
	userStore store.UserStorer,
	sheetVersionStore store.SheetVersionStorer,
	courseStore store.CourseStorer,
	scheduleStore store.ScheduleStorer,
	careerStore store.CareerStorer,
	academicPlanStore store.AcademicPlanStore,
	periodStore store.PeriodStore,
	emailApiKey string,
) *Services {
	return &Services{
		UserService:         NewUserService(userStore),
		SheetVersionService: NewSheetVersionService(sheetVersionStore),
		ExcelService:        NewExcelService(sheetVersionStore, courseStore),
		ScheduleService:     NewScheduleService(scheduleStore),
		AcademicPlanService: NewAcademicPlanService(academicPlanStore),
		EmailService:        NewEmailService(emailApiKey),
	}
}
