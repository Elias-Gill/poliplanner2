package service

import (
	// App layer
	apApp "github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	emailApp "github.com/elias-gill/poliplanner2/internal/app/email"
	excelApp "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	scheduleApp "github.com/elias-gill/poliplanner2/internal/app/schedule"
	sheetVersionApp "github.com/elias-gill/poliplanner2/internal/app/sheetVersion"
	userApp "github.com/elias-gill/poliplanner2/internal/app/user"

	// Domain layer
	// "github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	// "github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
	"github.com/elias-gill/poliplanner2/internal/domain/user"

	"github.com/elias-gill/poliplanner2/internal/config"
)

type Services struct {
	UserService         *userApp.UserService
	SheetVersionService *sheetVersionApp.SheetVersionService
	ScheduleService     *scheduleApp.ScheduleService
	ImportService       *excelApp.ImportService
	EmailService        *emailApp.EmailService
	AcademicPlanService *apApp.AcademicPlanService
}

// Convenience function to instantiate all the services in one call
func NewServices(
	userStore user.UserStorer,
	sheetVersionStore sheetVersion.SheetVersionStorer,
	importStorer excelApp.ImportStorer,
	scheduleStore schedule.ScheduleStorer,
	planStorer academicPlan.AcademicPlanStorer,
	courseStorer courseOffering.CourseStorer,
) *Services {
	return &Services{
		UserService:         userApp.NewUserService(userStore),
		SheetVersionService: sheetVersionApp.NewSheetVersionService(sheetVersionStore),
		ImportService:       excelApp.NewExcelImportService(importStorer, sheetVersionStore),
		EmailService:        emailApp.NewEmailService(config.Get().Email.APIKey),
		ScheduleService:     scheduleApp.NewScheduleService(scheduleStore),
		AcademicPlanService: apApp.NewAcademicPlanService(planStorer, courseStorer),
	}
}
