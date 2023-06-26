package services

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

func GetAllRequests(c *gin.Context) {
	logger := log.GetLogger()
	kc := utils.NewKeyClockClient(c.Request.Context())
	var requests []models.Request
	var err error

	var userID string
	if !kc.IsRole(utils.ManagerRole) {
		// Get authenticated user's ID
		userID = kc.GetUserID()
	}
	listByType := c.DefaultQuery("type", "")
	switch listByType {
	case string(models.RequestAddToGroup), string(models.RequestExtendServiceExpiry), "":
		logger.Debug("getting requests", zap.String("type", listByType), zap.String("user id", userID))
	default:
		logger.Error("Invalid request type", zap.String("type", listByType))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request type - %s", listByType)})
		return
	}
	requests, err = dbCon.GetRequestsByUserID(userID, listByType)
	if err != nil {
		logger.Error("failed to get requests", zap.String("user id", userID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched requests", zap.Any("requests", requests))
	if len(requests) == 0 {
		c.JSON(http.StatusOK, []models.Request{})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func GetRequest(c *gin.Context) {
	logger := log.GetLogger()
	id := c.Param("id")
	logger.Debug("getting requests", zap.String("request id", id))
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched request", zap.Any("request", request))
	c.JSON(http.StatusOK, request)
}

func UpdateServiceExpiryRequest(c *gin.Context) {
	logger := log.GetLogger()
	var request = models.GetRequest()
	userID := c.Request.Context().Value("userid").(string)

	if err := c.BindJSON(&request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("request body", zap.Any("request", request))

	if err := validateCreateRequestParams(request); len(err) > 0 {
		logger.Error("error in create request validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	serviceName := c.Param("name")
	// verify service with name exist
	service, err := kubeClient.GetService(serviceName)
	if err != nil {
		logger.Error("failed to get service", zap.String("service name", serviceName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("fetched service", zap.Any("service", service))

	// service shouldn't be extended if catalog already retired
	catalog, err := kubeClient.GetCatalog(service.Spec.Catalog.Name)
	if err != nil {
		logger.Error("failed to get catalog", zap.String("catalog name", service.Spec.Catalog.Name), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	if catalog.Spec.Retired {
		logger.Debug("catalog is retired, can't extend the expiry", zap.Any("catalog", catalog))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("catalog %s is retired, can't extend the expiry", catalog.Name)})
		return
	}

	// verify already request exist for service extension
	req, err := dbCon.GetRequestByServiceName(serviceName)
	if err != nil {
		logger.Error("failed to fetch the request", zap.String("service name", serviceName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the request from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", req))
	for _, request := range req {
		if request.State == models.RequestStateNew {
			logger.Error("user have already requested to extend service expiry")
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already requested to extend service expiry"})
			return
		}
	}

	// insert the request into the database
	if err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: request.Justification,
		RequestType:   models.RequestExtendServiceExpiry,
		ServiceExpiry: &models.ServiceExpiry{
			Name:   serviceName,
			Expiry: request.ServiceExpiry.Expiry,
		},
	}); err != nil {
		logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to insert the request into the db, err: %s", err.Error())})
		return
	}

	logger.Debug("successfully created request")
	c.Status(http.StatusCreated)
}

func NewGroupRequest(c *gin.Context) {
	logger := log.GetLogger()

	kc := utils.NewKeyClockClient(c.Request.Context())
	var request = models.GetRequest()
	// get the authenticated user's username and ID
	username := c.Request.Context().Value("username").(string)
	userID := c.Request.Context().Value("userid").(string)

	if err := c.BindJSON(&request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("request body", zap.Any("request", request))

	// validate request params
	if err := validateCreateRequestParams(request); len(err) > 0 {
		logger.Error("error in create request validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// check if the user is already a member of the group
	groupID := c.Param("id")
	grp, err := kc.GetGroup(groupID)
	if err != nil && err != utils.ErrorGroupNotFound {
		logger.Error("failed to get groups", zap.String("group id", groupID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if err == utils.ErrorGroupNotFound {
		logger.Error("group not found", zap.String("group id", groupID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	logger.Debug("fetched group", zap.Any("groups", grp))

	if kc.IsMemberOfGroup(grp.Name) {
		logger.Debug("user is already member of group", zap.String("group", grp.Name))
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already a member of this group."})
		return
	}

	// check if the user has already requested access to the group
	r, err := dbCon.GetRequestByGroupIDAndUserID(groupID, userID)
	if err != nil {
		logger.Error("failed to fetch requests", zap.String("group id", groupID), zap.String("user id", userID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", r))
	for _, request := range r {
		if request.State == models.RequestStateNew {
			logger.Debug("user is already requested access to this group", zap.String("group", grp.Name), zap.Any("request", r))
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already requested access to this group."})
			return
		}
	}

	// insert the request into the database
	if err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: request.Justification,
		RequestType:   models.RequestAddToGroup,
		GroupAdmission: &models.GroupAdmission{
			GroupID:   groupID,
			Group:     grp.Name,
			Requester: username,
		},
	}); err != nil {
		logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to insert the request into the db, err: %s", err.Error())})
		return
	}

	logger.Debug("successfully created request")
	c.Status(http.StatusCreated)
}

func ApproveRequest(c *gin.Context) {
	logger := log.GetLogger()
	kc := utils.NewKeyClockClient(c.Request.Context())
	if !kc.IsRole(utils.ManagerRole) {
		logger.Error("only admin can approve the requests")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to approve requests."})
		return
	}

	id := c.Param("id")
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		logger.Error("failed to get requests by id", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", request))
	switch request.RequestType {
	case models.RequestAddToGroup:
		if err := utils.NewKeyClockClient(c.Request.Context()).AddUserToGroup(request.UserID, request.GroupAdmission.GroupID); err != nil {
			logger.Error("failed to add user to group", zap.String("user id", request.UserID),
				zap.String("group id", request.GroupAdmission.GroupID), zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case models.RequestExtendServiceExpiry:
		if err := kubeClient.UpdateServiceExpiry(request.ServiceExpiry.Name, request.ServiceExpiry.Expiry); err != nil {
			logger.Error("failed to update service", zap.String("service name", request.ServiceExpiry.Name), zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
	}
	if err := dbCon.UpdateRequestState(id, models.RequestStateApproved); err != nil {
		logger.Error("failed to update request status in database", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to update the state field in the db, err: %s", err.Error())})
		return
	}
	logger.Debug("successfully approved request", zap.String("id", id))
	c.Status(http.StatusNoContent)
}

func RejectRequest(c *gin.Context) {
	logger := log.GetLogger()
	kc := utils.NewKeyClockClient(c.Request.Context())
	if !kc.IsRole(utils.ManagerRole) {
		logger.Error("only admin can approve the requests")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to reject requests."})
		return
	}

	id := c.Param("id")
	req, err := dbCon.GetRequestByID(id)
	if err != nil {
		logger.Error("failed to get requests by id", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", req))
	if err := dbCon.UpdateRequestState(id, models.RequestStateRejected); err != nil {
		logger.Error("failed to update request status in database", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to update the state field in the db, err: %s", err.Error())})
		return
	}
	logger.Debug("successfully rejected request", zap.String("id", id))
	c.Status(http.StatusNoContent)
}

func DeleteRequest(c *gin.Context) {
	logger := log.GetLogger()
	id := c.Param("id")
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		logger.Error("failed to get requests by id", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", request))
	kc := utils.NewKeyClockClient(c.Request.Context())
	userID := kc.GetUserID()

	// request can be deleted by user who created request or admin can delete
	if userID != request.UserID && !kc.IsRole(utils.ManagerRole) {
		logger.Error("only admin or request creater can delete the requests")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this request."})
		return
	}

	if err := dbCon.DeleteRequest(id); err != nil {
		logger.Error("failed to delete request in database", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to delete the record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("successfully deleted request", zap.String("id", id))
	c.Status(http.StatusNoContent)
}

func validateCreateRequestParams(request models.Request) []error {
	var errs []error
	if len(request.Justification) == 0 {
		errs = append(errs, errors.New("justification should be set"))
	}
	if len(request.Justification) > 500 {
		errs = append(errs, errors.New("justification must be 500 characters or less"))
	}
	if request.RequestType != models.RequestAddToGroup && request.RequestType != models.RequestExtendServiceExpiry {
		errs = append(errs, fmt.Errorf("invalid request_type is set, valid values are %s %s", models.RequestAddToGroup, models.RequestExtendServiceExpiry))
	}

	switch request.RequestType {
	case models.RequestAddToGroup:
	case models.RequestExtendServiceExpiry:
		if request.ServiceExpiry == nil || request.ServiceExpiry.Expiry.IsZero() {
			errs = append(errs, errors.New("expiry time should be set"))
		}
	default:
		errs = append(errs, fmt.Errorf("invalid request_type is set, valid values are %s %s", models.RequestAddToGroup, models.RequestExtendServiceExpiry))
	}
	return errs
}
