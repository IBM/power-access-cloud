package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/client"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
	"golang.org/x/crypto/ssh"
)

// GetAllKeys			godoc
// @Summary			Get all keys
// @Description		Get all keys
// @Tags			keys
// @Accept			json
// @Produce			json
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/keys [get]
// Get the Key values and update.
func GetAllKeysHandler(c *gin.Context) {
	keys, err := getAllKeys(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, keys)
}

func getAllKeys(c *gin.Context) ([]models.Key, error) {
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())

	var userID string
	if !kc.IsRole(utils.ManagerRole) {
		// Get authenticated user's ID
		userID = kc.GetUserID()
	}
	keys, err := dbCon.GetKeyByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys, err : %w", err)
	}
	return keys, nil
}

// GetKey				godoc
// @Summary				Get key
// @Description			Get key
// @Tags				keys
// @Accept				json
// @Produce				json
// @Param				Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param				id path string true "key-id to be fetched"
// @Success				200
// @Router				/api/v1/keys/{id} [get]
func GetKey(c *gin.Context) {
	id := c.Param("id")
	key, err := dbCon.GetKeyByID(id)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("key with id %s not found", id)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, key)
}

// CreateKey			godoc
// @Summary			Create key
// @Description		Create key
// @Tags			keys
// @Accept			json
// @Produce			json
// @Param			key body models.Key true "Create key"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/keys [post]
func CreateKey(c *gin.Context) {
	var key = models.GetNewKey()
	// Step0: Get the authenticated user's ID
	userID := c.Request.Context().Value("userid").(string)

	if err := utils.BindAndValidate(c, &key); err != nil {
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

	keys, err := dbCon.GetKeyByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get all keys from the db, err: %s", err.Error())})
		return
	}
	for _, storedKey := range keys {
		if storedKey.Name == key.Name {
			c.JSON(http.StatusConflict, gin.H{"error": "Key name should be unique"})
			return
		}
	}

	// Insert the request into the database
	if err := dbCon.CreateKey(&key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert the key into the db, err: %s", err.Error())})
		return
	}
	c.Status(http.StatusCreated)
}

// DeleteKey			godoc
// @Summary			Delete key
// @Description		Delete key
// @Tags			keys
// @Accept			json
// @Produce			json
// @Param			id path string true "key-id to be deleted"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/keys/{id} [delete]
func DeleteKeyHandler(c *gin.Context) {
	err := deleteKey(c, c.Param("id"))
	if err != nil {
		if errors.Is(err, utils.ErrNotAuthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("not authorized to perform this action: %v", err)})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
	}
	c.Status(http.StatusNoContent)
}

func deleteKey(c *gin.Context, id string) error {
	key, err := dbCon.GetKeyByID(id)
	config := client.GetConfigFromContext(c.Request.Context())
	kc := client.NewKeyCloakClient(config, c.Request.Context())
	if err != nil {
		return fmt.Errorf("failed to fetch the requested record from the db, err: %w", err)
	}

	if key.UserID != c.Request.Context().Value("userid").(string) && !kc.IsRole(utils.ManagerRole) {
		return utils.ErrNotAuthorized
	}
	if err := dbCon.DeleteKey(id); err != nil {
		return fmt.Errorf("failed to delete the key from the db, err: %w", err)
	}
	return nil
}
