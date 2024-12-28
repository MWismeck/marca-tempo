package api

import (
	"errors"
	"net/http"
	"strconv"
	"github.com/MWismeck/marca-tempo/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (api *API) getEmployees(c echo.Context) error {
	employees, err := api.DB.GetEmployees()
	if err != nil {
		return c.String(http.StatusNotFound, "Failed to get employees")
	}
	listOfEmployees := map[string][]schemas.EmployeeResponse{"employees:": schemas.NewResponse(employees)}

	return c.JSON(http.StatusOK, listOfEmployees)
}

func (api *API) createEmployee(c echo.Context) error {
	employeeReq := EmployeeRequest{}
	if err := c.Bind(&employeeReq); err != nil {
		return err
	}
	if err := employeeReq.Validate(); err != nil {
		log.Error().Err(err).Msgf("[api] error validating struct")
		return c.String(http.StatusBadRequest, "Error to validating employee")
	}

	employee := schemas.Employee{
		Name:   employeeReq.Name,
		Email:  employeeReq.Email,
		CPF:    employeeReq.CPF,
		RG:     employeeReq.RG,
		Age:    employeeReq.Age,
		Active: *employeeReq.Active,
	}

	if err := api.DB.AddEmployee(employee); err != nil {
		return c.String(http.StatusInternalServerError, "Error to create employee")
	}
	return c.JSON(http.StatusOK, employee)
}

func (api *API) getEmployeeId(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get employee ID")
	}
	employee, err := api.DB.GetEmployee(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.String(http.StatusNotFound, "Employee not found")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get employee")
	}
	return c.JSON(http.StatusOK, employee)
}

func (api *API) updateEmployee(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Fail to update employee")
	}
	recivedEmployee := schemas.Employee{}
	if err := c.Bind(&recivedEmployee); err != nil {
		return err
	}
	updatingEmployee, err := api.DB.GetEmployee(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.String(http.StatusNotFound, "Employee not found")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get employee")
	}

	employee := updateEmployeeInfo(recivedEmployee, updatingEmployee)
	if err := api.DB.UpdateEmployee(employee); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save employee")
	}

	return c.JSON(http.StatusOK, employee)
}

func updateEmployeeInfo(recivedEmployee, employee schemas.Employee) schemas.Employee {
	if recivedEmployee.Name != "" {
		employee.Name = recivedEmployee.Name
	}
	if recivedEmployee.CPF != "" {
		employee.CPF = recivedEmployee.CPF
	}
	if recivedEmployee.RG != "" {
		employee.RG = recivedEmployee.RG
	}
	if recivedEmployee.Email != "" {
		employee.Email = recivedEmployee.Email
	}
	if recivedEmployee.Age > 0 {
		employee.Age = recivedEmployee.Age
	}
	if recivedEmployee.Active != employee.Active {
		employee.Active = recivedEmployee.Active
	}
	if recivedEmployee.Workload == 0 {
		employee.Workload = recivedEmployee.Workload
	}
	// não foi adicionado um metodo para o manager

	return employee
}
func (api *API) deleteEmployee(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get employee ID")
	}
	employee, err := api.DB.GetEmployee(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.String(http.StatusNotFound, "Employee not found")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get employee")
	}
	if err := api.DB.DeleteEmployee(employee); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete employee")
	}
	return c.JSON(http.StatusOK, employee)
}
