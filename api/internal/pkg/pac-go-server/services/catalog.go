package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func GetAllCatalogs(c *gin.Context) {
	logger := log.GetLogger()
	catalogs, err := kubeClient.GetCatalogs()
	if err != nil {
		logger.Error("failed to get catalogs", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%v", err)})
		return
	}
	catalogsItems := convertToCatalogs(catalogs)
	logger.Debug("fetched catalogs", zap.Any("catalogs", catalogsItems))
	c.JSON(http.StatusOK, catalogsItems)
}

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
		logger.Error("failed to get catalog", zap.String("catalog name", catalogName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	catalogItem := convertToCatalog(catalog)
	logger.Debug("fetched catalog", zap.Any("catalog", catalogItem))
	c.JSON(http.StatusOK, catalogItem)
}

func CreateCatalog(c *gin.Context) {
	logger := log.GetLogger()
	catalog := models.Catalog{}
	if err := c.BindJSON(&catalog); err != nil {
		logger.Error("failed to bin request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to bind request, Error: %v", err.Error())})
		return
	}
	logger.Debug("create catalog request", zap.Any("request", catalog))
	if err := validateCreateCatalogParams(catalog); len(err) > 0 {
		logger.Error("error in create catalog validation", zap.Errors("errors", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	if err := kubeClient.CreateCatalog(createCatalogObject(catalog)); err != nil {
		logger.Error("failed to create catalog", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("successfully created catalog")
	c.Status(http.StatusCreated)
}

func DeleteCatalog(c *gin.Context) {
	logger := log.GetLogger()
	catalogName := c.Param("name")
	if catalogName == "" {
		logger.Error("catalog name is not set")
		c.JSON(http.StatusBadRequest, gin.H{"error": "catalog name is not set"})
		return
	}

	if err := kubeClient.DeleteCatalog(catalogName); err != nil {
		logger.Error("failed to delete catalog", zap.String("catalog name", catalogName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	logger.Debug("successfully deleted catalog", zap.String("catalog name", catalogName))
	c.Status(http.StatusNoContent)
}

func convertToCatalog(catalogItem pac.Catalog) models.Catalog {
	catalog := models.Catalog{
		ID:          string(catalogItem.UID),
		Type:        string(catalogItem.Spec.Type),
		Name:        catalogItem.Name,
		Description: catalogItem.Spec.Description,
		Capacity: models.Capacity{
			CPU:    catalogItem.Spec.Capacity.CPU,
			Memory: catalogItem.Spec.Capacity.Memory,
		},
		Expiry: catalogItem.Spec.Expiry,
		Status: models.CatalogStatus{
			Ready:   catalogItem.Status.Ready,
			Message: catalogItem.Status.Message,
		},
	}
	switch catalogItem.Spec.Type {
	case pac.CatalogTypeVM:
		catalog.VM = models.VM{
			CRN:           catalogItem.Spec.VM.CRN,
			ProcessorType: catalogItem.Spec.VM.ProcessorType,
			SystemType:    catalogItem.Spec.VM.SystemType,
			Image:         catalogItem.Spec.VM.Image,
			Network:       catalogItem.Spec.VM.Network,
		}
	}
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
	if catalog.Capacity.CPU.String() == "" && catalog.Capacity.CPU.IntValue() == 0 {
		errs = append(errs, errors.New("catalog cpu capacity should be set"))
	}
	if catalog.Capacity.Memory == 0 {
		errs = append(errs, errors.New("catalog memory capacity should be set"))
	}
	if catalog.Expiry == 0 {
		errs = append(errs, errors.New("catalog expiry should be set"))
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
				CPU:    catalog.Capacity.CPU,
				Memory: catalog.Capacity.Memory,
			},
			Expiry: catalog.Expiry,
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
		}
	}
	return catalogItem
}
