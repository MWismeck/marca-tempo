package api

import (
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"time"
)

// createTimeLog godoc
//
// @Summary      Create a time log
// @Description  Create a new time log entry for an employee
// @Tags         timeLogs
// @Accept       json
// @Produce      json
// @Param        body body schemas.TimeLog true "Time Log Data"
// @Success      201 {object} schemas.TimeLog
// @Failure      400 {string} string "Invalid time log data"
// @Failure      500 {string} string "Internal server error"
// @Router       /timeLogs [post]
func (api *API) createTimeLog(c echo.Context) error {
	timeLog := schemas.TimeLog{}

	// Bind e validação inicial
	if err := c.Bind(&timeLog); err != nil {
		log.Error().Err(err).Msg("Failed to bind time log data")
		return c.String(http.StatusBadRequest, "Invalid time log data")
	}

	if timeLog.EmployeeID == 0 {
		log.Error().Msg("Invalid employee ID")
		return c.String(http.StatusBadRequest, "Invalid employee ID")
	}
	if err := api.DB.DB.Create(&timeLog).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create time log")
		return c.String(http.StatusInternalServerError, "Error creating time log")
	}

	return c.JSON(http.StatusCreated, timeLog)
}

// getTimeLogs godoc
//
// @Summary      Get time logs
// @Description  Retrieve all time logs for a specific employee
// @Tags         timeLogs
// @Accept       json
// @Produce      json
// @Param        employee_id query int true "Employee ID"
// @Success      200 {array} schemas.TimeLog
// @Failure      400 {string} string "Invalid employee ID"
// @Failure      500 {string} string "Internal server error"
// @Router       /timeLogs [get]
func (api *API) getTimeLogs(c echo.Context) error {
	employeeIDStr := c.QueryParam("employee_id")

	// Validação de entrada
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		log.Error().Err(err).Msg("Invalid employee ID")
		return c.String(http.StatusBadRequest, "Invalid employee ID")
	}

	var timeLogs []schemas.TimeLog

	// Consulta no banco
	if err := api.DB.DB.Where("employee_id = ?", employeeID).Find(&timeLogs).Error; err != nil {
		log.Error().Err(err).Msgf("Failed to retrieve time logs for employee ID %d", employeeID)
		return c.String(http.StatusInternalServerError, "Error retrieving time logs")
	}

	return c.JSON(http.StatusOK, timeLogs)

}
// updateTimeLog godoc
//
// @Summary      Update a time log
// @Description  Update an existing time log entry for an employee
// @Tags         timeLogs
// @Accept       json
// @Produce      json
// @Param        id   path      int              true "Time Log ID"
// @Param        body body      schemas.TimeLog  true "Updated Time Log Data"
// @Success      200  {object}  schemas.TimeLog
// @Failure      400  {string}  string            "Invalid time log data"
// @Failure      404  {string}  string            "Time log not found"
// @Failure      500  {string}  string            "Internal server error"
// @Router       /timeLogs/{id} [put]
func (api *API) updateTimeLog(c echo.Context) error {
	// Extrai o ID do time log da URL
	timeLogID := c.Param("id")
	id, err := strconv.Atoi(timeLogID)
	if err != nil || id <= 0 {
		log.Error().Err(err).Msg("Invalid time log ID")
		return c.String(http.StatusBadRequest, "Invalid time log ID")
	}

	// Bind e validação dos dados recebidos no corpo da requisição
	var updatedTimeLog schemas.TimeLog
	if err := c.Bind(&updatedTimeLog); err != nil {
		log.Error().Err(err).Msg("Failed to bind time log data")
		return c.String(http.StatusBadRequest, "Invalid time log data")
	}

	// Busca o time log existente
	var existingTimeLog schemas.TimeLog
	if err := api.DB.DB.First(&existingTimeLog, id).Error; err != nil {
		log.Error().Err(err).Msgf("Time log with ID %d not found", id)
		return c.String(http.StatusNotFound, "Time log not found")
	}

	// Atualiza os campos de horário se necessário, e caso não seja passado, usa o horário atual
	if updatedTimeLog.EntryTime.IsZero() {
		updatedTimeLog.EntryTime = time.Now() // Hora atual do sistema
	}
	if updatedTimeLog.LunchExitTime.IsZero() {
		updatedTimeLog.LunchExitTime = time.Now() // Hora atual do sistema
	}
	if updatedTimeLog.LunchReturnTime.IsZero() {
		updatedTimeLog.LunchReturnTime = time.Now() // Hora atual do sistema
	}
	if updatedTimeLog.ExitTime.IsZero() {
		updatedTimeLog.ExitTime = time.Now() // Hora atual do sistema
	}

	// Atualiza os campos de horário no registro existente
	if !updatedTimeLog.EntryTime.IsZero() {
		existingTimeLog.EntryTime = updatedTimeLog.EntryTime
	}
	if !updatedTimeLog.ExitTime.IsZero() {
		existingTimeLog.ExitTime = updatedTimeLog.ExitTime
	}
	if !updatedTimeLog.LunchExitTime.IsZero() {
		existingTimeLog.LunchExitTime = updatedTimeLog.LunchExitTime
	}
	if !updatedTimeLog.LunchReturnTime.IsZero() {
		existingTimeLog.LunchReturnTime = updatedTimeLog.LunchReturnTime
	}

	// Defina um valor fixo para a carga horária diária
	workload := float32(8) // Exemplo: 8 horas de carga horária por dia

	// Calcula as horas extras, faltantes e o saldo
	extraHours, missingHours, balance := calculateHours(existingTimeLog.EntryTime, existingTimeLog.LunchExitTime, existingTimeLog.LunchReturnTime, existingTimeLog.ExitTime, workload)
	existingTimeLog.ExtraHours = extraHours
	existingTimeLog.MissingHours = missingHours
	existingTimeLog.Balance = balance

	// Salva as alterações
	if err := api.DB.DB.Save(&existingTimeLog).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update time log")
		return c.String(http.StatusInternalServerError, "Error updating time log")
	}

	// Retorna o time log atualizado
	return c.JSON(http.StatusOK, existingTimeLog)
}

