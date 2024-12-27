package db

import (
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
)

type EmployeeHandler struct{
	DB *gorm.DB
}


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

  func Init() *gorm.DB{
	db, err := gorm.Open(sqlite.Open("employee.db"), &gorm.Config{})
		if err != nil{
			log.Fatal().Err(err).Msgf("Failed to initialize SQLite: %s", err.Error())
		}
		db.AutoMigrate(&Employee{})
		return db
	}

	func NewEmployeeHandler(db *gorm.DB) *EmployeeHandler{
		return &EmployeeHandler{DB : db}
	}
	

	func(e *EmployeeHandler) AddEmployee(employee Employee) error {
		if result := e.DB.Create(&employee); result.Error != nil{
			log.Error().Msg("Failed to create employee")
			return result.Error
		}
		log.Info().Msg("Create Employee!")
		return nil
	}

	func (e *EmployeeHandler) GetEmployees ()([]Employee, error){
		employees := []Employee{}

		err := e.DB.Find(&employees).Error
		return employees, err
	}

	func (e *EmployeeHandler) GetEmployee (id int)(Employee, error){
		var employee Employee
        err := e.DB.First(&employee, id)
		return employee, err.Error
	}
 
	func (e *EmployeeHandler) UpdateEmployee (updateEmployee Employee)error{
		
		return e.DB.Save(&updateEmployee).Error
	}

	func (e *EmployeeHandler) DeleteEmployee (employee Employee)error{
		
		return e.DB.Delete(&employee).Error
	}
