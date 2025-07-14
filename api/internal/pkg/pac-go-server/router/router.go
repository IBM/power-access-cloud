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
	client       *gocloak.GoCloak
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

	authorizedAdmin := authorized.Group("")
	authorizedAdmin.Use(AllowAdminOnly)

	// Group routes
	authorized.GET("/groups", services.GetAllGroups)
	authorized.GET("/groups/:id", services.GetGroup)
	authorized.POST("/groups/:id/request", services.NewGroupRequest)
	// authorized.POST("/groups/:id/exit", services.ExitGroup)

	authorized.GET("/groups/:id/quota", services.GetQuota)

	// Request routes
	// /requests?type=group to list only group add requests
	authorized.GET("/requests", services.GetAllRequests)
	authorized.GET("/requests/:id", services.GetRequest)
	authorized.DELETE("/requests/:id", services.DeleteRequest)
	authorizedAdmin.POST("/requests/:id/approve", services.ApproveRequest)
	authorizedAdmin.POST("/requests/:id/reject", services.RejectRequest)

	// key related routes

	authorized.GET("/keys", services.GetAllKeysHandler)
	authorized.GET("/keys/:id", services.GetKey)
	authorized.POST("/keys", services.CreateKey)
	authorized.DELETE("/keys/:id", services.DeleteKeyHandler)

	// catalog related endpoints

	// List all catalogs like vm, ocp, k8s
	authorized.GET("/catalogs", services.GetAllCatalogs)
	authorized.GET("/catalogs/:name", services.GetCatalog)

	// only for admins
	{
		authorizedAdmin.POST("/catalogs", services.CreateCatalog)
		authorizedAdmin.DELETE("/catalogs/:name", services.DeleteCatalog)
		authorizedAdmin.PUT("/catalogs/:name/retire", services.RetireCatalog)

		authorizedAdmin.POST("/groups/:id/quota", services.CreateQuota)
		authorizedAdmin.PUT("/groups/:id/quota", services.UpdateQuota)
		authorizedAdmin.DELETE("/groups/:id/quota", services.DeleteQuota)

		authorizedAdmin.GET("/users", services.GetUsers)
		authorizedAdmin.GET("/users/:id", services.GetUser)
	}

	// user related endpoints
	authorized.DELETE("/users", services.DeleteUser)

	// service related endpoints

	// List all user provisioned services
	// services?all=true for admin to list all provisioned services
	authorized.GET("/services", services.GetAllServicesHandler)
	authorized.GET("/services/:name", services.GetService)
	authorized.POST("/services", services.CreateService)
	authorized.DELETE("/services/:name", services.DeleteServiceHandler)
	// Currently, for extending the service expiry
	authorized.PUT("/services/:name/expiry", services.UpdateServiceExpiryRequest)

	// quota related endpoints

	// list user quota
	authorized.GET("/quota", services.GetUserQuota)

	authorized.GET("/events", services.GetEvents)

	// terms and conditions related endpoints
	authorized.GET("/tnc", services.GetTermsAndConditionsStatus)
	authorized.POST("/tnc", services.AcceptTermsAndConditions)

	return router
}
