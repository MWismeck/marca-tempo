package api

import "fmt"

type EmployeeRequest struct {
	Name     string  `json:"name"`
	CPF      string  `json:"cpf"`
	RG       string  `json:"rg"`
	Email    string  `json:"email"`
	Age      int     `json:"age"`
	Active   *bool   `json:"active"` // using bool as a pointer to force true/false
	Workload float32 `json:"workload"`
	//IsManager *bool    `json:"ismanager"`
}

func errParamRequired(param, typ string) error {
	return fmt.Errorf("param '%s' of type '%s' is required", param, typ)
}

func (e *EmployeeRequest) Validate() error {
	if e.Name == "" {
		return errParamRequired("name", "string")
	}
	if e.CPF == "" {
		return errParamRequired("cpf", "string")
	}
	if e.RG == "" {
		return errParamRequired("rg", "string")
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
	return nil
}
