package main

import (
	"github.com/elias-gill/poliplanner2/config"
	"github.com/elias-gill/poliplanner2/db"
	
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

	// 

	defer db.CloseDB()
}
