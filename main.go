package main

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"

	log "github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/router"
)

func main() {
	cfg := config.Load()
	log.InitLogger(cfg.VerboseLogs)
	log.Info("Loading env configuraion")

	log.Debug("Initializing db")
	err := db.InitDB(cfg)
	if err != nil {
		panic(err)
	}
	defer db.CloseDB()

	service.InitializeServices(
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
	r.Use(service.SessionMiddleware)

	r.Route("/", router.NewAuthRouter())
	r.Route("/dashboard", router.NewDashboardRouter())
	r.Route("/schedule", router.NewSchedulesRouter())
	r.Route("/excel", router.NewAuthRouter())
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
