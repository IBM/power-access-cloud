package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllCatalogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllCatalogs"})
}

func GetCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetCatalog"})
}

func CreateCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateCatalog"})
}

func UpdateCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateCatalog"})
}

func DeleteCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteCatalog"})
}
