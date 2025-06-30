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
//	@Summary		Criar registro de ponto
//	@Description	Cria um novo registro de ponto para um funcionário
//	@Tags			timeLogs
//	@Accept			json
//	@Produce		json
//	@Param			body	body		schemas.TimeLog	true	"Dados do registro de ponto"
//	@Success		201		{object}	schemas.TimeLog
//	@Failure		400		{string}	string	"Dados inválidos ou funcionário não encontrado"
//	@Failure		500		{string}	string	"Erro interno do servidor"
//	@Router			/time_logs [post]
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
//	@Summary		Buscar registros de ponto
//	@Description	Retorna registros de ponto de um funcionário específico
//	@Tags			timeLogs
//	@Accept			json
//	@Produce		json
//	@Param			employee_email	query	string	true	"Email do funcionário"
//	@Success		200				{array}	schemas.TimeLog
//	@Failure		400				{string}	string	"Email do funcionário inválido"
//	@Failure		500				{string}	string	"Erro interno do servidor"
//	@Router			/time_logs [get]
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

// punchTime godoc
//
//	@Summary		Registrar ponto
//	@Description	Registra entrada, saída para almoço, retorno do almoço ou saída
//	@Tags			timeLogs
//	@Accept			json
//	@Produce		json
//	@Param			id				path	int		true	"ID do registro de ponto"
//	@Param			employee_email	query	string	true	"Email do funcionário"
//	@Success		200				{object}	schemas.TimeLog
//	@Success		201				{object}	schemas.TimeLog
//	@Failure		400				{string}	string	"Dados inválidos ou todos os pontos já registrados"
//	@Failure		500				{string}	string	"Erro interno do servidor"
//	@Router			/time_logs/{id} [put]
func (api *API) punchTime(c echo.Context) error {

	employeeEmail := c.QueryParam("employee_email")
	if employeeEmail == "" {
		log.Error().Msg("Employee email is required")
		return c.String(http.StatusBadRequest, "Employee email is required")
	}

	now := time.Now()
	year, month, day := now.Date()
	currentDate := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

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

	if timeLog.EntryTime.IsZero() {
		timeLog.EntryTime = now
	} else if timeLog.LunchExitTime.IsZero() {
		timeLog.LunchExitTime = now
	} else if timeLog.LunchReturnTime.IsZero() {
		timeLog.LunchReturnTime = now
	} else if timeLog.ExitTime.IsZero() {
		timeLog.ExitTime = now

		var employee schemas.Employee
		if err := api.DB.DB.Where("email = ?", employeeEmail).First(&employee).Error; err != nil {
			log.Error().Err(err).Msg("Failed to retrieve employee for workload calculation")
		} else {
			extraHours, missingHours, balance := api.CalculateHours(
				timeLog.EntryTime,
				timeLog.LunchExitTime,
				timeLog.LunchReturnTime,
				timeLog.ExitTime,
				employee.Workload,
			)

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
	if entryTime.IsZero() || lunchExitTime.IsZero() || lunchReturnTime.IsZero() || exitTime.IsZero() {
		return 0, 0, 0
	}

	if workload < 0.1 {
		workload = 40.0
		log.Warn().Msg("Workload not set or too small, using default of 40 hours per week")
	}

	dailyWorkload := workload / 5

	workedDuration := exitTime.Sub(entryTime) - (lunchReturnTime.Sub(lunchExitTime))

	workedHours := float32(workedDuration.Hours())

	log.Info().
		Float32("workload", workload).
		Float32("dailyWorkload", dailyWorkload).
		Float32("workedHours", workedHours).
		Str("entryTime", entryTime.Format(time.RFC3339)).
		Str("lunchExitTime", lunchExitTime.Format(time.RFC3339)).
		Str("lunchReturnTime", lunchReturnTime.Format(time.RFC3339)).
		Str("exitTime", exitTime.Format(time.RFC3339)).
		Msg("Calculating hours")

	extraHours = 0
	missingHours = 0

	if workedHours > dailyWorkload {
		extraHours = workedHours - dailyWorkload
	} else {
		missingHours = dailyWorkload - workedHours
	}

	balance = extraHours - missingHours

	log.Info().
		Float32("extraHours", extraHours).
		Float32("missingHours", missingHours).
		Float32("balance", balance).
		Msg("Calculation results")

	return
}

// exportToExcel godoc
//
//	@Summary		Exportar registros para Excel
//	@Description	Exporta todos os registros de ponto de um funcionário para Excel
//	@Tags			export
//	@Accept			json
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			employee_email	query	string	true	"Email do funcionário"
//	@Success		200				{file}	binary	"Arquivo Excel gerado com sucesso"
//	@Failure		400				{string}	string	"Email inválido ou funcionário não encontrado"
//	@Failure		500				{string}	string	"Erro interno do servidor"
//	@Router			/time_logs/export [get]
func (api *API) exportToExcel(c echo.Context) error {
	employeeEmail := c.QueryParam("employee_email")
	if employeeEmail == "" {
		log.Error().Msg("Invalid employee email")
		return c.String(http.StatusBadRequest, "Invalid employee email")
	}

	var employee schemas.Employee
	if err := api.DB.DB.Where("email = ?", employeeEmail).First(&employee).Error; err != nil {
		log.Error().Err(err).Msg("Employee not found")
		return c.String(http.StatusBadRequest, "Employee not found")
	}

	var timeLogs []schemas.TimeLog
	if err := api.DB.DB.Where("employee_email = ?", employeeEmail).Order("log_date DESC").Find(&timeLogs).Error; err != nil {
		log.Error().Err(err).Msgf("Failed to retrieve time logs for employee email %s", employeeEmail)
		return c.String(http.StatusInternalServerError, "Error retrieving time logs")
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Excel file")
		}
	}()

	sheetName := "Registros de Ponto"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new sheet")
		return c.String(http.StatusInternalServerError, "Error creating Excel file")
	}
	f.SetActiveSheet(index)

	f.SetCellValue(sheetName, "A1", "Relatório de Ponto")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Funcionário: %s", employee.Name))
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Email: %s", employee.Email))
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Data de Geração: %s", time.Now().Format("02/01/2006 15:04:05")))

	headers := []string{"Data", "Entrada", "Saída Almoço", "Retorno Almoço", "Saída", "Horas Extras", "Horas Faltantes", "Saldo", "Status", "Editado Por", "Data Edição", "Motivo Edição"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c6", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, log := range timeLogs {
		row := i + 7

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), log.LogDate.Format("02/01/2006"))

		// Formatação especial para registros editados
		isEdited := log.EditadoPorGerente != ""
		timeFormat := "15:04"
		if isEdited {
			timeFormat = "15:04*" // Adiciona asterisco para registros editados
		}

		if !log.EntryTime.IsZero() {
			timeStr := log.EntryTime.Format(timeFormat)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), timeStr)
		}

		if !log.LunchExitTime.IsZero() {
			timeStr := log.LunchExitTime.Format(timeFormat)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), timeStr)
		}

		if !log.LunchReturnTime.IsZero() {
			timeStr := log.LunchReturnTime.Format(timeFormat)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), timeStr)
		}

		if !log.ExitTime.IsZero() {
			timeStr := log.ExitTime.Format(timeFormat)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), timeStr)
		}

		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("%.2f", log.ExtraHours))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("%.2f", log.MissingHours))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("%.2f", log.Balance))

		// Colunas de informações de edição
		if isEdited {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "EDITADO")
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), log.EditadoPorGerente)
			if !log.EditadoEm.IsZero() {
				f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), log.EditadoEm.Format("02/01/2006 15:04"))
			}
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), log.MotivoEdicao)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "ORIGINAL")
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), "-")
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), "-")
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "-")
		}
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

	for i := 0; i < len(headers); i++ {
		colName := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, colName, colName, 20)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		log.Error().Err(err).Msg("Failed to write Excel to buffer")
		return c.String(http.StatusInternalServerError, "Error generating Excel file")
	}

	fileName := fmt.Sprintf("registros_ponto_%s_%s.xlsx", employee.Name, time.Now().Format("20060102"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// deleteTimeLog godoc
//
//	@Summary		Excluir registro de ponto
//	@Description	Remove um registro de ponto do sistema
//	@Tags			timeLogs
//	@Param			id	path		int	true	"ID do registro de ponto"
//	@Success		200	{string}	string	"Registro excluído com sucesso"
//	@Failure		400	{string}	string	"ID inválido"
//	@Failure		404	{string}	string	"Registro não encontrado"
//	@Failure		500	{string}	string	"Erro interno do servidor"
//	@Router			/time_logs/{id} [delete]
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

// editTimeLogByManager godoc
//
//	@Summary		Editar registro manualmente
//	@Description	Permite ao gerente editar manualmente um registro de ponto
//	@Tags			manager
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"ID do registro de ponto"
//	@Param			body	body	ManualEditRequest	true	"Dados para edição manual"
//	@Success		200		{object}	schemas.TimeLog
//	@Failure		400		{string}	string	"Dados inválidos ou motivo obrigatório"
//	@Failure		401		{string}	string	"Gerente não encontrado"
//	@Failure		403		{string}	string	"Sem permissão para editar funcionários de outra empresa"
//	@Failure		404		{string}	string	"Registro não encontrado"
//	@Failure		500		{string}	string	"Erro interno do servidor"
//	@Router			/time_logs/{id}/manual_edit [put]
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

	if updateData.MotivoEdicao == "" {
		return c.JSON(http.StatusBadRequest, "Motivo da edição é obrigatório")
	}

	if updateData.ManagerEmail == "" {
		return c.JSON(http.StatusBadRequest, "Email do gerente é obrigatório")
	}

	var timeLog schemas.TimeLog
	if err := api.DB.DB.First(&timeLog, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, "Registro não encontrado")
	}

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

	parseDateTime := func(dateTimeStr string) (time.Time, error) {
		if dateTimeStr == "" {
			return time.Time{}, nil
		}
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

	timeLog.EditadoPorGerente = manager.Name
	timeLog.EditadoEm = time.Now()
	timeLog.MotivoEdicao = updateData.MotivoEdicao

	// Recalcular horas se todos os horários estão preenchidos
	if !timeLog.EntryTime.IsZero() && !timeLog.LunchExitTime.IsZero() && 
	   !timeLog.LunchReturnTime.IsZero() && !timeLog.ExitTime.IsZero() {
		extraHours, missingHours, balance := api.CalculateHours(
			timeLog.EntryTime,
			timeLog.LunchExitTime,
			timeLog.LunchReturnTime,
			timeLog.ExitTime,
			employee.Workload,
		)
		
		timeLog.ExtraHours = extraHours
		timeLog.MissingHours = missingHours
		timeLog.Balance = balance
		
		log.Info().
			Int("timeLogId", id).
			Float32("extraHours", extraHours).
			Float32("missingHours", missingHours).
			Float32("balance", balance).
			Msg("[api] Horas recalculadas após edição")
	}

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

// requestTimeEdit godoc
//
//	@Summary		Solicitar alteração de ponto
//	@Description	Funcionário solicita alteração em seu registro de ponto
//	@Tags			manager
//	@Accept			json
//	@Produce		json
//	@Param			body	body		schemas.PontoSolicitacao	true	"Dados da solicitação"
//	@Success		201		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/employee/request_change [post]
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

// exportTimeLogsRange godoc
//
//	@Summary		Exportar registros por período
//	@Description	Exporta registros de ponto de um funcionário em um período específico
//	@Tags			export
//	@Accept			json
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			employee_email	query	string	true	"Email do funcionário"
//	@Param			start			query	string	true	"Data de início (YYYY-MM-DD)"
//	@Param			end				query	string	true	"Data de fim (YYYY-MM-DD)"
//	@Success		200				{file}	binary	"Arquivo Excel gerado com sucesso"
//	@Failure		400				{string}	string	"Parâmetros obrigatórios ou formato de data inválido"
//	@Failure		404				{string}	string	"Nenhum registro no período selecionado"
//	@Failure		500				{string}	string	"Erro interno do servidor"
//	@Router			/time_logs/export_range [get]
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

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Período"
	f.NewSheet(sheet)
	if index, err := f.GetSheetIndex(sheet); err == nil {
		f.SetActiveSheet(index)
	}

	headers := []string{"Data", "Entrada", "Saída Almoço", "Retorno", "Saída", "Extras", "Faltantes", "Saldo", "Status", "Editado Por", "Data Edição", "Motivo Edição"}
	for i, h := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}

	for i, log := range timeLogs {
		row := i + 2
		isEdited := log.EditadoPorGerente != ""
		
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), log.LogDate.Format("02/01/2006"))
		
		// Formatação com asterisco para registros editados
		timeFormat := "15:04"
		if isEdited {
			timeFormat = "15:04*"
		}
		
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), log.EntryTime.Format(timeFormat))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), log.LunchExitTime.Format(timeFormat))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), log.LunchReturnTime.Format(timeFormat))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), log.ExitTime.Format(timeFormat))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), log.ExtraHours)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), log.MissingHours)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), log.Balance)
		
		// Colunas de informações de edição
		if isEdited {
			f.SetCellValue(sheet, fmt.Sprintf("I%d", row), "EDITADO")
			f.SetCellValue(sheet, fmt.Sprintf("J%d", row), log.EditadoPorGerente)
			if !log.EditadoEm.IsZero() {
				f.SetCellValue(sheet, fmt.Sprintf("K%d", row), log.EditadoEm.Format("02/01/2006 15:04"))
			}
			f.SetCellValue(sheet, fmt.Sprintf("L%d", row), log.MotivoEdicao)
		} else {
			f.SetCellValue(sheet, fmt.Sprintf("I%d", row), "ORIGINAL")
			f.SetCellValue(sheet, fmt.Sprintf("J%d", row), "-")
			f.SetCellValue(sheet, fmt.Sprintf("K%d", row), "-")
			f.SetCellValue(sheet, fmt.Sprintf("L%d", row), "-")
		}
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

