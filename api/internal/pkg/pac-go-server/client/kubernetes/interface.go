package kubernetes

import pac "github.com/PDeXchange/pac/apis/app/v1alpha1"

type Client interface {
	GetCatalogs() (pac.CatalogList, error)
	GetCatalog(string) (pac.Catalog, error)
	CreateCatalog(pac.Catalog) error
	DeleteCatalog(string) error
}
