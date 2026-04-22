package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/client"
	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

const (
	earlyWarningHours = 24 * time.Hour
	timeDisplayFormat = "Jan 02, 2006 03:04 PM"
)

// GetMaintenanceWindows godoc
// @Summary			Get maintenance notification windows
// @Description		Get maintenance windows. Use ?all=true (admin only) to get all windows with audit details and pagination, otherwise returns only the earliest active window
// @Tags			maintenance
// @Accept			json
// @Produce			json
// @Param			all query bool false "Get all windows (admin only)"
// @Param			page query int false "Page number for pagination (admin only)"
// @Param			per_page query int false "Number of items per page (admin only)"
// @Success			200 {object} models.MaintenanceListResponse
// @Router			/api/v1/maintenance [get]
func GetMaintenanceWindows(c *gin.Context) {
	logger := log.GetLogger()

	// Check if admin wants all windows
	showAll := c.Query("all") == "true"

	windows, err := dbCon.GetAllMaintenanceWindows()
	if err != nil {
		logger.Error("failed to get maintenance windows from db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve maintenance windows"})
		return
	}

	now := time.Now()

	// Admin view: return all windows with audit details and pagination
	if showAll {
		adminResponses := buildAdminResponses(c, windows, now)
		if adminResponses == nil {
			return // Error already handled in buildAdminResponses
		}

		// Apply pagination
		pageInt, perPageInt := utils.GetCurrentPageAndPageCount(c)
		startIndex := int((pageInt - 1) * perPageInt)
		endIndex := int(pageInt * perPageInt)

		totalCount := int64(len(adminResponses))

		// Handle pagination bounds
		if startIndex >= len(adminResponses) {
			startIndex = len(adminResponses)
		}
		if endIndex > len(adminResponses) {
			endIndex = len(adminResponses)
		}

		paginatedResponses := adminResponses[startIndex:endIndex]
		totalPages := utils.GetTotalPages(totalCount, perPageInt)

		c.JSON(http.StatusOK, models.MaintenanceAdminListResponse{
			TotalPages:   totalPages,
			TotalItems:   totalCount,
			Maintenances: paginatedResponses,
			Links: models.Links{
				Self: c.Request.URL.String(),
				Next: getNextPageLink(c, pageInt, totalPages),
				Last: getLastPageLink(c, totalPages),
			},
		})
		return
	}

	// User view: return only the earliest active window (no pagination needed)
	responses := buildUserResponses(windows, now)
	c.JSON(http.StatusOK, models.MaintenanceListResponse{
		Maintenances: responses,
	})
}

// buildAdminResponses creates admin response with all windows and audit details
func buildAdminResponses(c *gin.Context, windows []models.MaintenanceWindow, now time.Time) []models.MaintenanceAdminResponse {
	// Verify admin access
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())

	if !kc.IsRole(utils.ManagerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return nil
	}

	adminResponses := make([]models.MaintenanceAdminResponse, 0, len(windows))
	for _, window := range windows {
		adminResponses = append(adminResponses, models.MaintenanceAdminResponse{
			ID:        window.ID.Hex(),
			Enabled:   window.Enabled,
			StartTime: window.StartTime,
			EndTime:   window.EndTime,
			Message:   window.Message,
			CreatedBy: window.CreatedBy,
			CreatedAt: window.CreatedAt,
			UpdatedBy: window.UpdatedBy,
			UpdatedAt: window.UpdatedAt,
			DeletedBy: window.DeletedBy,
			DeletedAt: window.DeletedAt,
			IsActive:  isWindowActive(window, now),
		})
	}

	return adminResponses
}

// buildUserResponses creates user response with only the earliest active window
func buildUserResponses(windows []models.MaintenanceWindow, now time.Time) []models.MaintenanceResponse {
	var earliestActiveWindow *models.MaintenanceWindow

	for i := range windows {
		window := &windows[i]

		// Skip if not enabled or not active
		if !window.Enabled || !isWindowActive(*window, now) {
			continue
		}

		// Keep only the earliest active window (by start time)
		if earliestActiveWindow == nil || window.StartTime.Before(earliestActiveWindow.StartTime) {
			earliestActiveWindow = window
		}
	}

	// Return only the earliest active window (or empty array if none)
	if earliestActiveWindow == nil {
		return []models.MaintenanceResponse{}
	}

	return []models.MaintenanceResponse{{
		ID:        earliestActiveWindow.ID.Hex(),
		Enabled:   earliestActiveWindow.Enabled,
		StartTime: earliestActiveWindow.StartTime,
		EndTime:   earliestActiveWindow.EndTime,
		Message:   earliestActiveWindow.Message,
		IsActive:  true,
	}}
}

