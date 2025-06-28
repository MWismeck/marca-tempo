package api

import (
	"fmt"
	"regexp"
)

type EmployeeRequest struct {
	Name        string  `json:"name"`
	CPF         string  `json:"cpf"`
	RG          string  `json:"rg"`
	Email       string  `json:"email"`
	Age         int     `json:"age"`
	Active      *bool   `json:"active"` // using bool as a pointer to force true/false
	Workload    float32 `json:"workload"`
	Password    string  `json:"password"`
	CompanyCNPJ string  `json:"company_cnpj"`
}

func errParamRequired(param, typ string) error {
	return fmt.Errorf("param '%s' of type '%s' is required", param, typ)
}

// Regex para validação de CPF, RG e caractere especial na senha
var (
	cpfRegex         = regexp.MustCompile(`^\d{11}$`)
	rgRegex          = regexp.MustCompile(`^\d{9}$`)
	specialCharRegex = regexp.MustCompile(`[!@#\$%\^&\*(),.?":{}|<>]`)
)

func (e *EmployeeRequest) Validate() error {
	if e.Name == "" {
		return errParamRequired("name", "string")
	}
	if !cpfRegex.MatchString(e.CPF) {
		return fmt.Errorf("CPF deve conter exatamente 11 dígitos numéricos")
	}
	if !rgRegex.MatchString(e.RG) {
		return fmt.Errorf("RG deve conter exatamente 9 dígitos numéricos")
	}
	if e.Email == "" {
		return errParamRequired("email", "string")
	}
	if e.Age == 0 {
		return errParamRequired("age", "int")
	}
	if e.Active == nil {
		return errParamRequired("active", "bool")
	}
	if e.Password == "" {
		return errParamRequired("password", "string")
	}

	// Regras de segurança para senha (exceto admin, controlado no handler)
	if len(e.Password) < 6 {
		return fmt.Errorf("Senha deve conter no mínimo 6 caracteres")
	}
	if !specialCharRegex.MatchString(e.Password) {
		return fmt.Errorf("Senha deve conter pelo menos um caractere especial")
	}

	return nil
}
