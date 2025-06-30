package api

import (
	"fmt"
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
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

	if err := c.Bind(&timeLog); err != nil {
		log.Error().Err(err).Msg("Failed to bind time log data")
		return c.String(http.StatusBadRequest, "Invalid time log data")
	}

	if timeLog.EmployeeEmail == "" {
		log.Error().Msg("Invalid employee email")
		return c.String(http.StatusBadRequest, "Invalid employee ID")
	}
	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", timeLog.EmployeeEmail).First(&employee).Error; err != nil {
		log.Error().Err(err).Msg("Employee not found")
		return c.String(http.StatusBadRequest, "Employee not found")
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

	employeeEmail := c.QueryParam("employee_email")
	if employeeEmail == "" {
		log.Error().Msg("Invalid employee email")
		return c.String(http.StatusBadRequest, "Invalid employee email")
	}

	var timeLogs []schemas.TimeLog

	if err := api.DB.DB.Where("employee_email = ?", employeeEmail).Find(&timeLogs).Error; err != nil {
		log.Error().Err(err).Msgf("Failed to retrieve time logs for employee email %s", employeeEmail)
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
func (api *API) punchTime(c echo.Context) error {

	employeeEmail := c.QueryParam("employee_email")
	if employeeEmail == "" {
		log.Error().Msg("Employee email is required")
		return c.String(http.StatusBadRequest, "Employee email is required")
	}

	// Get the current date and time in local timezone
	now := time.Now()
	// Create a date that represents today at midnight in the local timezone
	year, month, day := now.Date()
	currentDate := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	// Log the current date for debugging
	log.Info().
		Str("currentDate", currentDate.Format("2006-01-02")).
		Str("currentTime", now.Format("15:04:05")).
		Msg("Current date and time")

	var timeLog schemas.TimeLog
	if err := api.DB.DB.Where("employee_email = ? AND log_date = ?", employeeEmail, currentDate).First(&timeLog).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			timeLog = schemas.TimeLog{
				EmployeeEmail: employeeEmail,
				LogDate:       currentDate,
				EntryTime:     now,
			}
			if err := api.DB.DB.Create(&timeLog).Error; err != nil {
				log.Error().Err(err).Msg("Failed to create time log")
				return c.String(http.StatusInternalServerError, "Error creating time log")
			}

			return c.JSON(http.StatusCreated, timeLog)
		}
		log.Error().Err(err).Msg("Failed to retrieve time log")
		return c.String(http.StatusInternalServerError, "Error retrieving time log")
	}

	// Use the same 'now' variable
	if timeLog.EntryTime.IsZero() {
		timeLog.EntryTime = now
	} else if timeLog.LunchExitTime.IsZero() {
		timeLog.LunchExitTime = now
	} else if timeLog.LunchReturnTime.IsZero() {
		timeLog.LunchReturnTime = now
	} else if timeLog.ExitTime.IsZero() {
		timeLog.ExitTime = now

		// Get employee workload to calculate hours
		var employee schemas.Employee
		if err := api.DB.DB.Where("email = ?", employeeEmail).First(&employee).Error; err != nil {
			log.Error().Err(err).Msg("Failed to retrieve employee for workload calculation")
		} else {
			// Calculate extra hours, missing hours, and balance
			extraHours, missingHours, balance := api.CalculateHours(
				timeLog.EntryTime,
				timeLog.LunchExitTime,
				timeLog.LunchReturnTime,
				timeLog.ExitTime,
				employee.Workload,
			)

			// Update the time log with calculated values
			timeLog.ExtraHours = extraHours
			timeLog.MissingHours = missingHours
			timeLog.Balance = balance
		}
	} else {
		log.Warn().Msgf("All time log fields are already filled for employee %s", employeeEmail)
		return c.String(http.StatusBadRequest, "All time log fields are already filled for today")
	}

	if err := api.DB.DB.Save(&timeLog).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update time log")
		return c.String(http.StatusInternalServerError, "Error updating time log")
	}

	return c.JSON(http.StatusOK, timeLog)
}

