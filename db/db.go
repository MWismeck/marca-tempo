package db

import (
	"fmt"
	"log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Employee struct{
	gorm.Model
	
	Name string
	CPF int
	RG int
	Email string
	Age int
	Active bool
	Workload float32
	IsManager bool
} 

  func Init() *gorm.DB{
	db, err := gorm.Open(sqlite.Open("employee.db"), &gorm.Config{})
		if err != nil{
			log.Fatal(err)
		}
		db.AutoMigrate(&Employee{})
		return db
	}
	

	func AddEmplEmployee()  {
		db := Init()
		employee := Employee{
			Name : "José Maria",
			CPF : 9999999999 ,
			RG : 000000000 , 
			Email : "teste@teste.com",
			Age : 30 ,
			Active : true ,
			Workload : 35 ,
			IsManager : false,
		}
		if result := db.Create(&employee); result.Error != nil{
			fmt.Println("Error to create employee")
		}
		fmt.Println("Create employee !")
	}

 
