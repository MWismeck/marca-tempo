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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Fail to update employee")
	}
    recivedEmployee := db.Employee{}
	if err := c.Bind(&recivedEmployee); err != nil{
		return err
	}
	updatingEmployee, err := api.DB.GetEmployee(id)
	if errors.Is(err, gorm.ErrRecordNotFound){
		return c.String(http.StatusNotFound,"Employee not found")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError,"Failed to get employee")
	}

	employee := updateEmployeeInfo (recivedEmployee, updatingEmployee)
	if err := api.DB.UpdateEmployee(employee); err != nil {
		return c.String(http.StatusInternalServerError,"Failed to save employee")
	}


	return c.JSON(http.StatusOK, employee)
}

func updateEmployeeInfo (recivedEmployee, employee db.Employee)db.Employee{
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
	if recivedEmployee.Active != employee.Active{
       employee.Active = recivedEmployee.Active
	}
	if recivedEmployee.Workload == 0 {
		employee.Workload = recivedEmployee.Workload
	}
	// não foi adicionado um metodo para o manager

	return employee
}
func (api *API) deleteEmployee(c echo.Context) error{
	return c.String(http.StatusOK,"Delete a employee")
}


