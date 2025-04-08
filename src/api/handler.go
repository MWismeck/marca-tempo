package api

import (
	"errors"
	_"github.com/MWismeck/marca-tempo/src/docs"
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"golang.org/x/crypto/bcrypt"
	
)

// getEmployees godoc
//
//	@Summary        Get a list of employees
//	@Desciption     Retrive employees details
//	@Tags           employees
//	@Accept         json
//	@Produce        json
//	@Param          register path int false  "Registration"
//	@Sucess         200 {object} schemas.EmployeeResponse
//	@Failure        404
//	@Router         /Employees/ [get]
func (api *API) getEmployees(c echo.Context) error {
	employees, err := api.DB.GetEmployees()
	if err != nil {
		return c.String(http.StatusNotFound, "Failed to get employees")
	}
	active := c.QueryParam("active")

	if active != "" {
		act, err := strconv.ParseBool(active)
		if err != nil {
			log.Error().Err(err).Msgf("[api] error to parsing boolean")
			return c.String(http.StatusInternalServerError, "Failed to parse boolean")
		}
		employees, err = api.DB.GetFilteredEmployee(act)
	}

	listOfEmployees := map[string][]schemas.EmployeeResponse{"employees:": schemas.NewResponse(employees)}

	return c.JSON(http.StatusOK, listOfEmployees)
}

// createEmployee godoc
//
//	@Summary        Create employee
//	@Desciption     Create employee
//	@Tags           employees
//	@Accept         json
//	@Produce        json
//	@Sucess         200 {object} schemas.EmployeeResponse
//	@Failure        400
//	@Router         /Employees/ [post]
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

// getEmployeeId godoc
//
//	@Summary        Get a list of employees
//	@Desciption     Retrive employee details
//	@Tags           employees
//	@Accept         json
//	@Produce        json
//	@Sucess         200 {object} schemas.EmployeeResponse
//	@Failure        404
//	@Router         /Employee/{id} [get]
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

// updateEmployees godoc
//
//	@Summary        Update a employee
//	@Desciption     Update a employee details
//	@Tags           employees
//	@Accept         json
//	@Produce        json
//	@Sucess         200 {object} schemas.EmployeeResponse
//	@Failure        404
//	@Failure        500
//	@Router         /Employee/{id} [put]
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

// deleteEmployees godoc
//
//	@Summary        Delete a employee
//	@Desciption     Delete a employee details
//	@Tags           employees
//	@Accept         json
//	@Produce        json
//	@Sucess         200 {object} schemas.EmployeeResponse
//	@Failure        404
//	@Failure        500
//	@Router         /Employee/{id} [delete]
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



func (api *API) login(c echo.Context) error {
    loginReq := struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }{}

    if err := c.Bind(&loginReq); err != nil {
        return c.String(http.StatusBadRequest, "Invalid request")
    }

    var login schemas.Login
    if err := api.DB.DB.Where("email = ?", loginReq.Email).First(&login).Error; err != nil {
        return c.String(http.StatusUnauthorized, "Invalid email or password")
    }

    if !CheckPasswordHash(loginReq.Password, login.Password) {
        return c.String(http.StatusUnauthorized, "Invalid email or password")
    }

    // Get employee details
    var employee schemas.Employee
    if err := api.DB.DB.Where("email = ?", loginReq.Email).First(&employee).Error; err != nil {
        return c.String(http.StatusInternalServerError, "Error retrieving employee details")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "message":        "Login successful",
        "employee_id":    employee.ID,
        "employee_email": employee.Email,
        "employee_name":  employee.Name,
    })
}


type PasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}


func (api *API) createOrUpdatePassword(c echo.Context) error {
	var req PasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	
	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", req.Email).First(&employee).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Employee not found"})
	}

	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	
	var login schemas.Login
	if err := api.DB.DB.Where("email = ?", req.Email).First(&login).Error; err == nil {
		
		login.Password = string(hashedPassword)
		if err := api.DB.DB.Save(&login).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
		}
	} else {
		
		newLogin := schemas.Login{
			Email:    req.Email,
			Password: string(hashedPassword),
		}
		if err := api.DB.DB.Create(&newLogin).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create login"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}
