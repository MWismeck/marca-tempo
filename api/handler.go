package api

import (
	"net/http"
	"strconv"
	"github.com/MWismeck/marca-tempo/db"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"errors"
)

func (api *API) getEmployees(c echo.Context) error{
	employees, err := api.DB.GetEmployees()
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

func (api *API) getEmployeeId(c echo.Context) error {
	id,err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusInternalServerError,"Failed to get employee ID")
	}
	employee, err := api.DB.GetEmployee(id)
	if errors.Is(err, gorm.ErrRecordNotFound){
		return c.String(http.StatusNotFound,"Employee not found")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError,"Failed to get employee")
	}
	return c.JSON(http.StatusOK, employee)
}

func (api *API) updateEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Update a employee")
}

func (api *API) deleteEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Delete a employee")
}