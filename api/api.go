package api

import (
	"github.com/MWismeck/marca-tempo/db"
	"github.com/labstack/echo/v4"
	"github.com/swaggo/echo-swagger"
	_ "github.com/MWismeck/marca-tempo/docs"
)

type API struct {
	Echo *echo.Echo
	DB   *db.EmployeeHandler
}

// @title Marca Tempo
// @version 1.0
// @description This is a sample server Marca Tempo API
// @host localhost:8080
// @BasePath / 
// @schemes http
func NewServer() *API {
	e := echo.New()

	database := db.Init()
	employDB := db.NewEmployeeHandler(database)
	return &API{
		Echo: e,
		DB:   employDB,
	}
}

func (api *API) ConfigureRoutes() {

	api.Echo.GET("/employees/", api.getEmployees)
	api.Echo.POST("/employee/", api.createEmployee)
	api.Echo.GET("/employee/:id", api.getEmployeeId)
	api.Echo.PUT("/employee/:id", api.updateEmployee)
	api.Echo.DELETE("/employee/:id", api.deleteEmployee)
	api.Echo.GET("/swagger/*", echoSwagger.EchoWrapHandler())
}

func (api *API) Start() error {
	return api.Echo.Start(":8080")
}
