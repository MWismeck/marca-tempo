package main

import (
	"github.com/MWismeck/marca-tempo/src/api"
	"github.com/MWismeck/marca-tempo/src/db"
	"log"
)

func main() {
	
	database := db.Init()

	
	server := api.NewServer(database)

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