func (api *API) CalculateHours(entryTime, lunchExitTime, lunchReturnTime, exitTime time.Time, workload float32) (extraHours, missingHours, balance float32) {
	// Check if all time fields are filled
	if entryTime.IsZero() || lunchExitTime.IsZero() || lunchReturnTime.IsZero() || exitTime.IsZero() {
		// Return zeros if any time field is not filled
		return 0, 0, 0
	}

	// If workload is not set (0 or very small), use a default of 40 hours per week
	if workload < 0.1 {
		workload = 40.0
		log.Warn().Msg("Workload not set or too small, using default of 40 hours per week")
	}

	dailyWorkload := workload / 5

	workedDuration := exitTime.Sub(entryTime) - (lunchReturnTime.Sub(lunchExitTime))

	workedHours := float32(workedDuration.Hours())

	// Log the calculation details for debugging
	log.Info().
		Float32("workload", workload).
		Float32("dailyWorkload", dailyWorkload).
		Float32("workedHours", workedHours).
		Str("entryTime", entryTime.Format(time.RFC3339)).
		Str("lunchExitTime", lunchExitTime.Format(time.RFC3339)).
		Str("lunchReturnTime", lunchReturnTime.Format(time.RFC3339)).
		Str("exitTime", exitTime.Format(time.RFC3339)).
		Msg("Calculating hours")

	// Calculate extra or missing hours based on daily workload
	// If worked hours is greater than daily workload, then there are extra hours
	// Otherwise, there are missing hours
	extraHours = 0
	missingHours = 0

	if workedHours > dailyWorkload {
		extraHours = workedHours - dailyWorkload
	} else {
		missingHours = dailyWorkload - workedHours
	}

	// Calculate balance (can be positive or negative)
	balance = extraHours - missingHours

	// Log the calculation results
	log.Info().
		Float32("extraHours", extraHours).
		Float32("missingHours", missingHours).
		Float32("balance", balance).
		Msg("Calculation results")

	return
}

// exportToExcel godoc
//
// @Summary      Export time logs to Excel
// @Description  Export all time logs for a specific employee to Excel format
// @Tags         timeLogs
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        employee_email query string true "Employee Email"
// @Success      200 {file} binary "Excel file download"
// @Failure      400 {string} string "Invalid employee email"
// @Failure      500 {string} string "Internal server error"
// @Router       /time_logs/export [get]
func (api *API) exportToExcel(c echo.Context) error {
	employeeEmail := c.QueryParam("employee_email")
	if employeeEmail == "" {
		log.Error().Msg("Invalid employee email")
		return c.String(http.StatusBadRequest, "Invalid employee email")
	}

	// Get employee details
	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", employeeEmail).First(&employee).Error; err != nil {
		log.Error().Err(err).Msg("Employee not found")
		return c.String(http.StatusBadRequest, "Employee not found")
	}

	// Get time logs for the employee
	var timeLogs []schemas.TimeLog
	if err := api.DB.DB.Where("employee_email = ?", employeeEmail).Order("log_date DESC").Find(&timeLogs).Error; err != nil {
		log.Error().Err(err).Msgf("Failed to retrieve time logs for employee email %s", employeeEmail)
		return c.String(http.StatusInternalServerError, "Error retrieving time logs")
	}

	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Excel file")
		}
	}()

	// Create a new sheet
	sheetName := "Registros de Ponto"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new sheet")
		return c.String(http.StatusInternalServerError, "Error creating Excel file")
	}
	f.SetActiveSheet(index)

	// Add employee information
	f.SetCellValue(sheetName, "A1", "Relatório de Ponto")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Funcionário: %s", employee.Name))
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Email: %s", employee.Email))
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Data de Geração: %s", time.Now().Format("02/01/2006 15:04:05")))

	// Add headers
	headers := []string{"Data", "Entrada", "Saída Almoço", "Retorno Almoço", "Saída", "Horas Extras", "Horas Faltantes", "Saldo"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c6", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Add data
	for i, log := range timeLogs {
		row := i + 7 // Start from row 7 (after headers)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), log.LogDate.Format("02/01/2006"))

		if !log.EntryTime.IsZero() {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), log.EntryTime.Format("02/01/2006 15:04:05"))
		}

		if !log.LunchExitTime.IsZero() {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), log.LunchExitTime.Format("02/01/2006 15:04:05"))
		}

		if !log.LunchReturnTime.IsZero() {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), log.LunchReturnTime.Format("02/01/2006 15:04:05"))
		}

		if !log.ExitTime.IsZero() {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), log.ExitTime.Format("02/01/2006 15:04:05"))
		}

		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("%.2f", log.ExtraHours))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("%.2f", log.MissingHours))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("%.2f", log.Balance))
	}

	styleHeader, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#C6EFCE"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err == nil {
		f.SetCellStyle(sheetName, "A6", fmt.Sprintf("%c6", 'A'+len(headers)-1), styleHeader)
	}

	// Set column width
	for i := 0; i < len(headers); i++ {
		colName := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, colName, colName, 20)
	}

	// Generate a buffer with the Excel file
	buf, err := f.WriteToBuffer()
	if err != nil {
		log.Error().Err(err).Msg("Failed to write Excel to buffer")
		return c.String(http.StatusInternalServerError, "Error generating Excel file")
	}

	// Set response headers
	fileName := fmt.Sprintf("registros_ponto_%s_%s.xlsx", employee.Name, time.Now().Format("20060102"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// Send the file
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
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

	timeLogID := c.Param("id")
	id, err := strconv.Atoi(timeLogID)
	if err != nil || id <= 0 {
		log.Error().Err(err).Msg("Invalid time log ID")
		return c.String(http.StatusBadRequest, "Invalid time log ID")
	}

	var timeLog schemas.TimeLog
	if err := api.DB.DB.First(&timeLog, id).Error; err != nil {
		log.Error().Err(err).Msgf("Time log with ID %d not found", id)
		return c.String(http.StatusNotFound, "Time log not found")
	}

	if err := api.DB.DB.Delete(&timeLog).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete time log")
		return c.String(http.StatusInternalServerError, "Error deleting time log")
	}

	return c.String(http.StatusOK, "Time log deleted")
}

