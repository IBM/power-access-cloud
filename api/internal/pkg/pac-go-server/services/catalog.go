package services

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

// GetAllCatalogs	godoc
// @Summary		Get all catalogs
// @Description	Get all catalogs resource
// @Tags		catalogs
// @Accept		json
// @Produce		json
// @Param		Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success		200
// @Router		/api/v1/catalogs [get]
func GetAllCatalogs(c *gin.Context) {
	logger := log.GetLogger()
	catalogs, err := kubeClient.GetCatalogs()
	if err != nil {
		logger.Error("failed to get catalogs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	catalogsItems := convertToCatalogs(catalogs)
	logger.Debug("fetched catalogs", zap.Any("catalogs", catalogsItems))
	c.JSON(http.StatusOK, catalogsItems)
}

// GetCatalog			godoc
// @Summary			Get catalog as specified in request
// @Description		Get catalog resource
// @Tags			catalogs
// @Accept			json
// @Produce			json
// @Param			name path string true "catalog name to be fetched"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/catalogs/{name} [get]
func GetCatalog(c *gin.Context) {
	logger := log.GetLogger()
	catalogName := c.Param("name")
	if catalogName == "" {
		logger.Error("catalog name is not set")
		c.JSON(http.StatusBadRequest, gin.H{"error": "catalog name is not set"})
		return
	}

	catalog, err := kubeClient.GetCatalog(catalogName)
	if err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			logger.Error("catalog does not exists", zap.String("catalog name", catalogName))
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("catalog with name %s does not exists", catalogName)})
			return
		}
		logger.Error("failed to get catalog", zap.String("catalog name", catalogName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	catalogItem := convertToCatalog(catalog)
	logger.Debug("fetched catalog", zap.Any("catalog", catalogItem))
	c.JSON(http.StatusOK, catalogItem)
}

// CreateCatalog		godoc
// @Summary			Create catalog
// @Description		Create catalog resource
// @Tags			catalogs
// @Accept			json
// @Produce			json
// @Param			catalog body models.Catalog true "Create catalog"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/catalogs [post]
func CreateCatalog(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	catalog := models.Catalog{}
	if err := c.BindJSON(&catalog); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to bind request, Error: %v", err.Error())})
		return
	}

	logger.Debug("create catalog request", zap.Any("request", catalog))
	if err := validateCreateCatalogParams(catalog); len(err) > 0 {
		logger.Error("error in create catalog validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// if Expiry is not set, use Default value
	if catalog.Expiry == 0 {
		logger.Info("Catalog expiry is set to 0, which is invalid, using default expiry", zap.String("catalogName", catalog.Name), zap.Int("defaultExpiry", utils.DefaultExpirationDays))
		catalog.Expiry = utils.DefaultExpirationDays
	}

	if err := kubeClient.CreateCatalog(createCatalogObject(catalog)); err != nil {
		if errors.Is(err, utils.ErrResourceAlreadyExists) {
			logger.Error("catalog already exists", zap.String("catalog name", catalog.Name))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("catalog with name %s already exists", catalog.Name)})
			return
		}
		logger.Error("failed to create catalog", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	event, err := models.NewEvent(originator, originator, models.EventCatalogCreate)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Catalog %s created", catalog.Name))

	logger.Debug("successfully created catalog")
	c.Status(http.StatusCreated)
}

// DeleteCatalog		godoc
// @Summary			Delete catalog
// @Description		Delete catalog resource
// @Tags			catalogs
// @Accept			json
// @Produce			json
// @Param			name path string true "catalog name to be deleted"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/catalogs/{name} [delete]
func DeleteCatalog(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	catalogName := c.Param("name")
	if catalogName == "" {
		logger.Error("catalog name is not set")
		c.JSON(http.StatusBadRequest, gin.H{"error": "catalog name is not set"})
		return
	}

	if err := kubeClient.DeleteCatalog(catalogName); err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			logger.Error("catalog does not exists", zap.String("catalog name", catalogName))
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("catalog with name %s does not exists", catalogName)})
			return
		}
		logger.Error("failed to delete catalog", zap.String("catalog name", catalogName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	event, err := models.NewEvent(originator, originator, models.EventCatalogDelete)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Catalog %s deleted", catalogName))
	logger.Debug("successfully deleted catalog", zap.String("catalog name", catalogName))
	c.Status(http.StatusNoContent)
}

// RetireCatalog		godoc
// @Summary			Reire catalog
// @Description		Reire catalog resource
// @Tags			catalogs
// @Accept			json
// @Produce			json
// @Param			name path string true "catalog name to be retired"
// @Param			Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success			200
// @Router			/api/v1/catalogs/{name}/retire [put]
func RetireCatalog(c *gin.Context) {
	originator := c.Request.Context().Value("userid").(string)
	logger := log.GetLogger()
	catalogName := c.Param("name")
	if catalogName == "" {
		logger.Error("catalog name is not set")
		c.JSON(http.StatusBadRequest, gin.H{"error": "catalog name is not set"})
		return
	}

	if err := kubeClient.RetireCatalog(catalogName); err != nil {
		if errors.Is(err, utils.ErrResourceNotFound) {
			logger.Error("catalog does not exists", zap.String("catalog name", catalogName))
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("catalog with name %s does not exists", catalogName)})
			return
		}
		logger.Error("failed to retire catalog", zap.String("catalog name", catalogName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	event, err := models.NewEvent(originator, originator, models.EventCatalogRetire)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	defer func() {
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
		}
	}()

	event.SetLog(models.EventLogLevelINFO, fmt.Sprintf("Catalog %s retired", catalogName))
	logger.Debug("successfully retired catalog", zap.String("catalog name", catalogName))
	c.Status(http.StatusNoContent)
}

func convertToCatalog(catalogItem pac.Catalog) models.Catalog {
	catalog := models.Catalog{
		ID:          string(catalogItem.UID),
		Type:        string(catalogItem.Spec.Type),
		Name:        catalogItem.Name,
		Description: catalogItem.Spec.Description,
		Capacity: models.Capacity{
			Memory: catalogItem.Spec.Capacity.Memory,
		},
		Retired:                 catalogItem.Spec.Retired,
		Expiry:                  catalogItem.Spec.Expiry,
		ImageThumbnailReference: catalogItem.Spec.ImageThumbnailReference,
		Status: models.CatalogStatus{
			Ready:   catalogItem.Status.Ready,
			Message: catalogItem.Status.Message,
		},
	}
	switch catalogItem.Spec.Type {
	case pac.CatalogTypeVM:
		cpu, _ := utils.CastStrToFloat(catalogItem.Spec.VM.Capacity.CPU)
		catalog.VM = models.VM{
			CRN:           catalogItem.Spec.VM.CRN,
			ProcessorType: catalogItem.Spec.VM.ProcessorType,
			SystemType:    catalogItem.Spec.VM.SystemType,
			Image:         catalogItem.Spec.VM.Image,
			Network:       catalogItem.Spec.VM.Network,
			Capacity: models.Capacity{
				Memory: catalogItem.Spec.Capacity.Memory,
				CPU:    cpu,
			},
		}
	}
	cpu, _ := utils.CastStrToFloat(catalogItem.Spec.Capacity.CPU)
	catalog.Capacity.CPU = cpu
	return catalog
}

func convertToCatalogs(catalogList pac.CatalogList) []models.Catalog {
	catalogs := []models.Catalog{}
	for _, catalogItem := range catalogList.Items {
		catalogs = append(catalogs, convertToCatalog(catalogItem))
	}
	return catalogs
}

func validateCreateCatalogParams(catalog models.Catalog) []error {
	var errs []error
	if catalog.Type == "" {
		errs = append(errs, errors.New("catalog type should be set"))
	}
	//TODO: Consider having helper functions to get supported catalog types
	if catalog.Type != string(pac.CatalogTypeVM) {
		errs = append(errs, fmt.Errorf("invalid catalog type %s, only valid catalog is %v", catalog.Type, pac.CatalogTypeVM))
	}
	if catalog.Name == "" {
		errs = append(errs, errors.New("catalog name should be set"))
	}
	if catalog.Capacity.CPU == 0 {
		errs = append(errs, errors.New("catalog cpu capacity should be set"))
	}
	if catalog.Capacity.Memory == 0 {
		errs = append(errs, errors.New("catalog memory capacity should be set"))
	}
	if _, err := url.ParseRequestURI(catalog.ImageThumbnailReference); err != nil {
		errs = append(errs, errors.New("catalog image not valid"))
	}
	switch catalog.Type {
	case string(pac.CatalogTypeVM):
		vm := catalog.VM
		if vm.CRN == "" {
			errs = append(errs, errors.New("for catalog type VM crn should be set"))
		}
		if vm.SystemType == "" {
			errs = append(errs, errors.New("for catalog type VM system_type should be set"))
		}
		if vm.ProcessorType == "" {
			errs = append(errs, errors.New("for catalog type VM processor_type should be set"))
		}
		if vm.Image == "" {
			errs = append(errs, errors.New("for catalog type VM image should be set"))
		}
		if vm.Capacity.CPU == 0 {
			errs = append(errs, errors.New("for catalog type VM cpu capacity should be set"))
		}
		if vm.Capacity.Memory == 0 {
			errs = append(errs, errors.New("for catalog type VM memory capacity should be set"))
		}
	}
	return errs
}

func createCatalogObject(catalog models.Catalog) pac.Catalog {
	catalogItem := pac.Catalog{
		ObjectMeta: v1.ObjectMeta{
			Name:      catalog.Name,
			Namespace: "default",
		},
		Spec: pac.CatalogSpec{
			Type:        pac.CatalogType(catalog.Type),
			Description: catalog.Description,
			Capacity: pac.Capacity{
				CPU:    utils.CastFloatToStr(catalog.Capacity.CPU),
				Memory: catalog.Capacity.Memory,
			},
			Expiry:                  catalog.Expiry,
			ImageThumbnailReference: catalog.ImageThumbnailReference,
		},
	}
	switch catalog.Type {
	case string(pac.CatalogTypeVM):
		catalogItem.Spec.VM = pac.VMCatalog{
			CRN:           catalog.VM.CRN,
			ProcessorType: catalog.VM.ProcessorType,
			SystemType:    catalog.VM.SystemType,
			Image:         catalog.VM.Image,
			Network:       catalog.VM.Network,
			Capacity: pac.Capacity{
				CPU:    utils.CastFloatToStr(catalog.VM.Capacity.CPU),
				Memory: catalog.VM.Capacity.Memory,
			},
		}
	}
	return catalogItem
}
