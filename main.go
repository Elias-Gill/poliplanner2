package main

import (
	"context"
	"net/http"
	"path"

	service "github.com/elias-gill/poliplanner2/internal/app"
	"github.com/elias-gill/poliplanner2/internal/config"
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
	"github.com/elias-gill/poliplanner2/internal/http/routes/auth"
	"github.com/elias-gill/poliplanner2/internal/http/routes/dashboard"
	"github.com/elias-gill/poliplanner2/internal/http/routes/excel"
	"github.com/elias-gill/poliplanner2/internal/http/routes/guides"
	"github.com/elias-gill/poliplanner2/internal/http/routes/schedules"
	"github.com/elias-gill/poliplanner2/internal/http/routes/tools"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite"
	log "github.com/elias-gill/poliplanner2/logger"

	"github.com/go-chi/chi/v5"
)

func main() {
	config.MustLoad()
	cfg := config.Get()

	log.InitLogger(cfg.Logging.Verbose)
	log.Info("Logger initialized", "verbose", cfg.Logging.Verbose)
	log.Info("Loading env configuraion")

	log.Debug("Initializing db")
	conn, err := persistence.InitDB()
	if err != nil {
		panic(err)
	}
	defer conn.CloseDB()

	services := service.NewServices(
		sqlite.NewSqliteUserStore(conn.GetConnection()),
		sqlite.NewSqliteSheetVersionStore(conn.GetConnection()),
		sqlite.NewSqliteExcelImportStorer(conn.GetConnection()),
		sqlite.NewSqliteScheduleStore(conn.GetConnection()),
		sqlite.NewSqliteAcademicPlanStorer(conn.GetConnection()),
		sqlite.NewSqliteCourseOfferingStore(conn.GetConnection()),
		sqlite.NewSqliteSessionStore(conn.GetConnection()),
	)

	r := chi.NewRouter()

	// Register middlewares
	r.Use(middleware.NewSessionMiddleware(services.AuthService))

	// REFACTOR: separate special routes into more routers
	// login, special pages and auth router
	r.Route("/", auth.NewAuthRouter(services.UserService, services.AuthService, services.EmailService))

	r.Route("/dashboard", dashboard.NewDashboardRouter(services.ScheduleService, services.AcademicPlanService))
	r.Route("/schedule", schedules.NewSchedulesRouter(services.ScheduleService, services.AcademicPlanService))

	// User administration router
	// r.Route("/user", router.NewUserRouter(services.UserService, services.AuthService))

	// Misc routers
	r.Route("/tools", tools.NewToolsRouter())
	r.Route("/guides", guides.NewGuidesRouter())
	// r.Route("/courses", router.NewCourseRouter(services.CoursesService, services.CareerService))

	// Admin routers
	r.Route("/excel", excel.NewExcelRouter(services.ImportService))

	// Static files
	staticDir := http.Dir(path.Join(config.Get().Paths.BaseDir, "web", "static"))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))
	r.Handle("/sitemap.xml", http.FileServer(staticDir))
	r.Handle("/robots.txt", http.FileServer(staticDir))
	r.Handle("/favicon.ico", http.FileServer(staticDir))

	// 404 - Not found
	r.NotFound(NotFoundHandler)

	// Auto import new excel versions on startup
	go func() {
		// 30 seconds has to be more than enough, even when google drive is slow
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Excel.ScraperTimeout)
		defer cancel()
		// The result of this operation is irrelevant
		services.ImportService.AutoSync(ctx)
	}()

	// Start Server
	log.Info("Server is running", "addr", cfg.Server.Addr)
	err = http.ListenAndServe(cfg.Server.Addr, r)
	if err != nil {
		panic(err)
	}
}

// REFACTOR: que mierda hace esto aca
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages")
	w.Header().Set("Content-Type", "text/html")

	utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "404.html")).Execute(w, nil)
}
