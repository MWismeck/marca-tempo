package api

import (
	"errors"
	_ "github.com/MWismeck/marca-tempo/src/docs"
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
		return c.String(http.StatusBadRequest, "Error validating employee")
	}

	
	var company schemas.Company
	if err := api.DB.DB.Where("cnpj = ?", employeeReq.CompanyCNPJ).First(&company).Error; err != nil {
		log.Warn().Str("cnpj", employeeReq.CompanyCNPJ).Msg("[api] CNPJ não encontrado")
		return c.String(http.StatusBadRequest, "Empresa com este CNPJ não encontrada")
	}

	employee := schemas.Employee{
		Name:        employeeReq.Name,
		Email:       employeeReq.Email,
		CPF:         employeeReq.CPF,
		RG:          employeeReq.RG,
		Age:         employeeReq.Age,
		Active:      *employeeReq.Active,
		CompanyCNPJ: employeeReq.CompanyCNPJ,
	}
	hashedPassword, err := HashPassword(employeeReq.Password)
if err != nil {
    return c.String(http.StatusInternalServerError, "Erro ao processar senha")
}
login := schemas.Login{
    Email:    employee.Email,
    Password: hashedPassword,
}
if err := api.DB.DB.Create(&login).Error; err != nil {
    return c.String(http.StatusInternalServerError, "Erro ao salvar login")
}


	if err := api.DB.AddEmployee(employee); err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao criar funcionário")
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
		"role": func() string {
			switch {
			case employee.IsAdmin:
				return "admin"
			case employee.IsManager:
				return "manager"
			default:
				return "employee"
			}
		}(),
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
type CompanyRequest struct {
	Name   string `json:"name" validate:"required"`
	CNPJ   string `json:"cnpj" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
	Fone   string `json:"fone" validate:"required"`
	Active bool   `json:"active"`
}


func (api *API) createCompany(c echo.Context) error {
	var req CompanyRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// opcional: usar validator lib
	if req.CNPJ == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Nome e CNPJ são obrigatórios"})
	}

	var existing schemas.Company
	if err := api.DB.DB.Where("cnpj = ?", req.CNPJ).First(&existing).Error; err == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Empresa já cadastrada"})
	}

	company := schemas.Company{
		Name:   req.Name,
		CNPJ:   req.CNPJ,
		Email:  req.Email,
		Fone:   req.Fone,
		Active: req.Active,
	}

	if err := api.DB.DB.Create(&company).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar empresa"})
	}

	return c.JSON(http.StatusCreated, company)
}



func (api *API) listCompanies(c echo.Context) error {
	var companies []schemas.Company
	if err := api.DB.DB.Preload("Employees").Find(&companies).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao listar empresas"})
	}
	return c.JSON(http.StatusOK, companies)
}

func (api *API) createManager(c echo.Context) error {
	var req EmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Erro de validação: " + err.Error()})
	}

	// Verifica se a empresa com o CNPJ existe
	var company schemas.Company
	if err := api.DB.DB.Where("cnpj = ?", req.CompanyCNPJ).First(&company).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Empresa com este CNPJ não existe"})
	}

	// Gera hash da senha
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao gerar hash da senha"})
	}

	manager := schemas.Employee{
		Name:        req.Name,
		Email:       req.Email,
		CPF:         req.CPF,
		RG:          req.RG,
		Age:         req.Age,
		Active:      *req.Active,
		Workload:    req.Workload,
		IsManager:   true,
		IsAdmin:     false,
		CompanyCNPJ: req.CompanyCNPJ,
	}

	tx := api.DB.DB.Begin()
	if err := tx.Create(&manager).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao cadastrar gerente"})
	}

	login := schemas.Login{
		Email:    req.Email,
		Password: hashedPassword,
	}
	if err := tx.Create(&login).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar login do gerente"})
	}

	tx.Commit()
	return c.JSON(http.StatusCreated, manager)
}


func (api *API) listManagers(c echo.Context) error {
	var managers []schemas.Employee
	if err := api.DB.DB.Where("is_manager = ?", true).Find(&managers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar gerentes"})
	}
	return c.JSON(http.StatusOK, managers)
}
