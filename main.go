package main

import (
	"context"
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
	"github.com/elias-gill/poliplanner2/internal/http/render"
	"github.com/elias-gill/poliplanner2/internal/http/routes"
	"github.com/elias-gill/poliplanner2/internal/http/routes/auth"
	"github.com/elias-gill/poliplanner2/internal/http/routes/dashboard"
	"github.com/elias-gill/poliplanner2/internal/http/routes/excel"
	"github.com/elias-gill/poliplanner2/internal/http/routes/guides"
	"github.com/elias-gill/poliplanner2/internal/http/routes/schedules"
	"github.com/elias-gill/poliplanner2/internal/http/routes/tools"
	"github.com/elias-gill/poliplanner2/internal/http/routes/user"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence"
	log "github.com/elias-gill/poliplanner2/logger"

	services "github.com/elias-gill/poliplanner2/internal/service"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite"

	"github.com/go-chi/chi/v5"
)

func main() {
	log.Info("Loading env configurations")
	cfg, err := config.Load()
	if err != nil {
		log.Error("Config load has errors", "err", err.Error())
		return
	}

	log.InitLogger(cfg.Logging.Verbose)
	log.Info("Logger initialized", "verbose", cfg.Logging.Verbose)

	log.Info("Initializing db")
	conn, err := persistence.ConnectDB()
	if err != nil {
		panic(err)
	}
	defer conn.CloseDB()

	log.Info("Running migrations")
	err = persistence.RunMigrations()
	if err != nil {
		panic(err)
	}

	// Instantiate repositories
	sqliteStore := sqlite.NewSQLiteStorage(conn.GetConnection())

	// Start services
	servs := services.NewAppServices(services.RepositoriesInput{
		ExcelRepo:      sqliteStore.ExcelRepo,
		SyncRepo:       sqliteStore.SyncRepo,
		CourseRepo:     sqliteStore.CourseRepo,
		TeacherRepo:    sqliteStore.TeacherRepo,
		CurriculumRepo: sqliteStore.CurriculumRepo,
		PeriodRepo:     sqliteStore.PeriodRepo,
		SubjectRepo:    sqliteStore.SubjectRepo,
		CareerRepo:     sqliteStore.CareerRepo,
		AuthRepo:       sqliteStore.AuthRepo,
		UserRepo:       sqliteStore.UserRepo,
		TxManager:      sqliteStore.TxManager,
		ScheduleRepo:   sqliteStore.ScheduleRepo,
	})

	// Setup http routers
	r := initRouter(servs)

	// Auto import new excel versions on startup (concurrently)
	go func() {
		// 30 seconds has to be more than enough, even when google drive is slow
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Excel.ScraperTimeout)
		defer cancel()
		// The result of this operation is irrelevant
		servs.SyncService.AutoSync(ctx)
	}()

	// Start Server
	log.Info("Server is running", "addr", cfg.Server.Addr)
	err = http.ListenAndServe(cfg.Server.Addr, r)
	if err != nil {
		panic(err)
	}
}

// initRouter builds, configs middlewares, and maps all the routes for the application.
func initRouter(srvs *services.AppServices) chi.Router {
	// Start template manager
	tmplManager, err := render.NewTemplateManager(config.Get().Paths.TemplatesDir)
	if err != nil {
		log.Fatal("Error al cargar las plantillas", "error", err)
	}

	r := chi.NewRouter()

	// Register middlewares
	r.Use(middleware.NewSessionMiddleware(srvs.SessionService))

	r.Mount("/", auth.NewHandler(tmplManager, srvs.UserService, srvs.SessionService, srvs.EmailService).Routes())

	r.Mount("/dashboard", dashboard.NewHandler(tmplManager, srvs.ScheduleService, srvs.CourseService).Routes())

	r.Mount("/schedule", schedules.NewHandler(
		tmplManager,
		srvs.ScheduleService,
		srvs.CareerService,
		srvs.CourseService,
		srvs.CurriculumService,
	).Routes())

	r.Mount("/user", user.NewHandler(tmplManager, srvs.SessionService).Routes())

	// Misc routers
	r.Mount("/tools", tools.NewHandler(tmplManager).Routes())
	r.Mount("/guides", guides.NewHandler(tmplManager).Routes())

	// Admin routers
	r.Mount("/excel", excel.NewHandler(tmplManager, srvs.ExcelService, srvs.SyncService).Routes())

	// Static files and assets mapping
	staticDir := http.Dir(config.Get().Paths.AssetsDir)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))
	r.Handle("/sitemap.xml", http.FileServer(staticDir))
	r.Handle("/robots.txt", http.FileServer(staticDir))
	r.Handle("/service_worker.js", http.FileServer(staticDir))
	r.Handle("/favicon.ico", http.FileServer(staticDir))

	// Fallback 404 handler
	r.NotFound(routes.NotFound(tmplManager))

	return r
}
