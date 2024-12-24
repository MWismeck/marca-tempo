package db

import (
	"fmt"
	"log"
	"gorm.io/driver/sqlite"
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

  func Init() *gorm.DB{
	db, err := gorm.Open(sqlite.Open("employee.db"), &gorm.Config{})
		if err != nil{
			log.Fatal(err)
		}
		db.AutoMigrate(&Employee{})
		return db
	}
	

	func AddEmplEmployee(employee Employee) error {
		db := Init()
		
		if result := db.Create(&employee); result.Error != nil{
			return result.Error
		}
		fmt.Println("Create employee !")
		return nil
	}

	func GetEmployee ()([]Employee, error){
		employees := []Employee{}

		db := Init()
		err := db.Find(&employees).Error
		return employees, err
	}

 