func (api *API) editTimeLogByManager(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "ID inválido")
	}

	var updateData struct {
		EntryTime       string `json:"entry_time"`
		LunchExitTime   string `json:"lunch_exit_time"`
		LunchReturnTime string `json:"lunch_return_time"`
		ExitTime        string `json:"exit_time"`
		MotivoEdicao    string `json:"motivo_edicao"`
		ManagerEmail    string `json:"manager_email"`
	}

	if err := c.Bind(&updateData); err != nil {
		log.Error().Err(err).Msg("[api] Erro ao fazer bind dos dados de edição")
		return c.JSON(http.StatusBadRequest, "Dados inválidos")
	}

	log.Info().
		Int("timeLogId", id).
		Str("entryTime", updateData.EntryTime).
		Str("lunchExitTime", updateData.LunchExitTime).
		Str("lunchReturnTime", updateData.LunchReturnTime).
		Str("exitTime", updateData.ExitTime).
		Str("motivo", updateData.MotivoEdicao).
		Str("managerEmail", updateData.ManagerEmail).
		Msg("[api] Dados recebidos para edição")

	// VALIDAÇÃO: Motivo obrigatório
	if updateData.MotivoEdicao == "" {
		return c.JSON(http.StatusBadRequest, "Motivo da edição é obrigatório")
	}

	// VALIDAÇÃO: Email do gerente obrigatório
	if updateData.ManagerEmail == "" {
		return c.JSON(http.StatusBadRequest, "Email do gerente é obrigatório")
	}

	var timeLog schemas.TimeLog
	if err := api.DB.DB.First(&timeLog, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, "Registro não encontrado")
	}

	// VALIDAÇÃO: Verificar se gerente pode editar (mesma empresa)
	var manager schemas.Employee
	if err := api.DB.DB.Where("email = ? AND is_manager = ?", updateData.ManagerEmail, true).First(&manager).Error; err != nil {
		log.Error().Err(err).Msgf("[api] Gerente não encontrado: %s", updateData.ManagerEmail)
		return c.JSON(http.StatusUnauthorized, "Gerente não encontrado")
	}

	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", timeLog.EmployeeEmail).First(&employee).Error; err != nil {
		log.Error().Err(err).Msgf("[api] Funcionário não encontrado: %s", timeLog.EmployeeEmail)
		return c.JSON(http.StatusNotFound, "Funcionário não encontrado")
	}

	if manager.CompanyCNPJ != employee.CompanyCNPJ {
		log.Warn().
			Str("manager_cnpj", manager.CompanyCNPJ).
			Str("employee_cnpj", employee.CompanyCNPJ).
			Msg("[api] Tentativa de edição entre empresas diferentes")
		return c.JSON(http.StatusForbidden, "Você só pode editar funcionários da sua empresa")
	}

	// Função auxiliar para converter string para time.Time
	parseDateTime := func(dateTimeStr string) (time.Time, error) {
		if dateTimeStr == "" {
			return time.Time{}, nil
		}
		// Tenta diferentes formatos
		formats := []string{
			"2006-01-02T15:04",
			"2006-01-02T15:04:05",
			time.RFC3339,
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, dateTimeStr); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("formato de data/hora inválido: %s", dateTimeStr)
	}

	// Atualiza apenas os campos enviados
	if updateData.EntryTime != "" {
		if parsedTime, err := parseDateTime(updateData.EntryTime); err != nil {
			log.Error().Err(err).Str("entryTime", updateData.EntryTime).Msg("[api] Erro ao converter EntryTime")
			return c.JSON(http.StatusBadRequest, "Formato de data/hora inválido para EntryTime")
		} else {
			timeLog.EntryTime = parsedTime
		}
	}
	
	if updateData.LunchExitTime != "" {
		if parsedTime, err := parseDateTime(updateData.LunchExitTime); err != nil {
			log.Error().Err(err).Str("lunchExitTime", updateData.LunchExitTime).Msg("[api] Erro ao converter LunchExitTime")
			return c.JSON(http.StatusBadRequest, "Formato de data/hora inválido para LunchExitTime")
		} else {
			timeLog.LunchExitTime = parsedTime
		}
	}
	
	if updateData.LunchReturnTime != "" {
		if parsedTime, err := parseDateTime(updateData.LunchReturnTime); err != nil {
			log.Error().Err(err).Str("lunchReturnTime", updateData.LunchReturnTime).Msg("[api] Erro ao converter LunchReturnTime")
			return c.JSON(http.StatusBadRequest, "Formato de data/hora inválido para LunchReturnTime")
		} else {
			timeLog.LunchReturnTime = parsedTime
		}
	}
	
	if updateData.ExitTime != "" {
		if parsedTime, err := parseDateTime(updateData.ExitTime); err != nil {
			log.Error().Err(err).Str("exitTime", updateData.ExitTime).Msg("[api] Erro ao converter ExitTime")
			return c.JSON(http.StatusBadRequest, "Formato de data/hora inválido para ExitTime")
		} else {
			timeLog.ExitTime = parsedTime
		}
	}

	// Salvar motivo e dados da edição
	timeLog.EditadoPorGerente = manager.Name
	timeLog.EditadoEm = time.Now()
	timeLog.MotivoEdicao = updateData.MotivoEdicao

	if err := api.DB.DB.Save(&timeLog).Error; err != nil {
		log.Error().Err(err).Msg("[api] Erro ao salvar edição do time log")
		return c.JSON(http.StatusInternalServerError, "Erro ao salvar")
	}

	log.Info().
		Int("timeLogId", id).
		Str("managerEmail", updateData.ManagerEmail).
		Str("employeeEmail", timeLog.EmployeeEmail).
		Str("motivo", updateData.MotivoEdicao).
		Msg("[api] Time log editado pelo gerente")

	return c.JSON(http.StatusOK, timeLog)
}

