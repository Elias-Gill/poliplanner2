package main

import (
	"github.com/elias-gill/poliplanner2/config"
	"github.com/elias-gill/poliplanner2/db"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		panic(err.Error())
	}

	println("initializing db")
	err := db.InitDB(cfg)
	if err != nil {
		panic(err.Error())
	}
	db.CloseDB()
}
