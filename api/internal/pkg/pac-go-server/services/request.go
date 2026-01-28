package services

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/client"
	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
)

// GetAllRequests		godoc
// @Summary			Get all requests
// @Description		Get all requests
// @Tags			requests
// @Accept			json
// @Produce			json
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/requests [get]
func GetAllRequests(c *gin.Context) {
	logger := log.GetLogger()
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched requests", zap.Any("requests", requests))
	if len(requests) == 0 {
		c.JSON(http.StatusOK, []models.Request{})
		return
	}
	c.JSON(http.StatusOK, requests)
}

// GetRequest			godoc
// @Summary			Get request
// @Description		Get request
// @Tags			requests
// @Accept			json
// @Produce			json
// @Param			id path string true "request-id for request to be fetched"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/requests/{id} [get]
func GetRequest(c *gin.Context) {
	logger := log.GetLogger()
	id := c.Param("id")
	logger.Debug("getting requests", zap.String("request id", id))
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("request with id: %s not found", id)})
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched request", zap.Any("request", request))
	c.JSON(http.StatusOK, request)
}

// UpdateServiceExpiryRequest			godoc
// @Summary					Update service expiry request
// @Description				Update service expiry for a particular service
// @Tags					requests
// @Accept					json
// @Produce					json
// @Param					name path string true "service name"
// @Param					Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success					200
// @Router					/api/v1/services/{name}/expiry [put]
func UpdateServiceExpiryRequest(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	var expiryRequest = models.GetRequest()
	userID := c.Request.Context().Value("userid").(string)

	if err := utils.BindAndValidate(c, &expiryRequest); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		return
	}
	logger.Debug("request body", zap.Any("request", expiryRequest))

	if err := validateCreateRequestParams(expiryRequest); len(err) > 0 {
		logger.Error("error in create request validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	serviceName := c.Param("name")
	// verify service with name exist
	service, err := kubeClient.GetService(serviceName)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			logger.Error("service not found", zap.String("service name", serviceName))
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("service with name %s not found", serviceName)})
		}
		logger.Error("failed to get service", zap.String("service name", serviceName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("fetched service", zap.Any("service", service))

	// service shouldn't be extended it is already expired
	now := time.Now()
	if now.After(service.Spec.Expiry.Time) {
		logger.Error("service expired", zap.String("service name", serviceName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Service %s is expired, can't extend the expiry", serviceName)})
		return
	}
	// requested expiry should be after current expiry
	currentExpiry := service.Spec.Expiry.Time
	requestedExpiry := expiryRequest.ServiceExpiry.Expiry
	if requestedExpiry.Before(currentExpiry) {
		logger.Error("Requested expiry date should be after the current expiry date")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requested expiry date should be after the current expiry date"})
		return
	}
	// maximum extension allowed is 5 days from current expiry
	if requestedExpiry.After(currentExpiry.AddDate(0, 0, 5)) {
		logger.Error("Maximum extension allowed is 5 days from the current expiry date")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum extension allowed is 5 days from the current expiry date"})
		return
	}

	// service shouldn't be extended if catalog already retired
	catalog, err := kubeClient.GetCatalog(service.Spec.Catalog.Name)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			logger.Error("catalog does not exists", zap.String("catalog name", service.Spec.Catalog.Name))
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		logger.Error("failed to get catalog", zap.String("catalog name", service.Spec.Catalog.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch the request from the db, err: %s", err.Error())})
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
	id, err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: expiryRequest.Justification,
		RequestType:   models.RequestExtendServiceExpiry,
		ServiceExpiry: &models.ServiceExpiry{
			Name:   serviceName,
			Expiry: expiryRequest.ServiceExpiry.Expiry,
		},
	})
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert the request into the db, err: %s", err.Error())})
		return
	}

	event, err := models.NewEvent(userID, originator, models.EventServiceExpiryRequest)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create event, err: %s", err.Error())})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetNotifiyBoth()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("New Request submitted to change the expiry of the service, id: %s", id))

	logger.Debug("successfully created request")
	c.Status(http.StatusCreated)
}

// NewGroupRequest			godoc
// @Summary				New group request
// @Description			Request to switch to new group
// @Tags				requests
// @Accept				json
// @Produce				json
// @Param				id path string true "group-id for the requested group"
// @Param				Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success				200
// @Router				/api/v1/groups/{id}/request [post]
func NewGroupRequest(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()

	var request = models.GetRequest()
	// get the authenticated user's username and ID
	username := c.Request.Context().Value("username").(string)
	userID := c.Request.Context().Value("userid").(string)

	if err := utils.BindAndValidate(c, &request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
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
	grp, err := getGroup(c.Request.Context(), groupID)
	if err != nil && err != client.ErrorGroupNotFound {
		logger.Error("failed to get groups", zap.String("group id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if err == client.ErrorGroupNotFound {
		logger.Error("group not found", zap.String("group id", groupID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched group", zap.Any("groups", grp))

	if models.IsMemberOfGroup(c.Request.Context(), *grp.Name) {
		logger.Debug("user is already member of group", zap.String("group", *grp.Name))
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already a member of this group."})
		return
	}

	// check if the user has already requested access to the group
	r, err := dbCon.GetRequestByGroupIDAndUserID(groupID, userID)
	if err != nil {
		logger.Error("failed to fetch requests", zap.String("group id", groupID), zap.String("user id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	for _, request := range r {
		if request.State == models.RequestStateNew {
			logger.Debug("user is already requested access to this group", zap.String("group", *grp.Name), zap.Any("request", r))
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already requested access to this group."})
			return
		}
	}

	// insert the request into the database
	id, err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: request.Justification,
		RequestType:   models.RequestAddToGroup,
		GroupAdmission: &models.GroupAdmission{
			GroupID:   groupID,
			Group:     *grp.Name,
			Requester: username,
		}})
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert the request into the db, err: %s", err.Error())})
		return
	}

	event, err := models.NewEvent(userID, originator, models.EventGroupJoinRequest)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create event, err: %s", err.Error())})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetNotifiyBoth()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("New Request(%s) has been submitted to add to the group: %s", id, *grp.Name))

	logger.Debug("successfully created request")
	c.Status(http.StatusCreated)
}

// ExitGroup			godoc
// @Summary			Exit group request
// @Description		Request to exit from group
// @Tags			requests
// @Accept			json
// @Produce			json
// @Param			id path string true "group-id for the group to be exited from"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/groups/{id}/exit [post]
func ExitGroup(c *gin.Context) {
	logger := log.GetLogger()

	var request = models.GetRequest()
	// get the authenticated user's username and ID
	username := c.Request.Context().Value("username").(string)
	userID := c.Request.Context().Value("userid").(string)

	if err := utils.BindAndValidate(c, &request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		return
	}
	logger.Debug("request body", zap.Any("request", request))

	// validate request params
	if err := validateCreateRequestParams(request); len(err) > 0 {
		logger.Error("error in exit group request validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// check if the user is already a member of the group
	groupID := c.Param("id")
	grp, err := getGroup(c.Request.Context(), groupID)
	if err != nil && err != client.ErrorGroupNotFound {
		logger.Error("failed to get groups", zap.String("group id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if err == client.ErrorGroupNotFound {
		logger.Error("group not found", zap.String("group id", groupID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched group", zap.Any("groups", grp))

	if !models.IsMemberOfGroup(c.Request.Context(), *grp.Name) {
		logger.Debug("user is not member of group", zap.String("user", username), zap.String("group", *grp.Name))
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already not a member of this group."})
		return
	}

	// check if the user has already requested to exit from the group
	r, err := dbCon.GetRequestByGroupIDAndUserID(groupID, userID)
	if err != nil {
		logger.Error("failed to fetch requests", zap.String("group id", groupID), zap.String("user id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", r))
	for _, request := range r {
		if request.RequestType != models.RequestExitFromGroup {
			continue
		}
		if request.State == models.RequestStateNew {
			logger.Debug("user is already requested to exit form this group", zap.String("group", *grp.Name), zap.Any("request", r))
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already requested to exit from this group."})
			return
		}
	}

	// insert the request into the database
	id, err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: request.Justification,
		RequestType:   models.RequestExitFromGroup,
		GroupAdmission: &models.GroupAdmission{
			GroupID:   groupID,
			Group:     *grp.Name,
			Requester: username,
		},
	})
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert the request into the db, err: %s", err.Error())})
		return
	}

	event, err := models.NewEvent(userID, userID, models.EventGroupExitRequest)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create event, err: %s", err.Error())})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetNotifyAdmin()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Request has been submitted for exiting the group, id: %s", id))

	logger.Debug("successfully created request")
	c.Status(http.StatusCreated)
}

func deleteUserRequest(c *gin.Context) error {
	logger := log.GetLogger()

	var request = models.GetRequest()
	userID := c.Request.Context().Value("userid").(string)

	// TODO : Request details has to be send by UI
	// if err := c.BindJSON(&request); err != nil {
	// 	logger.Error("failed to bind request", zap.Error(err))
	// 	return fmt.Errorf("failed to bind request body, err: %s", err.Error())
	// }

	request.Justification = "User personal information to be deleted"

	logger.Debug("request body", zap.Any("request", request))

	// validate request params
	// TODO : Enable validation once information is coming from UI
	// if err := validateCreateRequestParams(request); len(err) > 0 {
	// 	logger.Error("error in delete user request validation", zap.Errors("errors", err))
	// 	return fmt.Errorf("failed to validate request body, err: %s", err.Error())
	// }

	// insert the request into the database
	id, err := dbCon.NewRequest(&models.Request{
		UserID:        userID,
		CreatedAt:     time.Now(),
		State:         models.RequestStateNew,
		Justification: request.Justification,
		RequestType:   models.RequestDeleteUser,
	})
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return fmt.Errorf("failed to insert the request into the db, err: %s", err.Error())
	}

	event, err := models.NewEvent(userID, userID, models.EventDeletUserRequest)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		return fmt.Errorf("failed to create event, err: %s", err.Error())
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetNotifyAdmin()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Request has been submitted for deleting the user with request-id: %s", id))

	logger.Debug("successfully created request")
	return nil
}

func DeleteUser(c *gin.Context) {
	logger := log.GetLogger()

	// Delete services
	var services []models.Service
	services, err := getAllServices(c)
	if err != nil {
		logger.Error("failed to get all services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	for _, service := range services {
		err := deleteService(c, service.Name)
		if err != nil {
			logger.Error("failed to delete service", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
	}

	// Delete keys
	var keys []models.Key
	keys, err = getAllKeys(c)
	if err != nil {
		logger.Error("failed to get all keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	for _, key := range keys {
		if err := deleteKey(c, key.ID.Hex()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
	}
	// Raise request to delete user
	err = deleteUserRequest(c)
	if err != nil {
		logger.Error("failed to raise request for user deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.Status(http.StatusCreated)
}

func deleteUserFromPreviousGroups(c *gin.Context, request *models.Request) {
	logger := log.GetLogger()
	config := client.GetConfigFromContext(c.Request.Context())
	groups, err := client.NewKeyCloakClient(config, c.Request.Context()).GetUserGroups(request.UserID)
	if err != nil {
		logger.Error("failed to get groups for user", zap.String("user id", request.UserID))
		return
	}
	for _, group := range groups {
		// Not revoking access for recently added group
		if *group.ID == request.GroupAdmission.GroupID {
			continue
		}
		if err := client.NewKeyCloakClient(config, c.Request.Context()).DeleteUserFromGroup(request.UserID, *group.ID); err != nil {
			logger.Error("failed to remove user from group", zap.String("user id", request.UserID),
				zap.String("group id", *group.ID), zap.Error(err))
		}
	}
}

// ApproveRequest		godoc
// @Summary			Approve request
// @Description		Approve request
// @Tags			requests
// @Accept			json
// @Produce			json
// @Param			id path string true "request-id for the request to be approved"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/requests/{id}/approve [post]
func ApproveRequest(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())
	if !kc.IsRole(utils.ManagerRole) {
		logger.Error("only admin can approve the requests")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to approve requests."})
		return
	}

	id := c.Param("id")
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("request with id: %s not found", id)})
		}
		logger.Error("failed to get requests by id", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", request))
	switch request.RequestType {
	case models.RequestAddToGroup:
		if err := client.NewKeyCloakClient(config, c.Request.Context()).AddUserToGroup(request.UserID, request.GroupAdmission.GroupID); err != nil {
			logger.Error("failed to add user to group", zap.String("user id", request.UserID),
				zap.String("group id", request.GroupAdmission.GroupID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		deleteUserFromPreviousGroups(c, request)
	case models.RequestExitFromGroup:
		if err := client.NewKeyCloakClient(config, c.Request.Context()).DeleteUserFromGroup(request.UserID, request.GroupAdmission.GroupID); err != nil {
			logger.Error("failed to remove user from group", zap.String("user id", request.UserID),
				zap.String("group id", request.GroupAdmission.GroupID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case models.RequestDeleteUser:
		if err := client.NewKeyCloakClient(config, c.Request.Context()).DeleteUser(request.UserID); err != nil {
			logger.Error("failed to delete user", zap.String("user id", request.UserID))
			c.JSON(getKeycloakHttpStatus(err), gin.H{"error": err.Error()})
			return
		}
		if err := dbCon.DeleteTermsAndConditionsByUserID(request.UserID); err != nil {
			logger.Error("failed to delete tnc status for user", zap.String("user id", request.UserID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case models.RequestExtendServiceExpiry:
		if err := kubeClient.UpdateServiceExpiry(request.ServiceExpiry.Name, request.ServiceExpiry.Expiry); err != nil {
			if errors.Is(err, utils.ErrResourceNotFound) {
				logger.Error("service does not exists", zap.String("service name", request.ServiceExpiry.Name))
				c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("service with name %s does not exists", request.ServiceExpiry.Name)})
				return
			}
			logger.Error("failed to update service", zap.String("service name", request.ServiceExpiry.Name), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
	}
	if err := dbCon.UpdateRequestState(id, models.RequestStateApproved); err != nil {
		logger.Error("failed to update request status in database", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update the state field in the db, err: %s", err.Error())})
		return
	}
	event, err := models.NewEvent(request.UserID, originator, models.EventTypeRequestApproved)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create event, err: %s", err.Error())})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetNotify()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Request has been successfully approved, id: %s", request.ID.Hex()))
	logger.Debug("successfully approved request", zap.String("id", id))
	c.Status(http.StatusNoContent)
}

// RejectRequest		godoc
// @Summary			Reject request
// @Description		Reject request
// @Tags			requests
// @Accept			json
// @Produce			json
// @Param			id path string true "request-id for the request to be rejected"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/requests/{id}/reject [post]
func RejectRequest(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())
	if !kc.IsRole(utils.ManagerRole) {
		logger.Error("only admin can approve the requests")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to reject requests."})
		return
	}

	var request models.Request
	if err := utils.BindAndValidate(c, &request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		return
	}

	if request.Comment == "" {
		logger.Error("comment is required to reject a request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment is required to reject a request."})
		return
	}

	id := c.Param("id")
	req, err := dbCon.GetRequestByID(id)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("request with id: %s not found", id)})
		}
		logger.Error("failed to get requests by id", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", req))

	if req.State == models.RequestStateRejected || req.State == models.RequestStateApproved {
		logger.Debug("request is already", zap.String("state", string(req.State)), zap.String("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Request is already %s.", string(req.State))})
		return
	}

	if err := dbCon.UpdateRequestStateWithComment(id, models.RequestStateRejected, request.Comment); err != nil {
		logger.Error("failed to update request status in database", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update the state field in the db, err: %s", err.Error())})
		return
	}
	event, err := models.NewEvent(req.UserID, originator, models.EventTypeRequestRejected)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create event, err: %s", err.Error())})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetNotify()
	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Request has been rejected, id: %s", request.ID.Hex()))
	logger.Debug("successfully rejected request", zap.String("id", id))
	c.Status(http.StatusNoContent)
}

// DeleteRequest		godoc
// @Summary			Delete request
// @Description		Delete request
// @Tags			requests
// @Accept			json
// @Produce			json
// @Param			id path string true "request-id for the request to be deleted"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/requests/{id} [delete]
func DeleteRequest(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	id := c.Param("id")
	request, err := dbCon.GetRequestByID(id)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("request with id: %s not found", id)})
		}
		logger.Error("failed to get requests by id", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}
	logger.Debug("fetched request", zap.Any("request", request))
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())
	userID := kc.GetUserID()
	isManagerRole := kc.IsRole(utils.ManagerRole)
	// request can be deleted by user who created request or admin can delete
	if userID != request.UserID && !isManagerRole {
		logger.Error("only admin or request creater can delete the requests")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this request."})
		return
	}

	if err := dbCon.DeleteRequest(id); err != nil {
		logger.Error("failed to delete request in database", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete the record from the db, err: %s", err.Error())})
		return
	}
	event, err := models.NewEvent(request.UserID, originator, models.EventTypeRequestDeleted)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create event, err: %s", err.Error())})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Request has been deleted, id: %s", request.ID.Hex()))
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

	switch request.RequestType {
	case models.RequestAddToGroup:
	case models.RequestExitFromGroup:
	case models.RequestExtendServiceExpiry:
		if request.ServiceExpiry == nil || request.ServiceExpiry.Expiry.IsZero() {
			errs = append(errs, errors.New("expiry time should be set"))
		} else if time.Now().After(request.ServiceExpiry.Expiry) {
			errs = append(errs, errors.New("expiry time should be in future"))
		}
	default:
		errs = append(errs, fmt.Errorf("invalid request_type: \"%s\" is set, valid values are %s %s %s", request.RequestType, models.RequestAddToGroup, models.RequestExitFromGroup, models.RequestExtendServiceExpiry))
	}
	return errs
}
