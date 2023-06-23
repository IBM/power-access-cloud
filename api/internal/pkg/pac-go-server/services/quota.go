package services

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

// Get the respective quota of the group ID passed.
func GetQuota(c *gin.Context) {
	logger := log.GetLogger()
	gid := c.Param("id")
	if err := checkGroupExists(c, gid); err != nil {
		logger.Error("Cannot find group by ID", zap.String("id", gid), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("The group ID - %s does not exist.", gid)})
		return
	}
	quotaDb, err := dbCon.GetQuotaForGroupID(gid)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("GetQuota : Error occured while checking quota", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "An error occured while retriving quota, contact PAC support.", "error": err.Error()})
		return
	}
	if quotaDb == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "A quota policy does not exist for this group ID. You need to create one first."})
		return
	}
	c.JSON(http.StatusOK, &quotaDb)
}

func CreateQuota(c *gin.Context) {
	var quota models.Quota
	logger := log.GetLogger()
	gid := c.Param("id")

	if err := checkGroupExists(c, gid); err != nil {
		logger.Error("Cannot find group by ID", zap.String("id", gid))
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("The group ID - %s does not exist.", gid)})
		return
	}

	if err := c.BindJSON(&quota); err != nil {
		logger.Error("Create quota - error while creating quota", zap.String("groupID", gid), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if quota.GroupID != "" && quota.GroupID != gid {
		logger.Error("GroupID must not be set in the request body, or must match the one set in request path")
		c.JSON(http.StatusBadRequest, gin.H{"message": "GroupID must not be set in the request body, or must match the one set in request path"})
		return
	}

	if err := utils.ValidateQuotaFields(c, quota.Capacity.CPU, quota.Capacity.Memory); err != nil {
		logger.Error("Quota validation has failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// TODO: check if the quota for the particular group first exists, before updating
	quotaDb, err := dbCon.GetQuotaForGroupID(gid)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("CreateQuota : Error occured while checking quota", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "An error occured while creating Quota, contact PAC support."})
		return
	}
	if quotaDb != nil {
		c.JSON(http.StatusConflict, gin.H{"message": "A quota policy already exists for this group ID. You may delete or update the existing quota."})
		return
	}
	if err := dbCon.NewQuota(&models.Quota{
		GroupID:  gid,
		Capacity: quota.Capacity,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to insert the quota into the database, Error: %s", err.Error())})
	}

	logger.Info("Created quota successfully", zap.String("groupID", gid), zap.Any("Capacity", quota.Capacity))
	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Created resource quota successfully for groupID %s", quota.GroupID)})
}

func UpdateQuota(c *gin.Context) {
	var quota models.Quota
	logger := log.GetLogger()
	gid := c.Param("id")

	if err := checkGroupExists(c, gid); err != nil {
		logger.Error("Cannot find group by ID", zap.String("id", gid))
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("The group ID - %s does not exist.", gid)})
		return
	}

	logger.Info("Update quota", zap.String("groupID", gid))

	if err := c.BindJSON(&quota); err != nil {
		logger.Error("Create quota - error while creating quota", zap.String("groupID", gid), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if quota.GroupID != "" && quota.GroupID != gid {
		logger.Error("GroupID must not be set in the request body, or must match the one set in request path")
		c.JSON(http.StatusBadRequest, gin.H{"message": "GroupID must not be set in the request body, or must match the one set in request path"})
		return
	}

	if err := utils.ValidateQuotaFields(c, quota.Capacity.CPU, quota.Capacity.Memory); err != nil {
		logger.Error("Quota validation has failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// TODO: check if the quota for the particular group first exists, before updating
	quotaDb, err := dbCon.GetQuotaForGroupID(gid)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("CreateQuota : Error occured while checking quota", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "An error occured while creating Quota, contact PAC support."})
		return
	}
	if quotaDb == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "A quota policy does not exist for this group ID. You need to create one first."})
		return
	}
	if err := dbCon.UpdateQuota(&models.Quota{
		GroupID:  gid,
		Capacity: quota.Capacity,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to insert the quota into the database, Error: %s", err.Error())})
		return
	}

	logger.Info("Updated quota successfully", zap.String("groupID", gid), zap.Any("Capacity", quota.Capacity))
	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Updated resource quota successfully for groupID %s", quota.GroupID)})
}

