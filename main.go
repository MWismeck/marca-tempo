package main

import (
	"github.com/MWismeck/marca-tempo/src/api"
	"github.com/MWismeck/marca-tempo/src/db"
	"log"
	"os/exec"
	"runtime"
	"time"
)

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		log.Println("Unsupported platform. Please open the browser manually at:", url)
		return
	}

	if err != nil {
		log.Println("Error opening browser:", err)
	}
}

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

	// Open the browser to the application
	openBrowser("http://localhost:8080")

	// Keep the main goroutine running
	select {}
}
