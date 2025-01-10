package api

import (
	"context"
	"time"
    "gorm.io/gorm"
	"github.com/MWismeck/marca-tempo/src/db"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/swaggo/echo-swagger"
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4/middleware"
	
	
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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"}, // Permitir origens específicas
        AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},       // Métodos HTTP permitidos
        AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType}, // Cabeçalhos permitidos
    }))

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

func (api *API) startPeriodicTasks() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Executa o setup imediatamente ao iniciar
	api.setupNewDay()

	for {
		select {
		case <-ticker.C:
			api.setupNewDay()
		}
	}
}

// setupNewDay cria registros para o novo dia
func (api *API) setupNewDay() {
	// Obtém a data atual
	currentDate := time.Now().Truncate(24 * time.Hour)

	// Lista todos os IDs de funcionários
	var employeeIDs []int
	if err := api.DB.DB.Table("employees").Select("id").Scan(&employeeIDs).Error; err != nil {
		log.Error().Err(err).Msg("Failed to retrieve employee IDs")
		return
	}

	// Para cada funcionário, cria um registro para o novo dia
	for _, id := range employeeIDs {
		newLog := schemas.TimeLog{
			ID : id,
			LogDate:    currentDate,
		}

		// Insere o registro somente se não existir um para o mesmo funcionário e dia
		err := api.DB.DB.Where("employee_id = ? AND log_date = ?", id, currentDate).
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


	api.Echo.POST("/time_logs/", api.createTimeLog) // Criar um novo ponto
	api.Echo.PUT("/time_logs/:id", api.updateTimeLog) // Atualizar um ponto
	api.Echo.GET("/time_logs/", api.getTimeLogs) // Buscar logs de ponto
	api.Echo.DELETE("/time_logs/:id", api.deleteTimeLog) // Deletar um ponto

	api.Echo.POST("/login", api.login)
	api.Echo.POST("/login/password", api.createOrUpdatePassword)


	api.Echo.GET("/swagger/*", echoSwagger.EchoWrapHandler())
}


