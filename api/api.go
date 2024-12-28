package api

import (
	"github.com/MWismeck/marca-tempo/db"
	"github.com/labstack/echo/v4"
)

type API struct {
	Echo *echo.Echo
	DB   *db.EmployeeHandler
}

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
}

func (api *API) Start() error {
	return api.Echo.Start(":8080")
}
