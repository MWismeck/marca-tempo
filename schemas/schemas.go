package schemas
import(
"gorm.io/gorm"

)

type Employee struct{
	gorm.Model
	
	Name string `json:"name"`
	CPF string  `json:"cpf"`
	RG string `json:"rg"`
	Email string `json:"email"`
	Age int `json:"age"`
	Active bool `json:"active"`
	Workload float32 `json:"workload"`
	IsManager bool `json:"ismanager"`
} 