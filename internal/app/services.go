package service

import (
	email "github.com/elias-gill/poliplanner2/internal/app/email"
	excelimport "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	sApp "github.com/elias-gill/poliplanner2/internal/app/schedule"
	svApp "github.com/elias-gill/poliplanner2/internal/app/sheetVersion"
	uApp "github.com/elias-gill/poliplanner2/internal/app/user"

	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	sheetversion "github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
	"github.com/elias-gill/poliplanner2/internal/domain/user"

	"github.com/elias-gill/poliplanner2/internal/config"
)

type Services struct {
	UserService         *uApp.UserService
	SheetVersionService *svApp.SheetVersionService
	ScheduleService     *sApp.ScheduleService
	ImportService       *excelimport.ImportService
	EmailService        *email.EmailService
}

// Convenience function to instantiate all the services in one call
func NewServices(
	userStore user.UserStorer,
	sheetVersionStore sheetversion.SheetVersionStorer,
	importStorer excelimport.ImportStorer,
	scheduleStore schedule.ScheduleStorer,
) *Services {
	return &Services{
		UserService:         uApp.NewUserService(userStore),
		SheetVersionService: svApp.NewSheetVersionService(sheetVersionStore),
		ImportService:       excelimport.NewExcelImportService(importStorer, sheetVersionStore),
		EmailService:        email.NewEmailService(config.Get().Email.APIKey),
		ScheduleService:     sApp.NewScheduleService(scheduleStore),
	}
}
