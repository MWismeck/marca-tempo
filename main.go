package main

import (
	"github.com/MWismeck/marca-tempo/src/api"
	"github.com/MWismeck/marca-tempo/src/db"
	"log"
	"time"
)



func main() {
	// Initialize the database
	database := db.Init()

	// Create and configure the server
	server := api.NewServer(database)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait a moment for the server to start
	time.Sleep(1 * time.Second)

	
	// Keep the main goroutine running
	select {}
}
