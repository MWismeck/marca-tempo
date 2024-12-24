package main

import (
	"log"

	"github.com/MWismeck/marca-tempo/api"
)

func main() {

	server := api.NewServer()

	server.ConfigureRoutes()

	if err := server.Start(); err != nil{
		log.Fatal(err)
	}
}