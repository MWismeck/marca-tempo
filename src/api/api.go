package api

import (
	"context"
	"time"

	"github.com/MWismeck/marca-tempo/src/db"
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
)

type API struct {
	Echo *echo.Echo
	DB   *db.EmployeeHandler
}

// @title Marca Tempo
// @version 1.0
// @description This is a sample server Marca Tempo API
// @host localhost:8080
// @BasePath /
// @schemes http
func NewServer(database *gorm.DB) *API {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080", "http://127.0.0.1:8081", "http://127.0.0.1:5500"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
	}))

	e.Static("/", "public")
	e.File("/", "public/index.html")
	employDB := db.NewEmployeeHandler(database)

	api := &API{
		Echo: e,
		DB:   employDB,
	}
	api.ConfigureRoutes()

	go api.startPeriodicTasks()

	log.Info().Msg("Server initialized successfully")
	return api
}

func (api *API) Start() error {
	log.Info().Msg("Starting server...")
	return api.Echo.Start(":8080")
}

func (api *API) Shutdown() error {
	log.Info().Msg("Shutting down server...")
	return api.Echo.Shutdown(context.Background())
}

func (api *API) startPeriodicTasks() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	api.setupNewDay()

	// Recalculate hours for existing time logs
	api.recalculateHoursForExistingLogs()

	for {
		select {
		case <-ticker.C:
			api.setupNewDay()
		}
	}
}

// recalculateHoursForExistingLogs recalculates extra hours, missing hours, and balance
// for all existing time logs that have all time fields filled
func (api *API) recalculateHoursForExistingLogs() {
	var timeLogs []schemas.TimeLog

	// Find all time logs that have all time fields filled
	if err := api.DB.DB.Where(
		"entry_time != ? AND lunch_exit_time != ? AND lunch_return_time != ? AND exit_time != ?",
		time.Time{}, time.Time{}, time.Time{}, time.Time{}).Find(&timeLogs).Error; err != nil {
		log.Error().Err(err).Msg("Failed to retrieve time logs for recalculation")
		return
	}

	log.Info().Msgf("Found %d time logs to recalculate", len(timeLogs))

	for _, timeLog := range timeLogs {
		// Get employee workload
		var employee schemas.Employee
		if err := api.DB.DB.Where("email = ?", timeLog.EmployeeEmail).First(&employee).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to retrieve employee for time log ID %d", timeLog.ID)

			// If employee not found, use a default workload of 40 hours
			employee.Workload = 40.0
			log.Warn().Msgf("Employee not found for time log ID %d, using default workload of 40 hours", timeLog.ID)
		}

		// If workload is not set, use a default of 40 hours
		if employee.Workload < 0.1 {
			employee.Workload = 40.0
			log.Warn().Msgf("Workload not set for employee %s, using default of 40 hours", timeLog.EmployeeEmail)
		}

		// Calculate extra hours, missing hours, and balance
		extraHours, missingHours, balance := api.CalculateHours(
			timeLog.EntryTime,
			timeLog.LunchExitTime,
			timeLog.LunchReturnTime,
			timeLog.ExitTime,
			employee.Workload,
		)

		// Log the time log details
		log.Info().
			Uint("timeLogID", timeLog.ID).
			Str("employeeEmail", timeLog.EmployeeEmail).
			Str("logDate", timeLog.LogDate.Format("2006-01-02")).
			Float32("workload", employee.Workload).
			Float32("dailyWorkload", employee.Workload/5).
			Float32("extraHours", extraHours).
			Float32("missingHours", missingHours).
			Float32("balance", balance).
			Msg("Recalculating hours for time log")

		// Update the time log with calculated values
		timeLog.ExtraHours = extraHours
		timeLog.MissingHours = missingHours
		timeLog.Balance = balance

		// Save the updated time log
		if err := api.DB.DB.Save(&timeLog).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to update time log ID %d", timeLog.ID)
		} else {
			log.Info().Msgf("Successfully updated time log ID %d", timeLog.ID)
		}
	}

	log.Info().Msg("Finished recalculating hours for existing time logs")
}

func (api *API) setupNewDay() {

	currentDate := time.Now().Truncate(24 * time.Hour)

	var employeeIDs []int
	if err := api.DB.DB.Table("employees").Select("id").Scan(&employeeIDs).Error; err != nil {
		log.Error().Err(err).Msg("Failed to retrieve employee IDs")
		return
	}

	for _, id := range employeeIDs {
		// Get the employee email from the ID
		var employee schemas.Employee
		if err := api.DB.DB.First(&employee, id).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to find employee with ID %d", id)
			continue
		}

		newLog := schemas.TimeLog{
			EmployeeEmail: employee.Email,
			LogDate:       currentDate,
		}

		err := api.DB.DB.Where("employee_email = ? AND log_date = ?", employee.Email, currentDate).
			FirstOrCreate(&newLog).Error
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create new log for employee %d", id)
		} else {
			log.Info().Msgf("Created new log for employee %d on %s", id, currentDate.Format("2006-01-02"))
		}
	}
}

func (api *API) ConfigureRoutes() {

	api.Echo.GET("/employees/", api.getEmployees)
	api.Echo.POST("/employee/", api.createEmployee)
	api.Echo.GET("/employee/:id", api.getEmployeeId)
	api.Echo.PUT("/employee/:id", api.updateEmployee)
	api.Echo.DELETE("/employee/:id", api.deleteEmployee)

	//  Routes time registration

	api.Echo.POST("/time_logs", api.createTimeLog)
	api.Echo.PUT("/time_logs/:id", api.punchTime)
	api.Echo.GET("/time_logs", api.getTimeLogs)
	api.Echo.GET("/time_logs/export", api.exportToExcel)
	api.Echo.DELETE("/time_logs/:id", api.deleteTimeLog)

	api.Echo.POST("/login", api.login)
	api.Echo.POST("/login/password", api.createOrUpdatePassword)

	adminGroup := api.Echo.Group("/admin")
	adminGroup.POST("/create_company", api.createCompany)
	adminGroup.GET("/companies", api.listCompanies)
	adminGroup.POST("/create_manager", api.createManager)
	adminGroup.GET("/managers", api.listManagers)
	api.Echo.PUT("/time_logs/:id/manual_edit", api.editTimeLogByManager)
	api.Echo.POST("/employee/request_change", api.requestTimeEdit)
	api.Echo.GET("/time_logs/export_range", api.exportTimeLogsRange)

	api.Echo.GET("/time-registration.html", func(c echo.Context) error {
		return c.File("public/time-registration.html")
	})

	api.Echo.GET("/swagger/*", echoSwagger.EchoWrapHandler())
}
