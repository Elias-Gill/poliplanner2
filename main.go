package main

import (
	"context"
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/auth"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db"
	"github.com/elias-gill/poliplanner2/internal/db/store/sqlite"
	log "github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/router"

	"github.com/go-chi/chi/v5"
)

// FIX: que no se pueda crear nuevo horario cuando no se puede resolver el periodo, ver ese flujo

func main() {
	config.MustLoad()
	cfg := config.Get()

	log.InitLogger(config.Get().Logging.Verbose)
	log.Info("Loading env configuraion")

	log.Debug("Initializing db")
	conn, err := db.InitDB()
	if err != nil {
		panic(err)
	}
	defer conn.CloseDB()

	services := service.NewServices(
		sqlite.NewSqliteUserStore(conn.GetConnection()),
		sqlite.NewSqliteSheetVersionStore(conn.GetConnection()),
		sqlite.NewSqliteCourseStore(conn.GetConnection()),
		sqlite.NewSqliteScheduleStore(conn.GetConnection()),
		sqlite.NewSqliteCareerStore(conn.GetConnection()),
		sqlite.NewSqlitePeriodStore(conn.GetConnection()),
		config.Get().Email.APIKey, // what the fuck is this doing here
	)

	// Configure http server
	r := chi.NewRouter()
	r.Use(auth.SessionMiddleware)

	r.Route("/", router.NewAuthRouter(services.UserService, services.EmailService))
	r.Route("/dashboard", router.NewDashboardRouter(services.ScheduleService))
	r.Route("/courses", router.NewCourseRouter(services.CoursesService, services.CareerService))
	r.Route("/schedule", router.NewSchedulesRouter(services.CoursesService, services.ScheduleService, services.CareerService, services.SheetVersionService))
	r.Route("/user", router.NewUserRouter(services.UserService))
	r.Route("/excel", router.NewExcelRouter(services.ExcelService))
	r.Route("/tools", router.NewMiscRouter())
	r.Route("/guides", router.NewGuidesRouter())

	// Static files
	staticDir := http.Dir(path.Join(config.Get().Paths.BaseDir, "web", "static"))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))
	r.Handle("/sitemap.xml", http.FileServer(staticDir))
	r.Handle("/robots.txt", http.FileServer(staticDir))
	r.Handle("/favicon.ico", http.FileServer(staticDir))

	// 404 - Not found
	r.NotFound(router.NotFoundHandler)

	go func() {
		// 30 seconds has to be more than enough, even when google drive is slow
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Excel.ScraperTimeout)
		defer cancel()
		// The result of this operation is irrelevant
		services.ExcelService.SearchOnStartup(ctx)
	}()

	// Start Server
	log.Info("Server is running", "addr", cfg.Server.Addr)
	err = http.ListenAndServe(cfg.Server.Addr, r)
	if err != nil {
		panic(err)
	}
}
