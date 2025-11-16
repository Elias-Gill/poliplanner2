package main

import (
	"github.com/elias-gill/poliplanner2/config"
	"github.com/elias-gill/poliplanner2/db"
)

func main() {
	config := config.Load()
	println("initializing db")
	err := db.InitDB(config)
	if err != nil {
		panic(err.Error())
	}
	db.CloseDB()
}