// isWindowActive checks if a maintenance window is currently active
func isWindowActive(window models.MaintenanceWindow, now time.Time) bool {
	earlyWarningTime := window.StartTime.Add(-earlyWarningHours)
	return window.Enabled && now.After(earlyWarningTime) && now.Before(window.EndTime)
}

// checkForOverlaps checks if the given time range overlaps with any existing enabled windows
// Returns error if overlap is found
func checkForOverlaps(startTime, endTime time.Time, excludeID string) error {
	windows, err := dbCon.GetAllMaintenanceWindows()
	if err != nil {
		return err
	}

	for _, window := range windows {
		// Skip the window being updated or disabled windows
		if (excludeID != "" && window.ID.Hex() == excludeID) || !window.Enabled {
			continue
		}

		// Check for overlap: startTime < window.EndTime AND endTime > window.StartTime
		if startTime.Before(window.EndTime) && endTime.After(window.StartTime) {
			return formatOverlapError(window.StartTime, window.EndTime)
		}
	}

	return nil
}

// formatOverlapError creates a user-friendly overlap error message
func formatOverlapError(existingStart, existingEnd time.Time) error {
	localStart := existingStart.In(time.Local).Format(timeDisplayFormat)
	localEnd := existingEnd.In(time.Local).Format(timeDisplayFormat)

	return fmt.Errorf("Start time overlaps with existing maintenance window (%s to %s). Please update the existing window's end time instead of creating a new one",
		localStart, localEnd)
}

// CreateMaintenanceWindow godoc
// @Summary			Create a new maintenance notification window
// @Description		Create a new maintenance notification window (admin only)
// @Tags			maintenance
// @Accept			json
// @Produce			json
// @Param			maintenance body models.MaintenanceCreateRequest true "Maintenance window configuration"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			201 {object} models.MaintenanceResponse
// @Router			/api/v1/maintenance [post]
func CreateMaintenanceWindow(c *gin.Context) {
	logger := log.GetLogger()

	var request models.MaintenanceCreateRequest
	if err := utils.BindAndValidate(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from Keycloak client
	userID := getUserID(c)
	now := time.Now()

	// Validate request
	if err := validateTimeRange(request.StartTime, request.EndTime, now, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check for overlapping windows
	if err := checkForOverlaps(request.StartTime, request.EndTime, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create maintenance window
	window := &models.MaintenanceWindow{
		ID:        primitive.NewObjectID(),
		Enabled:   request.Enabled,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
		Message:   request.Message,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedBy: userID,
		UpdatedAt: now,
	}

	if err := dbCon.CreateMaintenanceWindow(window); err != nil {
		logger.Error("failed to create maintenance window in db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance window"})
		return
	}

	logger.Info("maintenance window created",
		zap.String("id", window.ID.Hex()),
		zap.String("created_by", userID),
		zap.Bool("enabled", window.Enabled))

	logMaintenanceEvent(logger, userID, window.ID.Hex(), "maintenance_window_create", window.Enabled)

	c.JSON(http.StatusCreated, buildMaintenanceResponse(window))
}

// UpdateMaintenanceWindow godoc
// @Summary			Update a maintenance notification window
// @Description		Update an existing maintenance notification window (admin only)
// @Tags			maintenance
// @Accept			json
// @Produce			json
// @Param			id path string true "Maintenance Window ID"
// @Param			maintenance body models.MaintenanceUpdateRequest true "Maintenance window configuration"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200 {object} models.MaintenanceResponse
// @Router			/api/v1/maintenance/{id} [put]
func UpdateMaintenanceWindow(c *gin.Context) {
	logger := log.GetLogger()
	id := c.Param("id")

	var request models.MaintenanceUpdateRequest
	if err := utils.BindAndValidate(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := getUserID(c)
	now := time.Now()

	// Get existing window
	existingWindow, err := dbCon.GetMaintenanceWindowByID(id)
	if err != nil {
		logger.Error("failed to get maintenance window from db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve maintenance window"})
		return
	}

	if existingWindow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance window not found"})
		return
	}

	// Apply updates and validate
	if err := applyWindowUpdates(existingWindow, &request, userID, now, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := dbCon.UpdateMaintenanceWindow(existingWindow); err != nil {
		logger.Error("failed to update maintenance window in db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update maintenance window"})
		return
	}

	logger.Info("maintenance window updated",
		zap.String("id", existingWindow.ID.Hex()),
		zap.String("updated_by", userID),
		zap.Bool("enabled", existingWindow.Enabled))

	logMaintenanceEvent(logger, userID, existingWindow.ID.Hex(), "maintenance_window_update", existingWindow.Enabled)

	c.JSON(http.StatusOK, buildMaintenanceResponse(existingWindow))
}

// applyWindowUpdates applies the update request to the existing window and validates
func applyWindowUpdates(window *models.MaintenanceWindow, request *models.MaintenanceUpdateRequest, userID string, now time.Time, windowID string) error {
	// Determine final times
	finalStartTime := window.StartTime
	finalEndTime := window.EndTime

	if !request.StartTime.IsZero() {
		finalStartTime = request.StartTime
	}
	if !request.EndTime.IsZero() {
		finalEndTime = request.EndTime
	}

	// Validate new start time if changed
	if !request.StartTime.IsZero() && !request.StartTime.Equal(window.StartTime) {
		if request.StartTime.Before(now) {
			return fmt.Errorf("start_time cannot be in the past")
		}
	}

	// Validate time range
	if err := validateTimeRange(finalStartTime, finalEndTime, now, false); err != nil {
		return err
	}

	// Check for overlaps
	if err := checkForOverlaps(finalStartTime, finalEndTime, windowID); err != nil {
		return err
	}

	// Apply updates
	window.Enabled = request.Enabled
	if !request.StartTime.IsZero() {
		window.StartTime = request.StartTime
	}
	if !request.EndTime.IsZero() {
		window.EndTime = request.EndTime
	}
	if request.Message != "" {
		window.Message = request.Message
	}
	window.UpdatedBy = userID
	window.UpdatedAt = now

	return nil
}

// DeleteMaintenanceWindow godoc
// @Summary			Delete a maintenance notification window
// @Description		Delete a maintenance notification window (admin only)
// @Tags			maintenance
// @Accept			json
// @Produce			json
// @Param			id path string true "Maintenance Window ID"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			204 "No Content"
// @Router			/api/v1/maintenance/{id} [delete]
func DeleteMaintenanceWindow(c *gin.Context) {
	logger := log.GetLogger()
	id := c.Param("id")
	userID := getUserID(c)

	// Check if window exists
	window, err := dbCon.GetMaintenanceWindowByID(id)
	if err != nil {
		logger.Error("failed to get maintenance window from db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve maintenance window"})
		return
	}

	if window == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance window not found"})
		return
	}

	// Perform soft delete
	now := time.Now()
	if err := dbCon.DeleteMaintenanceWindow(id, userID, &now); err != nil {
		logger.Error("failed to soft delete maintenance window from db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete maintenance window"})
		return
	}

	logger.Info("maintenance window soft deleted",
		zap.String("id", id),
		zap.String("deleted_by", userID))

	logMaintenanceEvent(logger, userID, id, "maintenance_window_delete", false)

	c.Status(http.StatusNoContent)
}

