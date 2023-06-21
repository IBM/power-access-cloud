package services

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db"
)

var dbCon db.DB

func SetDB(db db.DB) {
	dbCon = db
}

func GetAllServices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllServices"})
}

func GetService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetService"})
}

func CreateService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateService"})
}

func UpdateService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateService"})
}

func DeleteService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteService"})
}
