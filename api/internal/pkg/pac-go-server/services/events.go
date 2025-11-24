package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/client"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
)

// GetEvents			godoc
// @Summary			Get events
// @Description		Get events
// @Tags			events
// @Accept			json
// @Produce			json
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/events [get]
// GetEvents returns all events
func GetEvents(c *gin.Context) {
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())

	pageInt, perPageInt := utils.GetCurrentPageAndPageCount(c)
	// Calculate the starting index and ending index for the current page
	startIndex := (pageInt - 1) * perPageInt

	var userID string
	if !kc.IsRole(utils.ManagerRole) {
		// Get authenticated user's ID
		userID = kc.GetUserID()
	}
	events, totalCount, err := dbCon.GetEventsByUserID(userID, startIndex, perPageInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate the total number of pages based on the perPage value
	totalPages := utils.GetTotalPages(totalCount, perPageInt)

	var response = models.EventResponse{
		TotalPages: totalPages,
		TotalItems: totalCount,
		Events:     events,
		Links: models.Links{
			Self: c.Request.URL.String(),
			Next: getNextPageLink(c, pageInt, totalPages),
			Last: getLastPageLink(c, totalPages),
		},
	}

	c.JSON(http.StatusOK, response)
}

func getNextPageLink(c *gin.Context, currentPage, totalPages int64) string {
	if currentPage >= totalPages {
		return ""
	}
	nextPage := currentPage + 1
	return getPaginationLink(c, nextPage)
}

func getLastPageLink(c *gin.Context, totalPages int64) string {
	if totalPages <= 1 {
		return ""
	}
	return getPaginationLink(c, totalPages)
}

func getPaginationLink(c *gin.Context, page int64) string {
	queryParams := c.Request.URL.Query()
	queryParams.Set("page", strconv.FormatInt(page, 10))
	return c.Request.URL.Path + "?" + queryParams.Encode()
}