// Função para calcular horas extras, horas faltantes e saldo
func calculateHours(entryTime, lunchExitTime, lunchReturnTime, exitTime time.Time, workload float32) (extraHours, missingHours, balance float32) {
	// Calcular o tempo trabalhado total, subtraindo a pausa para o almoço
	workedDuration := exitTime.Sub(entryTime) - (lunchReturnTime.Sub(lunchExitTime))

	// Converter a duração para float32 (em horas)
	workedHours := float32(workedDuration.Hours())

	// Calcular horas extras e faltantes
	if workedHours > workload {
		extraHours = workedHours - workload
		missingHours = 0
	} else {
		extraHours = 0
		missingHours = workload - workedHours
	}

	// Calcular o saldo de horas
	balance = extraHours - missingHours
	return
}


// deleteTimeLog godoc
//
// @Summary      Delete a time log
// @Description  Delete an existing time log entry for an employee
// @Tags         timeLogs
// @Param        id path int true "Time Log ID"
// @Success      200 {string} string "Time log deleted"
// @Failure      404 {string} string "Time log not found"
// @Failure      500 {string} string "Internal server error"
// @Router       /timeLogs/{id} [delete]
func (api *API) deleteTimeLog(c echo.Context) error {
	// Extrai o ID do time log da URL
	timeLogID := c.Param("id")
	id, err := strconv.Atoi(timeLogID)
	if err != nil || id <= 0 {
		log.Error().Err(err).Msg("Invalid time log ID")
		return c.String(http.StatusBadRequest, "Invalid time log ID")
	}

	// Busca o time log
	var timeLog schemas.TimeLog
	if err := api.DB.DB.First(&timeLog, id).Error; err != nil {
		log.Error().Err(err).Msgf("Time log with ID %d not found", id)
		return c.String(http.StatusNotFound, "Time log not found")
	}

	// Deleta o time log
	if err := api.DB.DB.Delete(&timeLog).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete time log")
		return c.String(http.StatusInternalServerError, "Error deleting time log")
	}

	// Retorna uma mensagem de sucesso
	return c.String(http.StatusOK, "Time log deleted")
}



