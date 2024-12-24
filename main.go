package main

import (
	"fmt"
	"net/http"
	"github.com/MWismeck/marca-tempo/db"
	"github.com/labstack/echo/v4"
)

func getEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"List of all employees")
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

func main() {
	e := echo.New()
	

	//ROUTES
	e.GET("/employee/", getEmployee)
	e.POST("/employee/", createEmployee)
	e.GET("/employee/:id", getEmployeeId)
	e.PUT("/employee/:id", updateEmployee)
	e.DELETE("/employee/:id", deleteEmployee)

	// START SERVER
	e.Logger.Fatal(e.Start(":8080"))
}