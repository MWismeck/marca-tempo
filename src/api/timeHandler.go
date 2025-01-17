package api

import (
	"github.com/MWismeck/marca-tempo/src/schemas"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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