func (api *API) requestTimeEdit(c echo.Context) error {
	var req schemas.PontoSolicitacao
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[api] Erro ao fazer bind da solicitação")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	if req.FuncionarioEmail == "" || req.Motivo == "" {
		log.Error().
			Str("funcionario_email", req.FuncionarioEmail).
			Str("motivo", req.Motivo).
			Msg("[api] Dados obrigatórios faltando")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email e motivo são obrigatórios"})
	}

	// Garantir que o status seja "pendente" se não foi definido
	if req.Status == "" {
		req.Status = "pendente"
	}

	log.Info().
		Str("funcionario_email", req.FuncionarioEmail).
		Str("motivo", req.Motivo).
		Time("data_solicitada", req.DataSolicitada).
		Str("status", req.Status).
		Msg("[api] Criando nova solicitação")

	if err := api.DB.DB.Create(&req).Error; err != nil {
		log.Error().Err(err).Msg("[api] Erro ao salvar solicitação no banco")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar solicitação"})
	}

	log.Info().
		Uint("solicitacao_id", req.ID).
		Str("funcionario_email", req.FuncionarioEmail).
		Str("status", req.Status).
		Msg("[api] Solicitação salva com sucesso")

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Solicitação registrada com sucesso",
		"id": req.ID,
		"status": req.Status,
	})
}

