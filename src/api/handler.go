package api

import (
	"errors"
	"net/http"
	"strconv"

	_ "github.com/MWismeck/marca-tempo/src/docs"
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// getEmployees godoc
//
//	@Summary		Listar funcionários
//	@Description	Retorna lista de funcionários com filtros opcionais
//	@Tags			employees
//	@Accept			json
//	@Produce		json
//	@Param			manager_email	query	string	false	"Email do gerente para filtrar funcionários da empresa"
//	@Param			active			query	boolean	false	"Filtrar por funcionários ativos/inativos"
//	@Success		200	{object}	map[string][]schemas.EmployeeResponse
//	@Failure		401	{string}	string	"Gerente não encontrado"
//	@Failure		404	{string}	string	"Funcionários não encontrados"
//	@Failure		500	{string}	string	"Erro interno do servidor"
//	@Router			/employees/ [get]
func (api *API) getEmployees(c echo.Context) error {
	managerEmail := c.QueryParam("manager_email")
	active := c.QueryParam("active")

	var employees []schemas.Employee

	if managerEmail != "" {
		var manager schemas.Employee
		if err := api.DB.DB.Where("email = ? AND is_manager = ?", managerEmail, true).First(&manager).Error; err != nil {
			log.Error().Err(err).Msgf("[api] Gerente não encontrado: %s", managerEmail)
			return c.String(http.StatusUnauthorized, "Gerente não encontrado")
		}

		query := api.DB.DB.Where("company_cnpj = ?", manager.CompanyCNPJ)

		if active != "" {
			if act, err := strconv.ParseBool(active); err == nil {
				query = query.Where("active = ?", act)
			}
		}

		if err := query.Find(&employees).Error; err != nil {
			log.Error().Err(err).Msg("[api] Erro ao buscar funcionários da empresa")
			return c.String(http.StatusInternalServerError, "Erro ao buscar funcionários")
		}
	} else {
		var err error
		employees, err = api.DB.GetEmployees()
		if err != nil {
			return c.String(http.StatusNotFound, "Failed to get employees")
		}

		if active != "" {
			act, err := strconv.ParseBool(active)
			if err != nil {
				log.Error().Err(err).Msgf("[api] error to parsing boolean")
				return c.String(http.StatusInternalServerError, "Failed to parse boolean")
			}
			employees, err = api.DB.GetFilteredEmployee(act)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to filter employees")
			}
		}
	}

	listOfEmployees := map[string][]schemas.EmployeeResponse{"employees:": schemas.NewResponse(employees)}

	return c.JSON(http.StatusOK, listOfEmployees)
}

// createEmployee godoc
//
//	@Summary		Criar funcionário
//	@Description	Cria um novo funcionário no sistema
//	@Tags			employees
//	@Accept			json
//	@Produce		json
//	@Param			body	body		EmployeeRequest	true	"Dados do funcionário"
//	@Success		200		{object}	schemas.Employee
//	@Failure		400		{string}	string	"Dados inválidos ou empresa não encontrada"
//	@Failure		500		{string}	string	"Erro interno do servidor"
//	@Router			/employee/ [post]
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
		Workload:    employeeReq.Workload,
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
//	@Summary		Buscar funcionário por ID
//	@Description	Retorna os dados de um funcionário específico
//	@Tags			employees
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID do funcionário"
//	@Success		200	{object}	schemas.Employee
//	@Failure		404	{string}	string	"Funcionário não encontrado"
//	@Failure		500	{string}	string	"Erro interno do servidor"
//	@Router			/employee/{id} [get]
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

// updateEmployee godoc
//
//	@Summary		Atualizar funcionário
//	@Description	Atualiza os dados de um funcionário
//	@Tags			employees
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"ID do funcionário"
//	@Param			body	body		schemas.Employee	true	"Dados atualizados do funcionário"
//	@Success		200		{object}	schemas.Employee
//	@Failure		404		{string}	string	"Funcionário não encontrado"
//	@Failure		500		{string}	string	"Erro interno do servidor"
//	@Router			/employee/{id} [put]
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
	if recivedEmployee.Workload != 0 {
		employee.Workload = recivedEmployee.Workload
	}

	return employee
}

// deleteEmployee godoc
//
//	@Summary		Excluir funcionário
//	@Description	Remove um funcionário do sistema
//	@Tags			employees
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID do funcionário"
//	@Success		200	{object}	schemas.Employee
//	@Failure		404	{string}	string	"Funcionário não encontrado"
//	@Failure		500	{string}	string	"Erro interno do servidor"
//	@Router			/employee/{id} [delete]
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

// login godoc
//
//	@Summary		Login do usuário
//	@Description	Autentica um usuário no sistema
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"Credenciais de login"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{string}	string	"Dados inválidos"
//	@Failure		401		{string}	string	"Email ou senha inválidos"
//	@Failure		500		{string}	string	"Erro interno do servidor"
//	@Router			/login [post]
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

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Message       string `json:"message"`
	EmployeeID    uint   `json:"employee_id"`
	EmployeeEmail string `json:"employee_email"`
	EmployeeName  string `json:"employee_name"`
	Role          string `json:"role"`
}

type PasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// createOrUpdatePassword godoc
//
//	@Summary		Criar ou atualizar senha
//	@Description	Cria ou atualiza a senha de um funcionário
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		PasswordRequest	true	"Dados para criação/atualização de senha"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/login/password [post]
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

// createCompany godoc
//
//	@Summary		Criar empresa
//	@Description	Cria uma nova empresa no sistema
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CompanyRequest	true	"Dados da empresa"
//	@Success		201		{object}	schemas.Company
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/admin/create_company [post]
func (api *API) createCompany(c echo.Context) error {
	var req CompanyRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

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

// listCompanies godoc
//
//	@Summary		Listar empresas
//	@Description	Retorna lista de todas as empresas cadastradas
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		schemas.Company
//	@Failure		500	{object}	map[string]string
//	@Router			/admin/companies [get]
func (api *API) listCompanies(c echo.Context) error {
	var companies []schemas.Company
	if err := api.DB.DB.Preload("Employees").Find(&companies).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao listar empresas"})
	}
	return c.JSON(http.StatusOK, companies)
}

// createManager godoc
//
//	@Summary		Criar gerente
//	@Description	Cria um novo gerente no sistema
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			body	body		EmployeeRequest	true	"Dados do gerente"
//	@Success		201		{object}	schemas.Employee
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/admin/create_manager [post]
func (api *API) createManager(c echo.Context) error {
	var req EmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Erro de validação: " + err.Error()})
	}

	var company schemas.Company
	if err := api.DB.DB.Where("cnpj = ?", req.CompanyCNPJ).First(&company).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Empresa com este CNPJ não existe"})
	}

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

// listManagers godoc
//
//	@Summary		Listar gerentes
//	@Description	Retorna lista de todos os gerentes cadastrados
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		schemas.Employee
//	@Failure		500	{object}	map[string]string
//	@Router			/admin/managers [get]
func (api *API) listManagers(c echo.Context) error {
	var managers []schemas.Employee
	if err := api.DB.DB.Where("is_manager = ?", true).Find(&managers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar gerentes"})
	}
	return c.JSON(http.StatusOK, managers)
}
