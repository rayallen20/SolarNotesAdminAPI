package main

import (
	"SolarNotesAdminAPI/conf"
	"fmt"
	"log"
)

func main() {
	config := &conf.Config{}
	err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", config)
}