func (api *API) exportTimeLogsRange(c echo.Context) error {
	email := c.QueryParam("employee_email")
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	if email == "" || startStr == "" || endStr == "" {
		return c.String(http.StatusBadRequest, "Parâmetros obrigatórios: employee_email, start, end")
	}

	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)
	if err1 != nil || err2 != nil {
		return c.String(http.StatusBadRequest, "Formato de data inválido")
	}

	// Inclui o final do dia na data final
	end = end.Add(24 * time.Hour)

	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return c.String(http.StatusBadRequest, "Funcionário não encontrado")
	}

	var timeLogs []schemas.TimeLog
	if err := api.DB.DB.
		Where("employee_email = ? AND log_date BETWEEN ? AND ?", email, start, end).
		Order("log_date").
		Find(&timeLogs).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar registros")
	}

	if len(timeLogs) == 0 {
		return c.String(http.StatusNotFound, "Nenhum registro no período selecionado")
	}

	// Geração Excel (mesma lógica do exportToExcel original)
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Período"
	f.NewSheet(sheet)
	if index, err := f.GetSheetIndex(sheet); err == nil {
		f.SetActiveSheet(index)
	}

	headers := []string{"Data", "Entrada", "Saída Almoço", "Retorno", "Saída", "Extras", "Faltantes", "Saldo"}
	for i, h := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}

	for i, log := range timeLogs {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), log.LogDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), log.EntryTime.Format("15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), log.LunchExitTime.Format("15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), log.LunchReturnTime.Format("15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), log.ExitTime.Format("15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), log.ExtraHours)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), log.MissingHours)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), log.Balance)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao gerar planilha")
	}

	filename := fmt.Sprintf("Relatorio_%s_%s.xlsx", employee.Name, time.Now().Format("200601021504"))
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (api *API) getManagerRequests(c echo.Context) error {
	managerEmail := c.QueryParam("manager_email")
	if managerEmail == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email do gerente é obrigatório"})
	}

	log.Info().
		Str("managerEmail", managerEmail).
		Msg("[api] Iniciando busca de solicitações para gerente")

	// Buscar empresa do gerente
	var manager schemas.Employee
	if err := api.DB.DB.Where("email = ? AND is_manager = ?", managerEmail, true).First(&manager).Error; err != nil {
		log.Error().Err(err).Msgf("[api] Gerente não encontrado: %s", managerEmail)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Gerente não encontrado"})
	}

	log.Info().
		Str("managerEmail", managerEmail).
		Str("managerName", manager.Name).
		Str("companyCNPJ", manager.CompanyCNPJ).
		Bool("isManager", manager.IsManager).
		Msg("[api] Gerente encontrado")

	// Buscar funcionários da mesma empresa
	var employeeEmails []string
	if err := api.DB.DB.Model(&schemas.Employee{}).
		Where("company_cnpj = ?", manager.CompanyCNPJ).
		Pluck("email", &employeeEmails).Error; err != nil {
		log.Error().Err(err).Msg("[api] Erro ao buscar funcionários da empresa")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar funcionários"})
	}

	log.Info().
		Str("companyCNPJ", manager.CompanyCNPJ).
		Int("employeeCount", len(employeeEmails)).
		Strs("employeeEmails", employeeEmails).
		Msg("[api] Funcionários da empresa encontrados")

	if len(employeeEmails) == 0 {
		log.Warn().
			Str("managerEmail", managerEmail).
			Str("companyCNPJ", manager.CompanyCNPJ).
			Msg("[api] Nenhum funcionário encontrado na empresa")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"pending":   []schemas.PontoSolicitacao{},
			"processed": []schemas.PontoSolicitacao{},
		})
	}

	// Buscar solicitações dos funcionários da empresa
	var allRequests []schemas.PontoSolicitacao
	if err := api.DB.DB.Where("funcionario_email IN ?", employeeEmails).
		Order("created_at DESC").
		Find(&allRequests).Error; err != nil {
		log.Error().Err(err).Msg("[api] Erro ao buscar solicitações")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar solicitações"})
	}

	log.Info().
		Int("totalRequests", len(allRequests)).
		Msg("[api] Total de solicitações encontradas")

	// Log detalhado de cada solicitação
	for i, req := range allRequests {
		log.Info().
			Int("index", i).
			Uint("requestId", req.ID).
			Str("funcionarioEmail", req.FuncionarioEmail).
			Str("status", req.Status).
			Str("motivo", req.Motivo).
			Time("createdAt", req.CreatedAt).
			Msg("[api] Solicitação encontrada")
	}

	// Separar solicitações pendentes e processadas
	var pending []schemas.PontoSolicitacao
	var processed []schemas.PontoSolicitacao

	for _, req := range allRequests {
		if req.Status == "pendente" {
			pending = append(pending, req)
		} else {
			processed = append(processed, req)
		}
	}

	log.Info().
		Int("pendingCount", len(pending)).
		Int("processedCount", len(processed)).
		Msg("[api] Solicitações separadas por status")

	// Buscar nomes dos funcionários para enriquecer os dados
	type RequestWithEmployeeName struct {
		schemas.PontoSolicitacao
		FuncionarioNome string `json:"funcionario_nome"`
	}

	var pendingWithNames []RequestWithEmployeeName
	var processedWithNames []RequestWithEmployeeName

	for _, req := range pending {
		var employee schemas.Employee
		if err := api.DB.DB.Where("email = ?", req.FuncionarioEmail).First(&employee).Error; err == nil {
			pendingWithNames = append(pendingWithNames, RequestWithEmployeeName{
				PontoSolicitacao: req,
				FuncionarioNome:  employee.Name,
			})
			log.Info().
				Uint("requestId", req.ID).
				Str("funcionarioEmail", req.FuncionarioEmail).
				Str("funcionarioNome", employee.Name).
				Msg("[api] Nome do funcionário adicionado à solicitação pendente")
		} else {
			log.Error().Err(err).
				Uint("requestId", req.ID).
				Str("funcionarioEmail", req.FuncionarioEmail).
				Msg("[api] Erro ao buscar nome do funcionário para solicitação pendente")
		}
	}

	for _, req := range processed {
		var employee schemas.Employee
		if err := api.DB.DB.Where("email = ?", req.FuncionarioEmail).First(&employee).Error; err == nil {
			processedWithNames = append(processedWithNames, RequestWithEmployeeName{
				PontoSolicitacao: req,
				FuncionarioNome:  employee.Name,
			})
		}
	}

	log.Info().
		Str("managerEmail", managerEmail).
		Int("finalPendingCount", len(pendingWithNames)).
		Int("finalProcessedCount", len(processedWithNames)).
		Msg("[api] Solicitações finais carregadas para gerente")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"pending":   pendingWithNames,
		"processed": processedWithNames,
	})
}

