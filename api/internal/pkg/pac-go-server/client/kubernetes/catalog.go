package kubernetes

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kClient "sigs.k8s.io/controller-runtime/pkg/client"

	pac "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
)

func (client KubeClient) GetCatalogs() (pac.CatalogList, error) {
	catalogs := pac.CatalogList{}
	if err := client.kubeClient.List(context.Background(), &catalogs); err != nil {
		return catalogs, fmt.Errorf("failed to get catalogs Error: %v", err)
	}
	return catalogs, nil
}

func (client KubeClient) GetCatalog(name string) (pac.Catalog, error) {
	catalog := pac.Catalog{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &catalog); err != nil {
		if apierrors.IsNotFound(err) {
			return catalog, utils.ErrResourceNotFound
		}
		return catalog, fmt.Errorf("failed to get catalog with name %s Error: %v", name, err)
	}
	return catalog, nil
}

func (client KubeClient) CreateCatalog(catalog pac.Catalog) error {
	if err := client.kubeClient.Create(context.Background(), &catalog); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return utils.ErrResourceAlreadyExists
		}
		return fmt.Errorf("failed to create catalog Error: %v", err)
	}
	return nil
}

func (client KubeClient) DeleteCatalog(name string) error {
	catalog := pac.Catalog{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &catalog); err != nil {
		if apierrors.IsNotFound(err) {
			return utils.ErrResourceNotFound
		}
		return fmt.Errorf("failed to delete catalog with name %s Error: %v", name, err)
	}
	if err := client.kubeClient.Delete(context.Background(), &catalog); err != nil {
		return fmt.Errorf("failed to delete catalog with name %s Error: %v", name, err)
	}
	return nil
}

func (client KubeClient) RetireCatalog(name string) error {
	catalog := pac.Catalog{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &catalog); err != nil {
		if apierrors.IsNotFound(err) {
			return utils.ErrResourceNotFound
		}
		return fmt.Errorf("failed to retire catalog with name %s Error: %v", name, err)
	}
	catalog.Spec.Retired = true
	if err := client.kubeClient.Update(context.Background(), &catalog); err != nil {
		return fmt.Errorf("failed to retire catalog with name %s Error: %v", name, err)
	}
	return nil
}
