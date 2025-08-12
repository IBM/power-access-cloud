package services

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

// GetAllGroups			godoc
// @Summary				Get all groups
// @Description			Get all groups
// @Tags				groups
// @Accept				json
// @Produce				json
// @Param				Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success				200
// @Router				/api/v1/groups [get]
func GetAllGroups(c *gin.Context) {
	logger := log.GetLogger()
	var err error
	groups := []models.Group{}
	grps, err := getGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var groupIds []string
	for _, group := range grps {
		groups = append(groups, models.Group{
			Name:       *group.Name,
			ID:         *group.ID,
			Membership: models.IsMemberOfGroup(c.Request.Context(), *group.Name),
		})
		groupIds = append(groupIds, *group.ID)
	}

	logger.Debug("fetching groups quota")
	groupsQuota, err := dbCon.GetGroupsQuota(groupIds)
	if err != nil {
		logger.Error("failed to get groups quota", zap.Any("group ids", groupIds), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Debug("fetched groups quota", zap.Any("groups quota", groupsQuota))

	// construct a map with quota associated for group
	quotaMap := make(map[string]models.Capacity)
	for _, groupQuota := range groupsQuota {
		quotaMap[groupQuota.GroupID] = groupQuota.Capacity
	}

	// update groups quota
	for index, group := range groups {
		if quota, ok := quotaMap[group.ID]; ok {
			groups[index].Quota = quota
		}
	}
	logger.Debug("groups with quota", zap.Any("groups", groups))
	c.JSON(http.StatusOK, groups)
}

// GetGroup				godoc
// @Summary				Get group
// @Description			Get group as specified in request
// @Tags				groups
// @Accept				json
// @Produce				json
// @Param				id path string true "group-id to be fetched"
// @Param				Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success				200
// @Router				/api/v1/groups/{id} [get]
func GetGroup(c *gin.Context) {
	id := c.Param("id")
	grp, err := getGroup(c.Request.Context(), id)
	if err != nil && err != client.ErrorGroupNotFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if err == client.ErrorGroupNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	quota, err := dbCon.GetQuotaForGroupID(id)
	if err != nil {
		if errors.Unwrap(err) != mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// if quota is not found, set quota to 0
		quota = &models.Quota{}
	}
	group := models.Group{
		Name:       *grp.Name,
		ID:         *grp.ID,
		Membership: models.IsMemberOfGroup(c.Request.Context(), *grp.Name),
		Quota:      quota.Capacity,
	}
	c.JSON(http.StatusOK, group)
}