func (api *API) updateRequestStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	var updateData struct {
		Status            string `json:"status"`
		ComentarioGerente string `json:"comentario_gerente"`
		GerenteEmail      string `json:"gerente_email"`
	}

	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Validações
	if updateData.Status != "aprovado" && updateData.Status != "rejeitado" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Status deve ser 'aprovado' ou 'rejeitado'"})
	}

	if updateData.GerenteEmail == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email do gerente é obrigatório"})
	}

	if updateData.ComentarioGerente == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comentário do gerente é obrigatório"})
	}

	// Buscar a solicitação
	var request schemas.PontoSolicitacao
	if err := api.DB.DB.First(&request, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Solicitação não encontrada"})
	}

	// Verificar se ainda está pendente
	if request.Status != "pendente" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Solicitação já foi processada"})
	}

	// Verificar se o gerente pode processar (mesma empresa)
	var manager schemas.Employee
	if err := api.DB.DB.Where("email = ? AND is_manager = ?", updateData.GerenteEmail, true).First(&manager).Error; err != nil {
		log.Error().Err(err).Msgf("[api] Gerente não encontrado: %s", updateData.GerenteEmail)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Gerente não encontrado"})
	}

	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", request.FuncionarioEmail).First(&employee).Error; err != nil {
		log.Error().Err(err).Msgf("[api] Funcionário não encontrado: %s", request.FuncionarioEmail)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Funcionário não encontrado"})
	}

	if manager.CompanyCNPJ != employee.CompanyCNPJ {
		log.Warn().
			Str("manager_cnpj", manager.CompanyCNPJ).
			Str("employee_cnpj", employee.CompanyCNPJ).
			Msg("[api] Tentativa de processar solicitação entre empresas diferentes")
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Você só pode processar solicitações de funcionários da sua empresa"})
	}

	// Atualizar a solicitação
	request.Status = updateData.Status
	request.ComentarioGerente = updateData.ComentarioGerente
	request.GerenteEmail = updateData.GerenteEmail
	request.ProcessadoEm = time.Now()

	if err := api.DB.DB.Save(&request).Error; err != nil {
		log.Error().Err(err).Msg("[api] Erro ao salvar solicitação processada")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao processar solicitação"})
	}

	log.Info().
		Int("requestId", id).
		Str("status", updateData.Status).
		Str("managerEmail", updateData.GerenteEmail).
		Str("employeeEmail", request.FuncionarioEmail).
		Str("comentario", updateData.ComentarioGerente).
		Msg("[api] Solicitação processada pelo gerente")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Solicitação processada com sucesso",
		"request": request,
	})
}
