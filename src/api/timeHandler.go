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

    
    currentDate := time.Now().Truncate(24 * time.Hour)

    
    var timeLog schemas.TimeLog
    if err := api.DB.DB.Where("employee_email = ? AND log_date = ?", employeeEmail, currentDate).First(&timeLog).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            
            timeLog = schemas.TimeLog{
                EmployeeEmail: employeeEmail,
                LogDate:       currentDate,
                EntryTime:     time.Now(),
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

    
    now := time.Now()
    if timeLog.EntryTime.IsZero() {
        timeLog.EntryTime = now
    } else if timeLog.LunchExitTime.IsZero() {
        timeLog.LunchExitTime = now
    } else if timeLog.LunchReturnTime.IsZero() {
        timeLog.LunchReturnTime = now
    } else if timeLog.ExitTime.IsZero() {
        timeLog.ExitTime = now
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




func calculateHours(entryTime, lunchExitTime, lunchReturnTime, exitTime time.Time, workload float32) (extraHours, missingHours, balance float32) {
	
	workedDuration := exitTime.Sub(entryTime) - (lunchReturnTime.Sub(lunchExitTime))

	
	workedHours := float32(workedDuration.Hours())

	
	if workedHours > workload {
		extraHours = workedHours - workload
		missingHours = 0
	} else {
		extraHours = 0
		missingHours = workload - workedHours
	}

	
	balance = extraHours - missingHours
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

    // Apply some styling
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
