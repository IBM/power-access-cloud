package services

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client/kubernetes"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

const maxServiceNameLength = 50

var dbCon db.DB
var kubeClient kubernetes.Client

func SetDB(db db.DB) {
	dbCon = db
}

func SetKubeClient(client kubernetes.Client) {
	kubeClient = client
}

func GetAllServices(c *gin.Context) {
	logger := log.GetLogger()
	var services pac.ServiceList
	var err error

	kc := client.NewKeyClockClient(c.Request.Context())
	userId := kc.GetUserID()

	listAllServices := c.DefaultQuery("all", "false")

	// list all the services for admin
	if listAllServices == "true" && kc.IsRole(utils.ManagerRole) {
		logger.Debug("listing all the services")
		services, err = kubeClient.GetServices("")
		if err != nil {
			logger.Error("failed to get services", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
	} else {
		logger.Debug("listing all the services of user", zap.String("user id", userId))
		services, err = kubeClient.GetServices(userId)
		if err != nil {
			logger.Error("failed to get services", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
	}
	serviceItems := convertToServices(services)
	logger.Debug("fetched services", zap.Any("services", serviceItems))
	c.JSON(http.StatusOK, serviceItems)
}

func GetService(c *gin.Context) {
	logger := log.GetLogger()
	serviceName := c.Param("name")
	if serviceName == "" {
		logger.Error("serviceName name is not set")
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceName name is not set"})
		return
	}
	service, err := kubeClient.GetService(serviceName)
	if err != nil {
		logger.Error("failed to get service", zap.String("service name", serviceName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	kc := client.NewKeyClockClient(c.Request.Context())
	userId := kc.GetUserID()

	// should not return service if the user is not admin or not owner of service
	if !kc.IsRole(utils.ManagerRole) {
		if service.Spec.UserID != userId {
			logger.Error("user is not the owner of service", zap.String("user id", userId), zap.String("service name", serviceName))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user id: %s is not owner of service %s", userId, service.Name)})
			return
		}
	}
	serviceItem := convertToService(service)
	logger.Debug("fetched service", zap.Any("service", serviceItem))
	c.JSON(http.StatusOK, serviceItem)
}

func CreateService(c *gin.Context) {
	logger := log.GetLogger()
	service := models.Service{}
	// bind user request
	if err := c.BindJSON(&service); err != nil {
		logger.Error("failed to bin request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to bind request, Error: %v", err.Error())})
		return
	}
	logger.Debug("create service request", zap.Any("request", service))

	// validate request params
	if err := validateCreateServiceParams(service); len(err) > 0 {
		logger.Error("error in create service validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// fetch catalog information
	catalog, err := kubeClient.GetCatalog(service.CatalogName)
	if err != nil {
		logger.Error("failed to get catalog", zap.String("catalog name", service.CatalogName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("catalog details", zap.String("name", service.CatalogName), zap.Any("catalog", catalog))

	if catalog.Spec.Retired {
		logger.Error("catalog is retired cannot deploy service", zap.Any("catalog", catalog))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("catalog %s is retired, cannot deploy service", catalog.Name)})
		return
	}

	if !catalog.Status.Ready {
		logger.Error("catalog is not in ready state cannot deploy service", zap.Any("catalog", catalog))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("catalog %s is not in ready state, cannot deploy service", catalog.Name)})
		return
	}

	// fetch userId
	kc := client.NewKeyClockClient(c.Request.Context())
	userId := kc.GetUserID()
	logger.Debug("user id", zap.String("userid", userId))

	// fetch ssh key of user
	sshKeys, err := dbCon.GetKeyByUserID(userId)
	if err != nil {
		logger.Error("failed to get ssh key for user", zap.String("userid", userId), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	if len(sshKeys) == 0 {
		logger.Error("no ssh keys found", zap.String("userid", userId), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "no ssh keys found"})
		return
	}
	var keys []string
	for _, userKey := range sshKeys {
		keys = append(keys, userKey.Content)
	}

	// fetch the user quota
	quota, err := getUserQuota(c)
	if err != nil {
		logger.Error("failed to get user quota", zap.String("userid", userId), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("user quota", zap.Any("quota", quota))

	// fetch the user used quota across all provisioned services
	usedQuota, err := getUsedQuota(userId)
	if err != nil {
		logger.Error("failed to get used quota", zap.String("userid", userId), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get used quota %v", err)})
		return
	}
	logger.Debug("user used quota", zap.Any("used quota", usedQuota))

	// calculate the total user capacity need to provision service
	neededCapacity, err := AddCapacity(usedQuota, catalog.Spec.Capacity)
	if err != nil {
		logger.Error("failed to needed capacity", zap.String("userid", userId), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get needed capacity %v", err)})
		return
	}
	logger.Debug("needed capacity", zap.Any("needed capacity", neededCapacity))

	// calculate the remaining capacity of user if provision this service
	remainingCapacity := models.Capacity{
		CPU:    quota.CPU - neededCapacity.CPU,
		Memory: quota.Memory - neededCapacity.Memory,
	}
	logger.Debug("remaining capacity", zap.Any("remaining capacity", remainingCapacity))

	if remainingCapacity.CPU < 0 || remainingCapacity.Memory < 0 {
		logger.Error("user does not have sufficient quota to provision service", zap.Any("required capacity", catalog.Spec.Capacity),
			zap.Any("user quota", quota), zap.Any("used capacity", usedQuota))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user does not have quota to provision resource, Quota: %v Required: %v Used: %v",
			quota, catalog.Spec.Capacity, usedQuota)})
		return
	}

	service.UserID = userId
	service.Expiry = time.Now().Add(time.Hour * 24 * time.Duration(catalog.Spec.Expiry))
	// generate unique service name
	serviceName := generateServiceName(service)

	// create service
	logger.Debug("service create params", zap.String("service name", serviceName), zap.Any("service", service), zap.Any("sshKey", sshKeys))
	if err := kubeClient.CreateService(createServiceObject(serviceName, keys, service)); err != nil {
		logger.Error("failed to create service", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("successfully created service")
	c.Status(http.StatusCreated)
}

func DeleteService(c *gin.Context) {
	logger := log.GetLogger()
	serviceName := c.Param("name")
	if serviceName == "" {
		logger.Error("service name is not set")
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is not set"})
		return
	}
	kc := client.NewKeyClockClient(c.Request.Context())
	userId := kc.GetUserID()

	//allow admin to delete the not owned services as well
	if kc.IsRole(utils.ManagerRole) {
		userId = ""
	}
	if err := kubeClient.DeleteService(serviceName, userId); err != nil {
		logger.Error("failed to delete service", zap.String("service name", serviceName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("successfully deleted service", zap.String("service name", serviceName))

	// Delete associated service-expiry-extension requests if present
	// Not generating event for request deletion as service is already deleted (To be discussed)
	req, err := dbCon.GetRequestByServiceName(serviceName)
	if err != nil {
		logger.Error("failed to fetch the request", zap.String("service name", serviceName), zap.Error(err))
		c.Status(http.StatusNoContent)
		return
	}
	logger.Debug("fetched request", zap.Any("request", req))
	for _, request := range req {
		if request.RequestType == models.RequestExtendServiceExpiry {
			if err := dbCon.DeleteRequest(request.ID.String()); err != nil {
				logger.Error("failed to delete request in database", zap.String("id", request.ID.String()), zap.Error(err))
			}
		}

	}
	c.Status(http.StatusNoContent)
}

func convertToService(serviceItem pac.Service) models.Service {
	service := models.Service{
		ID:          string(serviceItem.UID),
		UserID:      serviceItem.Spec.UserID,
		DisplayName: serviceItem.Spec.DisplayName,
		Name:        serviceItem.Name,
		CatalogName: serviceItem.Spec.Catalog.Name,
		Expiry:      serviceItem.Spec.Expiry.Time,
		Status: models.ServiceStatus{
			State:      string(serviceItem.Status.State),
			Message:    serviceItem.Status.Message,
			AccessInfo: serviceItem.Status.AccessInfo,
		},
	}
	return service
}

func convertToServices(serviceList pac.ServiceList) []models.Service {
	services := []models.Service{}
	for _, serviceItem := range serviceList.Items {
		services = append(services, convertToService(serviceItem))
	}
	return services
}

func validateCreateServiceParams(service models.Service) []error {
	var errs []error
	if service.DisplayName == "" {
		errs = append(errs, errors.New("display name should be set"))
	}
	if service.CatalogName == "" {
		errs = append(errs, errors.New("catalog name should be set"))
	}
	return errs
}

func createServiceObject(name string, sshKeys []string, service models.Service) pac.Service {
	serviceItem := pac.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kubernetes.DefaultNamespace,
		},
		Spec: pac.ServiceSpec{
			UserID:      service.UserID,
			DisplayName: service.DisplayName,
			Expiry: metav1.Time{
				Time: service.Expiry,
			},
			Catalog: corev1.LocalObjectReference{
				Name: service.CatalogName,
			},
			SSHKeys: sshKeys,
		},
	}
	return serviceItem
}

func generateServiceName(service models.Service) string {
	catalogName := service.CatalogName
	if len(catalogName) > maxServiceNameLength-6 {
		catalogName = catalogName[:(maxServiceNameLength - 6)]
	}
	name := fmt.Sprintf("%s-%s", catalogName, utilrand.String(5))
	return name
}

// getUsedQuota calculates and returns the total capacity consumed by user provisioned service
func getUsedQuota(userId string) (models.Capacity, error) {
	var consumedCapacity models.Capacity
	catalogMap := make(map[string]float64)
	serviceList, err := kubeClient.GetServices(userId)
	if err != nil {
		return consumedCapacity, fmt.Errorf("failed to get user services %v", err)
	}
	// fetch the catalogs associated with the service
	for _, svc := range serviceList.Items {
		// ignore the expired services
		if svc.Status.State == pac.ServiceStateExpired {
			continue
		}
		if _, ok := catalogMap[svc.Spec.Catalog.Name]; ok {
			catalogMap[svc.Spec.Catalog.Name] += 1
		} else {
			catalogMap[svc.Spec.Catalog.Name] = 1
		}
	}

	// calculate the total capacity of all the services
	for catalogName, count := range catalogMap {
		catalog, err := kubeClient.GetCatalog(catalogName)
		if err != nil {
			return consumedCapacity, fmt.Errorf("failed to get catalog, name %s %v", catalogName, err)
		}
		cpu, err := utils.CastStrToFloat(catalog.Spec.Capacity.CPU)
		if err != nil {
			return consumedCapacity, err
		}
		consumedCapacity.CPU += count * cpu
		consumedCapacity.Memory += int(count) * catalog.Spec.Capacity.Memory
	}
	return consumedCapacity, nil
}

//TODO: Move to utils if needed

func AddCapacity(capacity models.Capacity, catalogCapacity pac.Capacity) (models.Capacity, error) {
	cpu, err := utils.CastStrToFloat(catalogCapacity.CPU)
	if err != nil {
		return models.Capacity{}, err
	}
	return models.Capacity{
		CPU:    capacity.CPU + cpu,
		Memory: capacity.Memory + catalogCapacity.Memory,
	}, nil
}
