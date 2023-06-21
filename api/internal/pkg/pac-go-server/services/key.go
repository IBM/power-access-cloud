package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllKeys(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllKeys"})
}

func GetKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetKey"})
}

func CreateKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateKey"})
}

func DeleteKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteKey"})
}
