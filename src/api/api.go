package api

import (
	"context"
	"time"
    "gorm.io/gorm"
	"github.com/MWismeck/marca-tempo/src/db"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/swaggo/echo-swagger"
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
	// Inicializa o Echo
	e := echo.New()

	// Cria o EmployeeHandler com a instância do banco de dados
	employDB := db.NewEmployeeHandler(database)

	// Configura o servidor e rotas
	api := &API{
		Echo: e,
		DB:   employDB,
	}
	api.ConfigureRoutes()

	// Inicia tarefas periódicas em uma goroutine
	go api.startPeriodicTasks()

	log.Info().Msg("Server initialized successfully")
	return api
}

// Start inicia o servidor
func (api *API) Start() error {
	log.Info().Msg("Starting server...")
	return api.Echo.Start(":8080") // Porta do servidor
}

// Shutdown encerra o servidor
func (api *API) Shutdown() error {
	log.Info().Msg("Shutting down server...")
	return api.Echo.Shutdown(context.Background())
}

// startPeriodicTasks gerencia tarefas como reset de logs de ponto
func (api *API) startPeriodicTasks() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Executa o reset imediatamente ao iniciar
	api.resetTimeLogs()

	for {
		select {
		case <-ticker.C:
			api.resetTimeLogs()
		}
	}
}

// resetTimeLogs limpa os logs de ponto no banco de dados
func (api *API) resetTimeLogs() {
	if err := api.DB.DB.Exec("DELETE FROM time_logs").Error; err != nil {
		log.Error().Err(err).Msg("Error resetting time logs")
	} else {
		log.Info().Msg("Time logs reset successfully")
	}
}

func (api *API) ConfigureRoutes() {

	api.Echo.GET("/employees/", api.getEmployees)
	api.Echo.POST("/employee/", api.createEmployee)
	api.Echo.GET("/employee/:id", api.getEmployeeId)
	api.Echo.PUT("/employee/:id", api.updateEmployee)
	api.Echo.DELETE("/employee/:id", api.deleteEmployee)

	//  Routes time registration


	api.Echo.POST("/time_logs/", api.createTimeLog) // Criar um novo ponto
	api.Echo.PUT("/time_logs/:id", api.updateTimeLog) // Atualizar um ponto
	api.Echo.GET("/time_logs/", api.getTimeLogs) // Buscar logs de ponto
	api.Echo.DELETE("/time_logs/:id", api.deleteTimeLog) // Deletar um ponto


	api.Echo.GET("/swagger/*", echoSwagger.EchoWrapHandler())
}


