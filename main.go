package main

import (
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

	log.InitLogger(config.Get().VerboseLogs)
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
	r.Route("/schedule", router.NewSchedulesRouter(services.SubjectService, services.ScheduleService, services.SheetVersionService, services.CareerService))
	r.Route("/excel", router.NewExcelRouter(services.ExcelService))
	r.Route("/misc", router.NewMiscRouter())
	r.Route("/guides", router.NewGuidesRouter())
	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	// 404 - Not found
	r.NotFound(router.NotFoundHandler)

	// Start Server
	log.Info("Server is running", "addr", cfg.ServerAddr)
	err = http.ListenAndServe(cfg.ServerAddr, r)
	if err != nil {
		panic(err)
	}
}
