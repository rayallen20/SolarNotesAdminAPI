package main

import (
	"SolarNotesAdminAPI/conf"
	"SolarNotesAdminAPI/db"
	"log"
)

func main() {
	config := &conf.Config{}
	err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Connect(config.Database)
	if err != nil {
		log.Fatal(err)
	}
}
