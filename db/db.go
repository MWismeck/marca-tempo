package db

import (
	"fmt"
	"log"
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
			log.Fatal(err)
		}
		db.AutoMigrate(&Employee{})
		return db
	}

	func NewEmployeeHandler(db *gorm.DB) *EmployeeHandler{
		return &EmployeeHandler{DB : db}
	}
	

	func(e *EmployeeHandler) AddEmplEmployee(employee Employee) error {
		if result := e.DB.Create(&employee); result.Error != nil{
			return result.Error
		}
		fmt.Println("Create employee !")
		return nil
	}

	func (e *EmployeeHandler) GetEmployee ()([]Employee, error){
		employees := []Employee{}

		err := e.DB.Find(&employees).Error
		return employees, err
	}

 
