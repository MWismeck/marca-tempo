package api

import (
	"fmt"
	"net/http"
	"github.com/MWismeck/marca-tempo/db"
	"github.com/labstack/echo/v4"
)
type API struct{
	Echo *echo.Echo
	DB *db.EmployeeHandler
}

func NewServer()*API{
	e := echo.New()

	database := db.Init()
	employDB := db.NewEmployeeHandler(database)
	return &API{
		Echo: e,
		DB: employDB,
	}
}


func(api *API) ConfigureRoutes(){
	
	api.Echo.GET("/employee/", api.getEmployee)
	api.Echo.POST("/employee/", api.createEmployee)
	api.Echo.GET("/employee/:id", api.getEmployeeId)
	api.Echo.PUT("/employee/:id", api.updateEmployee)
	api.Echo.DELETE("/employee/:id", api.deleteEmployee)
}

func(api *API) Start()error{
	return api.Echo.Start(":8080")
}

func (api *API) getEmployee(c echo.Context) error{
	employees, err := api.DB.GetEmployee()
	if err != nil {
		return c.String(http.StatusNotFound,"Failed to get employees")
	}
	return c.JSON(http.StatusOK, employees)
}

func (api *API) createEmployee(c echo.Context)error{
	employee := db.Employee{}
	if err := c.Bind(&employee); err != nil{
		return err
	}
	if err := api.DB.AddEmplEmployee(employee); err != nil{
		return c.String(http.StatusInternalServerError,"Error to create employee")
	}
	return c.String(http.StatusOK, "Create employee")
}

func (api *API) getEmployeeId(c echo.Context) error{
	id := c.Param("id")
	getEmploy := fmt.Sprintf("Get %s employee", id)
	return c.String(http.StatusOK, getEmploy)
}

func (api *API) updateEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Update a employee")
}

func (api *API) deleteEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Delete a employee")
}
