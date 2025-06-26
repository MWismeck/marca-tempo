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
	
	
	dailyWorkload := workload / 7
	
	
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

	var updateData schemas.TimeLog
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, "Dados inválidos")
	}

	var timeLog schemas.TimeLog
	if err := api.DB.DB.First(&timeLog, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, "Registro não encontrado")
	}

	// Atualiza apenas os campos enviados
	if !updateData.EntryTime.IsZero() {
		timeLog.EntryTime = updateData.EntryTime
	}
	if !updateData.LunchExitTime.IsZero() {
		timeLog.LunchExitTime = updateData.LunchExitTime
	}
	if !updateData.LunchReturnTime.IsZero() {
		timeLog.LunchReturnTime = updateData.LunchReturnTime
	}
	if !updateData.ExitTime.IsZero() {
		timeLog.ExitTime = updateData.ExitTime
	}

	// Marca edição do gerente
	timeLog.EditadoPorGerente = updateData.EditadoPorGerente
	timeLog.EditadoEm = time.Now()

	if err := api.DB.DB.Save(&timeLog).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, "Erro ao salvar")
	}

	return c.JSON(http.StatusOK, timeLog)
}

func (api *API) requestTimeEdit(c echo.Context) error {
	var req schemas.PontoSolicitacao
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	if req.FuncionarioEmail == "" || req.Motivo == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email e motivo são obrigatórios"})
	}

	if err := api.DB.DB.Create(&req).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar solicitação"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Solicitação registrada com sucesso"})
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
	f.SetActiveSheet(f.GetSheetIndex(sheet))

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


