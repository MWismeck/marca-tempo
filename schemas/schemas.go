package schemas

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	gorm.Model

	Name     string  `json:"name"`
	CPF      string  `json:"cpf"`
	RG       string  `json:"rg"`
	Email    string  `json:"email"`
	Age      int     `json:"age"`
	Active   bool    `json:"active"`
	Workload float32 `json:"workload"`
	//IsManager bool `json:"ismanager"`
}
type EmployeeResponse struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
	Name      string    `json:"name"`
	CPF       string    `json:"cpf"`
	RG        string    `json:"rg"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Active    bool      `json:"active"`
	Workload  float32   `json:"workload"`
}

func NewResponse(employees []Employee) []EmployeeResponse {
	employeesResponse := []EmployeeResponse{}

	for _, employee := range employees {

		employeeResponse := EmployeeResponse{
			ID:        int(employee.ID),
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
			DeletedAt: employee.DeletedAt.Time,
			Name:      employee.Name,
			Email:     employee.Email,
			CPF:       employee.CPF,
			RG:        employee.RG,
			Age:       employee.Age,
			Active:    employee.Active,
		}
		employeesResponse = append(employeesResponse, employeeResponse)

	}
	return employeesResponse

}
