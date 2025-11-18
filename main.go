package main

import (
	"github.com/elias-gill/poliplanner2/config"
	"github.com/elias-gill/poliplanner2/db"
	
	log "github.com/elias-gill/poliplanner2/logger"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		panic(err.Error())
	}

	log.Logger.Debug("initializing db")
	err := db.InitDB(cfg)
	if err != nil {
		panic(err.Error())
	}
	defer db.CloseDB()
}