// Helper functions

// getUserID extracts user ID from Keycloak context
func getUserID(c *gin.Context) string {
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())
	return kc.GetUserID()
}

// validateTimeRange validates that end time is after start time and optionally checks if start is in future
func validateTimeRange(startTime, endTime, now time.Time, checkFuture bool) error {
	if checkFuture && startTime.Before(now) {
		return fmt.Errorf("start_time cannot be in the past")
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	return nil
}

// buildMaintenanceResponse creates a public maintenance response
func buildMaintenanceResponse(window *models.MaintenanceWindow) models.MaintenanceResponse {
	return models.MaintenanceResponse{
		ID:        window.ID.Hex(),
		Enabled:   window.Enabled,
		StartTime: window.StartTime,
		EndTime:   window.EndTime,
		Message:   window.Message,
	}
}

// logMaintenanceEvent logs a maintenance event to the database
func logMaintenanceEvent(logger *zap.Logger, userID, windowID, eventType string, enabled bool) {
	event, err := models.NewEvent(userID, userID, models.EventType(eventType))
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			logger.Error("failed to create event", zap.Error(err))
		}
	}()

	var eventLog string
	switch eventType {
	case "maintenance_window_create":
		eventLog = fmt.Sprintf("Maintenance window created by admin %s (ID: %s, Enabled: %v)", userID, windowID, enabled)
	case "maintenance_window_update":
		eventLog = fmt.Sprintf("Maintenance window updated by admin %s (ID: %s, Enabled: %v)", userID, windowID, enabled)
	case "maintenance_window_delete":
		eventLog = fmt.Sprintf("Maintenance window deleted by admin %s (ID: %s)", userID, windowID)
	}

	event.SetLog(models.EventLogLevelINFO, eventLog)
}
