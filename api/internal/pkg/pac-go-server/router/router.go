package router

import (
	"net/http"
	"os"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/tbaehler/gin-keycloak/pkg/ginkeycloak"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/services"

	_ "github.com/joho/godotenv/autoload"
)

var (
	clientId     = os.Getenv("KEYCLOAK_CLIENT_ID")
	clientSecret = os.Getenv("KEYCLOAK_CLIENT_SECRET")
	realm        = os.Getenv("KEYCLOAK_REALM")
	hostname     = os.Getenv("KEYCLOAK_HOSTNAME")

	client *gocloak.GoCloak
)

func init() {
	client = gocloak.NewClient(hostname)
}

func CreateRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	authorized := router.Group("/api/v1")
	authorized.Use(ginkeycloak.Auth(ginkeycloak.AuthCheck(), ginkeycloak.KeycloakConfig{
		Url:   hostname,
		Realm: realm,
	}))
	authorized.Use(RetrospectKeycloakToken)

	// Group routes
	authorized.GET("/groups", services.GetAllGroups)
	authorized.GET("/groups/:id", services.GetGroup)
	authorized.POST("/groups/:id/request", services.NewRequest)

	authorized.GET("/groups/:id/quota", services.GetQuota)
	authorized.POST("/groups/:id/quota", services.CreateQuota)
	authorized.PUT("/groups/:id/quota", services.UpdateQuota)
	authorized.DELETE("/groups/:id/quota", services.DeleteQuota)

	// Request routes
	// /requests?type=group to list only group add requests
	authorized.GET("/requests", services.GetAllRequests)
	authorized.GET("/requests/:id", services.GetRequest)
	authorized.DELETE("/request/:id", services.DeleteRequest)
	authorized.POST("/requests/:id/approve", services.ApproveRequest)
	authorized.POST("/requests/:id/reject", services.RejectRequest)

	// key related routes

	authorized.GET("/keys", services.GetAllKeys)
	authorized.GET("/keys/:id", services.GetKey)
	authorized.POST("/keys", services.CreateKey)
	authorized.DELETE("/keys/:id", services.DeleteKey)

	// catalog related endpoints

	// List all catalogs like vm, ocp, k8s
	authorized.GET("/catalogs", services.GetAllCatalogs)
	authorized.GET("/catalogs/:name", services.GetCatalog)

	// only for admins
	{
		authorized.POST("/catalogs", services.CreateCatalog)
		authorized.DELETE("/catalogs/:name", services.DeleteCatalog)
	}

	// service related endpoints

	// List all user provisioned services
	// services?all=true for admin to list all provisioned services
	authorized.GET("/services", services.GetAllServices)
	authorized.GET("/services/:name", services.GetService)
	authorized.POST("/services", services.CreateService)
	authorized.DELETE("/services/:name", services.DeleteService)
	// Currently, for extending the service expiry
	authorized.PUT("/services/:id/expiry", services.NewRequest)

	// quota related endpoints

	// list user quota
	authorized.GET("/quota", services.GetUserQuota)

	return router
}
