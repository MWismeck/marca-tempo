package api

import (
	"fmt"
	"net/http"
	"github.com/MWismeck/marca-tempo/db"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)
type API struct{
	Echo *echo.Echo
	DB *gorm.DB
}

func NewServer()*API{
	e := echo.New()

	db := db.Init()
	return &API{
		Echo: e,
		DB: db,
	}
}


func(api *API) ConfigureRoutes(){
	
	api.Echo.GET("/employee/", getEmployee)
	api.Echo.POST("/employee/", createEmployee)
	api.Echo.GET("/employee/:id", getEmployeeId)
	api.Echo.PUT("/employee/:id", updateEmployee)
	api.Echo.DELETE("/employee/:id", deleteEmployee)
}

func(api *API) Start()error{
	return api.Echo.Start(":8080")
}

func getEmployee(c echo.Context) error{
	employees, err :=db.GetEmployee()
	if err != nil {
		return c.String(http.StatusNotFound,"Failed to get employees")
	}
	return c.JSON(http.StatusOK, employees)
}

func createEmployee(c echo.Context)error{
	employee := db.Employee{}
	if err := c.Bind(&employee); err != nil{
		return err
	}
	if err := db.AddEmplEmployee(employee); err != nil{
		return c.String(http.StatusInternalServerError,"Error to create employee")
	}
	return c.String(http.StatusOK, "Create employee")
}

func getEmployeeId(c echo.Context) error{
	id := c.Param("id")
	getEmploy := fmt.Sprintf("Get %s employee", id)
	return c.String(http.StatusOK, getEmploy)
}

func updateEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Update a employee")
}

func deleteEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Delete a employee")
}
