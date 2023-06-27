package services

import (
	"fmt"
	"net/http"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

// Get the Key values and update.
func GetAllKeys(c *gin.Context) {
	kc := utils.NewKeyClockClient(c.Request.Context())

	var userID string
	if !kc.IsRole(utils.ManagerRole) {
		// Get authenticated user's ID
		userID = kc.GetUserID()
	}
	keys, err := dbCon.GetKeyByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func GetKey(c *gin.Context) {
	id := c.Param("id")
	key, err := dbCon.GetKeyByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, key)
}

func CreateKey(c *gin.Context) {
	var key = models.GetNewKey()
	// Step0: Get the authenticated user's ID
	userID := c.Request.Context().Value("userid").(string)

	if err := c.BindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key.UserID = userID
	if key.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content cannot be empty."})
		return
	}

	if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key.Content)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ssh key"})
		return
	}
	// Validate the Key name length
	if len(key.Name) > 32 || key.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be 32 characters and cannot empty."})
		return
	}
	// Insert the request into the database
	if err := dbCon.CreateKey(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to insert the key into the db, err: %s", err.Error())})
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteKey(c *gin.Context) {
	id := c.Param("id")
	key, err := dbCon.GetKeyByID(id)
	kc := utils.NewKeyClockClient(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to fetch the requested record from the db, err: %s", err.Error())})
		return
	}

	if key.UserID != c.Request.Context().Value("userid").(string) && !kc.IsRole(utils.ManagerRole) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this key."})
		return
	}

	if err := dbCon.DeleteKey(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to delete the key from the db, err: %s", err.Error())})
		return
	}

	c.Status(http.StatusNoContent)
}
