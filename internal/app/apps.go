package service

import (
	// App layer
	apApp "github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/app/auth"
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

type UseCases struct {
	User         *userApp.User
	SheetVersion *sheetVersionApp.SheetVersion
	Schedule     *scheduleApp.Schedule
	ExcelImport  *excelApp.ExcelImporter
	Email        *emailApp.EmailSender
	AcademicPlan *apApp.AcademicPlan
	Auth         *auth.AuthManager
}

// Convenience function to instantiate all the services in one call
func NewUseCases(
	userStore user.UserRepository,
	sheetVersionStore sheetVersion.SheetVersionRepository,
	importStore excelApp.ImportRepository,
	scheduleStore schedule.ScheduleRepository,
	planStore academicPlan.AcademicPlanRepository,
	courseStore courseOffering.CourseRepository,
	sessionStore auth.SessionRepository,
) *UseCases {
	return &UseCases{
		User:         userApp.New(userStore),
		SheetVersion: sheetVersionApp.New(sheetVersionStore),
		ExcelImport:  excelApp.New(importStore, sheetVersionStore),
		Email:        emailApp.New(config.Get().Email.APIKey),
		Schedule:     scheduleApp.New(scheduleStore),
		AcademicPlan: apApp.New(planStore, courseStore),
		Auth:         auth.New(userStore, sessionStore),
	}
}
