package main

import (
	"SolarNotesAdminAPI/conf"
	"log"
)

func main() {
	config := &conf.Config{}
	err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
}
