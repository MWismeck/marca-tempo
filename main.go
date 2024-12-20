package main

import (
	"net/http"
	
	"github.com/labstack/echo/v4"
)

func getEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"List of all employees")
}

func main() {
	e := echo.New()

	//ROUTES
	e.GET("/employee", getEmployee)
	// START SERVER
	e.Logger.Fatal(e.Start(":8080"))

	
}