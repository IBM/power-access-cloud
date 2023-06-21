package services

import (
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client/kubernetes"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db"
)

var dbCon db.DB
var kubeClient kubernetes.Client

func SetDB(db db.DB) {
	dbCon = db
}

func SetKubeClient(client kubernetes.Client) {
	kubeClient = client
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

func DeleteService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteService"})
}
