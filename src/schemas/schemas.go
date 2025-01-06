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

	Login Login `gorm:"foreignKey:Email;constraint:OnDelete:CASCADE"` // Relacionamento 1:1 com Login
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
type TimeLog struct {
	gorm.Model
	EmployeeID          uint      `json:"employee_id"`
	EntryTime           time.Time `json:"entry_time"`      // Entrada
	LunchExitTime       time.Time `json:"lunch_exit_time"` // Saída para o almoço
	LunchReturnTime     time.Time `json:"lunch_return_time"` // Retorno do almoço
	ExitTime            time.Time `json:"exit_time"`       // Saída do expediente
	ExtraHours          float32   `json:"extra_hours"`     // Horas extras
	MissingHours        float32   `json:"missing_hours"`   // Horas faltantes
	Balance             float32   `json:"balance"`         // Saldo de horas
	Workload            float32   `json:"workload"`        // Carga horária

}
type Login struct {
    gorm.Model
    Email    string `json:"email" gorm:"unique"` 
    Password string `json:"password"`
}


