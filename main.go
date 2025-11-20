package main

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/config"
	"github.com/elias-gill/poliplanner2/db"
	"github.com/elias-gill/poliplanner2/db/store"
	"github.com/elias-gill/poliplanner2/service"
	"github.com/go-chi/chi/v5"

	"github.com/elias-gill/poliplanner2/controller"
	log "github.com/elias-gill/poliplanner2/logger"
)

func main() {
	log.Logger.Info("Loading env configuraion")
	cfg := config.Load()

	log.Logger.Debug("Initializing db")
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
	r.Use(service.SessionMidleware)

	r.Route("/user", controller.NewUsersRouter()).
		Route("/schedules", controller.NewUsersRouter()).
		Route("/excel", controller.NewUsersRouter()).
		Route("/misc", controller.NewUsersRouter()).
		Route("/help", controller.NewUsersRouter())

	// Start Server
	http.ListenAndServe(":3000", r)

	defer db.CloseDB()
}
