package service

import (
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/repository"
	"github.com/elias-gill/poliplanner2/internal/repository/academic"
	"github.com/elias-gill/poliplanner2/internal/repository/auth"
	"github.com/elias-gill/poliplanner2/internal/repository/excel"
	"github.com/elias-gill/poliplanner2/internal/repository/schedule"
	"github.com/elias-gill/poliplanner2/internal/repository/user"

	academicSrv "github.com/elias-gill/poliplanner2/internal/service/academic"
	authSrv "github.com/elias-gill/poliplanner2/internal/service/auth"
	"github.com/elias-gill/poliplanner2/internal/service/email"
	excelSrv "github.com/elias-gill/poliplanner2/internal/service/excel"
	scheduleSrv "github.com/elias-gill/poliplanner2/internal/service/schedule"
	userSrv "github.com/elias-gill/poliplanner2/internal/service/user"
)

// AppServices centralizes all instantiated application services.
type AppServices struct {
	// Academic
	PeriodService     *academicSrv.PeriodService
	CourseService     *academicSrv.CourseService
	CurriculumService *academicSrv.CurriculumService
	CareerService     *academicSrv.CareerService

	ExcelService    *excelSrv.ExcelService
	SyncService     *excelSrv.SyncService
	UserService     *userSrv.UserService
	SessionService  *authSrv.SessionService
	EmailService    *email.EmailSender
	ScheduleService *scheduleSrv.ScheduleService
}

// RepositoriesInput groups the required interfaces to build the services.
type RepositoriesInput struct {
	// Academic repos
	CourseRepo     academic.CourseRepository
	TeacherRepo    academic.TeacherRepository
	CurriculumRepo academic.CurriculumRepository
	PeriodRepo     academic.PeriodRepository
	SubjectRepo    academic.SubjectRepository
	CareerRepo     academic.CareerRepository

	// Parsing repos
	ExcelRepo excel.ExcelRepository
	SyncRepo  excel.SyncRepository

	// Schedules
	ScheduleRepo schedule.ScheduleRepository

	// Authorization and authentication
	AuthRepo auth.AuthRepository
	UserRepo user.UserRepository

	TxManager repository.TxManager
}

// NewAppServices builds and wires all services together, managing inter-service dependencies.
func NewAppServices(repos RepositoriesInput) *AppServices {
	// Services that don't depend on other services
	periodService := academicSrv.NewPeriodService(repos.PeriodRepo)
	userService := userSrv.NewUserService(repos.UserRepo)
	authService := authSrv.NewSessionService(repos.UserRepo, repos.AuthRepo)
	emailService := email.New(config.Get().Email.APIKey)

	// Services that depend on previously created services
	excelService := excelSrv.NewExcelService(
		repos.ExcelRepo,
		repos.CourseRepo,
		repos.TeacherRepo,
		repos.CurriculumRepo,
		repos.PeriodRepo,
		repos.SubjectRepo,
		repos.CareerRepo,
		repos.TxManager,
		periodService,
	)

	syncService := excelSrv.NewSyncService(
		excelSrv.NewDiscoveryService(config.Get().Excel.GoogleAPIKey),
		excelService,
		repos.SyncRepo,
	)

	courseService := academicSrv.NewCourseService(repos.CourseRepo, periodService)

	curriculumService := academicSrv.NewCurriculumService(repos.CurriculumRepo, repos.CareerRepo)

	careerService := academicSrv.NewCareerService(repos.CareerRepo)

	scheduleService := scheduleSrv.New(repos.ScheduleRepo)

	return &AppServices{
		// Academic
		PeriodService:     periodService,
		CourseService:     courseService,
		CurriculumService: curriculumService,
		CareerService:     careerService,

		// Parsing
		ExcelService: excelService,
		SyncService:  syncService,

		// User
		SessionService:  authService,
		UserService:     userService,
		ScheduleService: scheduleService,

		// Misc
		EmailService: emailService,
	}
}
