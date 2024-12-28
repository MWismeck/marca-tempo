package db

import (
	"github.com/MWismeck/marca-tempo/schemas"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type EmployeeHandler struct {
	DB *gorm.DB
}

func Init() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("employee.db"), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to initialize SQLite: %s", err.Error())
	}
	db.AutoMigrate(&schemas.Employee{})
	return db
}

func NewEmployeeHandler(db *gorm.DB) *EmployeeHandler {
	return &EmployeeHandler{DB: db}
}

func (e *EmployeeHandler) AddEmployee(employee schemas.Employee) error {
	if result := e.DB.Create(&employee); result.Error != nil {
		log.Error().Msg("Failed to create employee")
		return result.Error
	}
	log.Info().Msg("Create Employee!")
	return nil
}

func (e *EmployeeHandler) GetEmployees() ([]schemas.Employee, error) {
	employees := []schemas.Employee{}

	err := e.DB.Find(&employees).Error
	return employees, err
}

func (e *EmployeeHandler) GetEmployee(id int) (schemas.Employee, error) {
	var employee schemas.Employee
	err := e.DB.First(&employee, id)
	return employee, err.Error
}

func (e *EmployeeHandler) UpdateEmployee(updateEmployee schemas.Employee) error {

	return e.DB.Save(&updateEmployee).Error
}

func (e *EmployeeHandler) DeleteEmployee(employee schemas.Employee) error {

	return e.DB.Delete(&employee).Error
}
