package main

import (
	"context"
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/auth"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	log "github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/router"

	"github.com/go-chi/chi/v5"
)

func main() {
	config.MustLoad()
	cfg := config.Get()

	log.InitLogger(config.Get().Logging.Verbose)
	log.Info("Loading env configuraion")

	log.Debug("Initializing db")
	err := db.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.CloseDB()

	services := service.NewServices(
		db.GetConnection(),
		store.NewSqliteUserStore(),
		store.NewSqliteSheetVersionStore(),
		store.NewSqliteCareerStore(),
		store.NewSqliteSubjectStore(),
		store.NewSqliteScheduleStore(),
		store.NewSqliteScheduleDetailStore(),
	)

	// Configure http server
	r := chi.NewRouter()
	r.Use(auth.SessionMiddleware)

	r.Route("/", router.NewAuthRouter(services.UserService))
	r.Route("/dashboard", router.NewDashboardRouter(services.ScheduleService))
	r.Route("/subject", router.NewSubjectRouter(services.SubjectService, services.SheetVersionService, services.CareerService))
	r.Route("/schedule", router.NewSchedulesRouter(services.SubjectService, services.ScheduleService, services.SheetVersionService, services.CareerService))
	r.Route("/excel", router.NewExcelRouter(services.ExcelService))
	r.Route("/misc", router.NewMiscRouter())
	r.Route("/guides", router.NewGuidesRouter())

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	r.Handle("/sitemap.xml", http.FileServer(http.Dir("./web/static")))
	r.Handle("/robots.txt", http.FileServer(http.Dir("./web/static")))
	r.Handle("/favicon.ico", http.FileServer(http.Dir("./web/static")))

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
