package db

import (
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type EmployeeHandler struct {
	DB *gorm.DB
}

func Init() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("employee.db"), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to initialize SQLite: %s", err.Error())
	}
	db.AutoMigrate(&schemas.Employee{}, &schemas.Login{}, &schemas.TimeLog{})
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
	err := e.DB.First(&employee, id).Error
	return employee, err
}

func (e *EmployeeHandler) UpdateEmployee(updateEmployee schemas.Employee) error {

	return e.DB.Save(&updateEmployee).Error
}

func (e *EmployeeHandler) DeleteEmployee(employee schemas.Employee) error {

	return e.DB.Delete(&employee).Error
}

func (e *EmployeeHandler) GetFilteredEmployee(active bool) ([]schemas.Employee, error) {
	filteredEmployees := []schemas.Employee{}

	err := e.DB.Where("active= ?", active).Find(&filteredEmployees).Error

	return filteredEmployees, err
}


func (e *EmployeeHandler) AddTimeLog(timeLog schemas.TimeLog) error {
	if result := e.DB.Create(&timeLog); result.Error != nil {
		log.Error().Msg("Failed to create time log")
		return result.Error
	}
	log.Info().Msg("Time log created successfully")
	return nil
}

func (e *EmployeeHandler) UpdateExitTime(timeLogID uint, exitTime time.Time) error {
	var timeLog schemas.TimeLog
	if err := e.DB.First(&timeLog, timeLogID).Error; err != nil {
		return err
	}

	timeLog.ExitTime = exitTime
	if err := e.DB.Save(&timeLog).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update exit time")
		return err
	}
	log.Info().Msg("Exit time updated successfully")
	return nil
}


func (e *EmployeeHandler) GetTimeLogsByEmployeeID(employeeID uint) ([]schemas.TimeLog, error) {
	// First get the employee to find their email
	var employee schemas.Employee
	if err := e.DB.First(&employee, employeeID).Error; err != nil {
		return nil, err
	}
	
	var timeLogs []schemas.TimeLog
	err := e.DB.Where("employee_email = ?", employee.Email).Find(&timeLogs).Error
	return timeLogs, err
}

func (e *EmployeeHandler) GetTimeLogsByEmail(email string) ([]schemas.TimeLog, error) {
	var timeLogs []schemas.TimeLog
	err := e.DB.Where("employee_email = ?", email).Find(&timeLogs).Error
	return timeLogs, err
}
