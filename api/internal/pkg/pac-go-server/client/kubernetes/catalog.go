package kubernetes

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kClient "sigs.k8s.io/controller-runtime/pkg/client"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
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
			return catalog, fmt.Errorf("catalog with name %s does not exist", name)
		}
		return catalog, fmt.Errorf("failed to get catalog with name %s Error: %v", name, err)
	}
	return catalog, nil
}

func (client KubeClient) CreateCatalog(catalog pac.Catalog) error {
	if err := client.kubeClient.Create(context.Background(), &catalog); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("catalog with name %s already exist", catalog.Name)
		}
		return fmt.Errorf("failed to create catalog Error: %v", err)
	}
	return nil
}

func (client KubeClient) DeleteCatalog(name string) error {
	catalog := pac.Catalog{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &catalog); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete catalog with name %s Error: %v", name, err)
	}
	if err := client.kubeClient.Delete(context.Background(), &catalog); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete catalog with name %s Error: %v", name, err)
	}
	return nil
}
