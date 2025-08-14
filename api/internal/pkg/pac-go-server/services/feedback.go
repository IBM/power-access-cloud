package services

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateFeedback		godoc
// @Summary			Create Feedback given by user
// @Description		Create feedback resource
// @Tags			feedbacks
// @Accept			json
// @Produce			json
// @Param			feedback body models.Feedback true "Create feedback"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/feedbacks [post]
func CreateFeedback(c *gin.Context) {
	var feedback *models.Feedback
	if err := c.BindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.Request.Context().Value("userid").(string)
	feedback.UserID = userID
	feedback.CreatedAt = time.Now()
	logger := log.GetLogger()
	if err := feedback.ValidateFeedback(); len(err) > 0 {
		logger.Error("error while validating feedback request", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	event, err := models.NewEvent(userID, userID, models.EventFeedbackCreate)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create event, err: %s", err.Error())})
		return
	}
	if err := dbCon.InsertFeedback(feedback); err != nil {
		logger.Error("failed to insert user feedback into the database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to insert the feedback into the database, Error: %s", err.Error())})
		return
	}
	logger.Info("Successfully stored feedback given by user")
	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Feedback has been submitted by the user : %s", userID))
	event.SetNotifyAdmin()
	c.Status(http.StatusCreated)
}

// GetFeedback		godoc
// @Summary			Get feedbacks submitted by users
// @Description		Get feedback resource
// @Tags			feedbacks
// @Accept			json
// @Produce			json
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param        	page     query     int     false  "Page number for pagination"
// @Param        	per_page    query     int     false  "Number of items per page"
// @Success			200
// @Router			/api/v1/feedbacks [get]
func GetFeedback(c *gin.Context) {
	logger := log.GetLogger()
	pageInt, perPageInt := utils.GetCurrentPageAndPageCount(c)
	startIndex := (pageInt - 1) * perPageInt
	feedback, totalCount, err := dbCon.GetFeedbacks(models.FeedbacksFilter{}, startIndex, perPageInt)
	if err != nil {
		logger.Error("failed to get feedbacks from db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	// Calculate the total number of pages based on the perPage value
	totalPages := utils.GetTotalPages(totalCount, perPageInt)
	var response = models.FeedbackResponse{
		TotalPages: totalPages,
		TotalItems: totalCount,
		Feedbacks:  feedback,
		Links: models.Links{
			Self: c.Request.URL.String(),
			Next: getNextPageLink(c, pageInt, totalPages),
			Last: getLastPageLink(c, totalPages),
		},
	}
	c.JSON(http.StatusOK, response)
}
