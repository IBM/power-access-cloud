package services

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTermsAndConditionsStatus(c *gin.Context) {
	userID := c.Request.Context().Value("userid").(string)
	tnc, err := dbCon.GetTermsAndConditionsByUserID(userID)
	if err != nil && errors.Unwrap(err) != mongo.ErrNoDocuments {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint("failed to get terms and conditions status: ", err.Error())})
		return
	} else if errors.Unwrap(err) == mongo.ErrNoDocuments {
		c.JSON(http.StatusOK, models.TermsAndConditions{UserID: userID})
		return
	}
	c.JSON(http.StatusOK, tnc)
}

func AcceptTermsAndConditions(c *gin.Context) {
	userID := c.Request.Context().Value("userid").(string)
	tnc, err := dbCon.GetTermsAndConditionsByUserID(userID)
	if err != nil && errors.Unwrap(err) != mongo.ErrNoDocuments {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint("failed to get terms and conditions status: ", err.Error())})
		return
	}
	if tnc != nil && tnc.Accepted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "terms and conditions already accepted"})
		return
	}
	timestamp := time.Now()
	tnc = &models.TermsAndConditions{
		UserID:     userID,
		Accepted:   true,
		AcceptedAt: &timestamp,
	}
	if err := dbCon.AcceptTermsAndConditions(tnc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint("failed to accept terms and conditions: ", err.Error())})
		return
	}
	c.Status(http.StatusCreated)
}
