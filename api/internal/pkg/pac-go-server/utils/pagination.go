package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCurrentPageAndPageCount(c *gin.Context) (pageInt, perPageInt int64) {
	page := c.DefaultQuery("page", "1")         // Get the page number from the query parameter
	perPage := c.DefaultQuery("per_page", "10") // Get the number of items per page from the query parameter

	// Convert the page and perPage values to integers
	pageInt, _ = strconv.ParseInt(page, 10, 64)
	perPageInt, _ = strconv.ParseInt(perPage, 10, 64)

	return pageInt, perPageInt
}

func GetTotalPages(totalCount, perPage int64) int64 {
	totalPages := totalCount / perPage
	if totalCount%perPage != 0 {
		totalPages++
	}
	return totalPages
}
