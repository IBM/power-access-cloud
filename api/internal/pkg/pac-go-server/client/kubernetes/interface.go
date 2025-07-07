package kubernetes

import (
	"time"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
)

//go:generate mockgen -destination=mock_kubernetes_client.go -package=kubernetes . Client
type Client interface {
	GetCatalogs() (pac.CatalogList, error)
	GetCatalog(string) (pac.Catalog, error)
	CreateCatalog(pac.Catalog) error
	DeleteCatalog(string) error
	RetireCatalog(string) error

	GetServices(string string) (pac.ServiceList, error)
	GetService(string) (pac.Service, error)
	CreateService(pac.Service) error
	UpdateServiceExpiry(string, time.Time) error
	DeleteService(string, string) error
}
