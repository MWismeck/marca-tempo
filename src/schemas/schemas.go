package schemas

import (
	"gorm.io/gorm"
	"time"
)

type Employee struct {
	gorm.Model
	Name      string  `json:"name"`
	CPF       string  `json:"cpf"`
	RG        string  `json:"rg"`
	Email     string  `json:"email" gorm:"type:varchar(255);unique"`
	Age       int     `json:"age"`
	Active    bool    `json:"active"`
	Workload  float32 `json:"workload"`
	IsManager bool    `json:"ismanager"`
	IsAdmin   bool    `json:"is_admin" gorm:"default:false"` // Novo campo para admin

	// Referência à empresa pelo CNPJ
	CompanyCNPJ string  `json:"company_cnpj" gorm:"type:varchar(20);not null"` // FK
	Company     Company `gorm:"foreignKey:CompanyCNPJ;references:CNPJ;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Login    Login     `gorm:"foreignKey:Email;references:Email;constraint:OnDelete:CASCADE"`
	TimeLogs []TimeLog `gorm:"foreignKey:EmployeeEmail;references:Email"`
}

type PontoSolicitacao struct {
	gorm.Model
	FuncionarioEmail string    `json:"funcionario_email" gorm:"type:varchar(255);not null"`
	DataSolicitada   time.Time `json:"data_solicitada"`
	Motivo           string    `json:"motivo" gorm:"type:text"`
	Status           string    `json:"status" gorm:"default:'pendente'"` // pendente, aprovado, rejeitado
}

type Company struct {
	gorm.Model
	Name      string     `json:"name"`
	CNPJ      string     `json:"cnpj" gorm:"unique;not null"` // CNPJ como chave única
	Email     string     `json:"email"`
	Fone      string     `json:"fone"`
	Active    bool       `json:"active"`
	Employees []Employee `gorm:"foreignKey:CompanyCNPJ;references:CNPJ"` // One-to-many via CNPJ
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
	EmployeeEmail     string    `json:"employee_email" gorm:"type:varchar(255);not null"`
	LogDate           time.Time `json:"log_date" gorm:"not null"`
	EntryTime         time.Time `json:"entry_time,omitempty"`
	LunchExitTime     time.Time `json:"lunch_exit_time,omitempty"`
	LunchReturnTime   time.Time `json:"lunch_return_time,omitempty"`
	ExitTime          time.Time `json:"exit_time,omitempty"`
	ExtraHours        float32   `json:"extra_hours" gorm:"default:0"`
	MissingHours      float32   `json:"missing_hours" gorm:"default:0"`
	Balance           float32   `json:"balance" gorm:"default:0"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	EditadoPorGerente string    `json:"editado_por_gerente" gorm:"type:varchar(255)"`
	EditadoEm         time.Time `json:"editado_em"`
	MotivoEdicao      string    `json:"motivo_edicao" gorm:"type:text"`
}

type Login struct {
	gorm.Model
	Email    string `json:"email" gorm:"type:varchar(255);unique;not null"`
	Password string `json:"password" gorm:"not null"`
}
