package services

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

func GetAllGroups(c *gin.Context) {

	var groups []models.Group

	groups, err := utils.NewKeyClockClient(c.Request.Context()).GetGroups()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, group := range groups {
		groups[i].Membership = utils.NewKeyClockClient(c.Request.Context()).IsMemberOfGroup(group.Name)
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroup(c *gin.Context) {
	id := c.Param("id")
	grp, err := utils.NewKeyClockClient(c.Request.Context()).GetGroup(id)
	if err != nil && err != utils.ErrorGroupNotFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if err == utils.ErrorGroupNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, grp)
}