// getManagerRequests godoc
//
//	@Summary		Buscar solicitações do gerente
//	@Description	Retorna solicitações de alteração de ponto para o gerente
//	@Tags			manager
//	@Accept			json
//	@Produce		json
//	@Param			manager_email	query	string	true	"Email do gerente"
//	@Success		200				{object}	map[string]interface{}
//	@Failure		400				{object}	map[string]string
//	@Failure		401				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/manager/requests [get]
func (api *API) getManagerRequests(c echo.Context) error {
	managerEmail := c.QueryParam("manager_email")
	if managerEmail == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email do gerente é obrigatório"})
	}

	log.Info().
		Str("managerEmail", managerEmail).
		Msg("[api] Iniciando busca de solicitações para gerente")

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

// updateRequestStatus godoc
//
//	@Summary		Atualizar status da solicitação
//	@Description	Gerente aprova ou rejeita uma solicitação de alteração
//	@Tags			manager
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"ID da solicitação"
//	@Param			body	body	UpdateRequestStatusRequest	true	"Dados para atualização do status"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/manager/requests/{id}/status [put]
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

	if updateData.Status != "aprovado" && updateData.Status != "rejeitado" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Status deve ser 'aprovado' ou 'rejeitado'"})
	}

	if updateData.GerenteEmail == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email do gerente é obrigatório"})
	}

	if updateData.ComentarioGerente == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comentário do gerente é obrigatório"})
	}

	var request schemas.PontoSolicitacao
	if err := api.DB.DB.First(&request, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Solicitação não encontrada"})
	}

	if request.Status != "pendente" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Solicitação já foi processada"})
	}

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

// Definições de tipos para documentação Swagger
type ManualEditRequest struct {
	EntryTime       string `json:"entry_time"`
	LunchExitTime   string `json:"lunch_exit_time"`
	LunchReturnTime string `json:"lunch_return_time"`
	ExitTime        string `json:"exit_time"`
	MotivoEdicao    string `json:"motivo_edicao" validate:"required"`
	ManagerEmail    string `json:"manager_email" validate:"required,email"`
}

type UpdateRequestStatusRequest struct {
	Status            string `json:"status" validate:"required"`
	ComentarioGerente string `json:"comentario_gerente" validate:"required"`
	GerenteEmail      string `json:"gerente_email" validate:"required,email"`
}
