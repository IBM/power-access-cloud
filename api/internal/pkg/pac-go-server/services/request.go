package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

// TODO: Update accroding to new request spec modification
func GetAllRequests(c *gin.Context) {
	kc := utils.NewKeyClockClient(c.Request.Context())
	var requests []models.Request

	var userID string
	if !kc.IsRole(utils.ManagerRole) {
		// Get authenticated user's ID
		userID = kc.GetUserID()
	}

	requests, err := dbCon.GetRequestsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusOK, struct{}{})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func GetRequest(c *gin.Context) {
	id := c.Param("id")

	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// TODO: Update accroding to new request spec modification
func NewRequest(c *gin.Context) {
	kc := utils.NewKeyClockClient(c.Request.Context())
	var request = models.GetRequest()
	// Step0: Get the authenticated user's username and ID
	username := c.Request.Context().Value("username").(string)
	userID := c.Request.Context().Value("userid").(string)

	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Step1: Validate the request
	if len(request.Justification) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Justification must be 500 characters or less."})
		return
	}

	// Step2: Check if the user is already a member of the group
	groupID := c.Param("id")
	grp, err := kc.GetGroup(groupID)
	if err != nil && err != utils.ErrorGroupNotFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if err == utils.ErrorGroupNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	if kc.IsMemberOfGroup(grp.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already a member of this group."})
		return
	}

	// Step3: Check if the user has already requested access to the group
	r, err := dbCon.GetRequestByGroupIDAndUserID(groupID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	if r != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already requested access to this group."})
		return
	}

	// Step4: Insert the request into the database
	if err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: request.Justification,
		GroupAdmission: models.GroupAdmission{
			GroupID:   groupID,
			Group:     grp.Name,
			Requester: username,
		},
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to insert the request into the db, err: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, request)
}

// TODO: Update accroding to new request spec modification
func ApproveRequest(c *gin.Context) {
	kc := utils.NewKeyClockClient(c.Request.Context())
	if !kc.IsRole(utils.ManagerRole) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to approve requests."})
		return
	}

	id := c.Param("id")
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}

	if err := utils.NewKeyClockClient(c.Request.Context()).AddUserToGroup(request.UserID, request.GroupAdmission.GroupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := dbCon.UpdateRequestState(id, models.RequestStateApproved); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to update the state field in the db, err: %s", err.Error())})
		return
	}

	c.Status(http.StatusNoContent)
}

// TODO: Update accroding to new request spec modification
func RejectRequest(c *gin.Context) {
	kc := utils.NewKeyClockClient(c.Request.Context())
	if !kc.IsRole(utils.ManagerRole) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to reject requests."})
		return
	}

	id := c.Param("id")
	_, err := dbCon.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}

	if err := dbCon.UpdateRequestState(id, models.RequestStateRejected); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to update the state field in the db, err: %s", err.Error())})
		return
	}

	c.Status(http.StatusNoContent)
}

// TODO: Update accroding to new request spec modification
func DeleteRequest(c *gin.Context) {
	id := c.Param("id")
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}

	if request.UserID != c.Request.Context().Value("userid").(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this request."})
		return
	}

	if err := dbCon.DeleteRequest(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to delete the record from the db, err: %s", err.Error())})
		return
	}

	c.Status(http.StatusNoContent)
}