func DeleteQuota(c *gin.Context) {
	logger := log.GetLogger()

	gid := c.Param("id")

	// Check if the group ID is valid, else return 404
	if err := checkGroupExists(c, gid); err != nil {
		logger.Error("Cannot find group by ID", zap.String("id", gid))
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("The group ID - %s does not exist.", gid)})
		return
	}

	if err := dbCon.DeleteQuota(gid); err != nil {
		logger.Error("Quota could not be deleted", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error while deleting quota"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted resource quota successfully"})
}

// checkGroupExists : Check if the group exists first before
func checkGroupExists(c *gin.Context, gid string) error {
	logger := log.GetLogger()
	_, err := utils.NewKeyClockClient(c.Request.Context()).GetGroup(gid)
	if err != nil && err != utils.ErrorGroupNotFound {
		logger.Info("Error while retriving groupID from KeyCloak", zap.String("id", gid), zap.Error(err))
		return err
	} else if err == utils.ErrorGroupNotFound {
		logger.Error("No group exists in keycloak", zap.String("id", gid))
		return err
	}
	return nil
}

func GetUserQuota(c *gin.Context) {
	logger := log.GetLogger()
	var userQuota, usedQuota, availableQuota models.Capacity
	var err error
	kc := utils.NewKeyClockClient(c.Request.Context())
	userID := kc.GetUserID()
	logger.Debug("getting quota for user", zap.String("user id", userID))

	userGroups := kc.GetUserGroups()
	logger.Debug("user groups", zap.Any("user groups", userGroups))

	if len(userGroups) != 0 {
		logger.Debug("fetching user quota for groups")
		var userGroupIds []string
		for _, grp := range userGroups {
			userGroupIds = append(userGroupIds, grp.ID)
		}
		groupsQuota, err := dbCon.GetGroupsQuota(userGroupIds)
		if err != nil {
			logger.Error("failed to get quota", zap.String("user id", userID), zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get quota, err: %s", err.Error())})
			return
		}
		logger.Debug("user group quota", zap.String("user id", userID), zap.Any("group quota", groupsQuota))
		userQuota = getMaxCapacity(groupsQuota)
		logger.Debug("user maximum quota", zap.Any("user quota", userQuota))
	} else {
		logger.Debug("user does not belong to any group", zap.String("user id", userID))
	}
	usedQuota, err = usedCapacity(userID)
	if err != nil {
		logger.Error("failed to get used quota", zap.String("userid", userID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get used quota %v", err)})
		return
	}
	availableQuota.CPU = userQuota.CPU - usedQuota.CPU
	availableQuota.Memory = userQuota.Memory - usedQuota.Memory
	
	// in case of negative available quota set it to 0
	if availableQuota.CPU < 0 {
		availableQuota.CPU = 0
	}
	if availableQuota.Memory < 0 {
		availableQuota.Memory = 0
	}
	logger.Debug("quotas of user", zap.Any("used quota", usedQuota), zap.Any("available quota", availableQuota))
	c.JSON(http.StatusOK, gin.H{"user_quota": userQuota, "used_quota": usedQuota, "available_quota": availableQuota})
}

func getMaxCapacity(quotas []models.Quota) models.Capacity {
	var maxCPU float64
	var maxMemory int

	for _, quota := range quotas {
		if quota.Capacity.CPU > maxCPU {
			maxCPU = quota.Capacity.CPU
		}
		if quota.Capacity.Memory > maxMemory {
			maxMemory = quota.Capacity.Memory
		}
	}
	return models.Capacity{
		CPU:    maxCPU,
		Memory: maxMemory,
	}
}
