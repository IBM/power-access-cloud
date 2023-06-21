package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllQuotas"})
}

func GetQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetQuota"})
}

func CreateQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateQuota"})
}

func UpdateQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateQuota"})
}

func DeleteQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteQuota"})
}
