package main

import (
	"net/http"
	
	"github.com/labstack/echo/v4"
)

func getEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"List of all employees")
}
func createEmployee(c echo.Context)error{
	return c.String(http.StatusOK, "Create student")
}
func getEmployeeId(c echo.Context) error{
	id := c.Param("id")
	return c.String(http.StatusOK,"Get a employee")
}
func updateEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Update a employee")
}
func deleteEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Delete a employee")
}

func main() {
	e := echo.New()
	// START SERVER
	e.Logger.Fatal(e.Start(":8080"))

	//ROUTES
	e.GET("/employee", getEmployee)
	e.POST("/employee", createEmployee)
	e.GET("/employee/:id", getEmployeeId)
	e.PUT("/employee/:id", updateEmployee)
	e.DELETE("/employee/:id", deleteEmployee)

	
}