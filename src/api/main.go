package main

import (
	"github.com/MWismeck/marca-tempo/api"
	"github.com/MWismeck/marca-tempo/db"
	"log"
)

func main() {
	// Inicializa o banco de dados
	database := db.Init()

	// Inicializa e inicia o servidor
	server := api.NewServer(database)

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
