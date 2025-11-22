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
	log.GetLogger().Info("Loading env configuraion")

	log.GetLogger().Debug("Initializing db")
	err := db.InitDB(cfg)
	if err != nil {
		panic(err.Error())
	}
	conn := db.GetConnection()

	service.InitializeServices(
		store.NewSqliteUserStore(conn),
		store.NewSqliteSheetVersionStore(conn),
		store.NewSqliteCareerStore(conn),
		store.NewSqliteSubjectStore(conn),
		store.NewSqliteScheduleStore(conn),
		store.NewSqliteScheduleDetailStore(conn),
	)

	// Configure http server
	r := chi.NewRouter()
	r.Use(service.SessionMiddleware)

	r.Route("/", router.NewAuthRouter())
	r.Route("/schedule", router.NewAuthRouter())
	r.Route("/excel", router.NewAuthRouter())
	r.Route("/misc", router.NewMiscRouter())
	r.Route("/guides", router.NewGuidesRouter())
	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Start Server
	log.GetLogger().Info("Server runnign in port :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err.Error())
	}

	defer db.CloseDB()
}
