package sqlite

import (
	"database/sql"

	// Repositories
	"github.com/elias-gill/poliplanner2/internal/repository"
	"github.com/elias-gill/poliplanner2/internal/repository/academic"
	auth "github.com/elias-gill/poliplanner2/internal/repository/auth"
	"github.com/elias-gill/poliplanner2/internal/repository/excel"
	"github.com/elias-gill/poliplanner2/internal/repository/schedule"
	"github.com/elias-gill/poliplanner2/internal/repository/user"

	// Sqlite implementations
	academicImpl "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/academic"
	authImpl "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/auth"
	excelImpl "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/excel"
	scheduleImpl "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/schedule"
	txManImpl "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	userImpl "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/user"
)

type SQLiteStorage struct {
	TxManager repository.TxManager

	// Excel Repositories
	ExcelRepo excel.ExcelRepository
	SyncRepo  excel.SyncRepository

	// Academic Repositories
	CareerRepo     academic.CareerRepository
	CourseRepo     academic.CourseRepository
	PeriodRepo     academic.PeriodRepository
	SubjectRepo    academic.SubjectRepository
	TeacherRepo    academic.TeacherRepository
	CurriculumRepo academic.CurriculumRepository

	ScheduleRepo schedule.ScheduleRepository

	// Auth and authentication
	UserRepo user.UserRepository
	AuthRepo auth.AuthRepository
}

// NewSQLiteStorage initializes and returns a pointer to SQLiteStorage.
// Returning a pointer is idiomatic in Go for component containers/heavy structs.
func NewSQLiteStorage(conn *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{
		TxManager: txManImpl.NewSQLTxManager(conn),

		// Excel
		ExcelRepo: excelImpl.NewExcelRepository(conn),
		SyncRepo:  excelImpl.NewSyncRepository(conn),

		// Academic
		CareerRepo:     academicImpl.NewCareerRepository(conn),
		CourseRepo:     academicImpl.NewCourseRepository(conn),
		PeriodRepo:     academicImpl.NewPeriodRepository(conn),
		SubjectRepo:    academicImpl.NewSubjectRepository(conn),
		TeacherRepo:    academicImpl.NewTeacherRepository(conn),
		CurriculumRepo: academicImpl.NewCurriculumRepository(conn),

		ScheduleRepo: scheduleImpl.NewScheduleRepository(conn),

		// Auth and authorization
		UserRepo: userImpl.NewUserRepository(conn),
		AuthRepo: authImpl.NewAuthRepository(conn),
	}
}
