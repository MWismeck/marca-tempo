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

	Login Login `gorm:"foreignKey:Email;constraint:OnDelete:CASCADE"`
	TimeLogs []TimeLog `gorm:"foreignKey:EmployeeEmail;references:Email;"` 
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
	ID              int       `json:"id" gorm:"primaryKey"`
	EmployeeEmail   string    `json:"employee_email" gorm:"not null"` 
	LogDate         time.Time `json:"log_date" gorm:"not null"`
	EntryTime       time.Time `json:"entry_time,omitempty"`
	LunchExitTime   time.Time `json:"lunch_exit_time,omitempty"`
	LunchReturnTime time.Time `json:"lunch_return_time,omitempty"`
	ExitTime        time.Time `json:"exit_time,omitempty"`
	ExtraHours      float32   `json:"extra_hours" gorm:"default:0"`
	MissingHours    float32   `json:"missing_hours" gorm:"default:0"`
	Balance         float32   `json:"balance" gorm:"default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type Login struct {
    gorm.Model
    Email    string `json:"email" gorm:"unique"` 
    Password string `json:"password"`
}


