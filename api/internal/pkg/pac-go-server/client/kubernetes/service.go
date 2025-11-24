package kubernetes

import (
	"context"
	"fmt"

	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kClient "sigs.k8s.io/controller-runtime/pkg/client"

	pac "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
)

func (client KubeClient) GetServices(userId string) (pac.ServiceList, error) {
	var services, servicesItems pac.ServiceList
	if err := client.kubeClient.List(context.Background(), &servicesItems); err != nil {
		return servicesItems, fmt.Errorf("failed to get services Error: %v", err)
	}

	if userId == "" {
		return servicesItems, nil
	}
	for _, service := range servicesItems.Items {
		if service.Spec.UserID == userId {
			services.Items = append(services.Items, service)
		}
	}
	services.TypeMeta = servicesItems.TypeMeta
	services.ListMeta = servicesItems.ListMeta
	return services, nil
}

func (client KubeClient) GetService(name string) (pac.Service, error) {
	service := pac.Service{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &service); err != nil {
		if apierrors.IsNotFound(err) {
			return service, utils.ErrResourceNotFound
		}
		return service, fmt.Errorf("failed to get service with name %s Error: %v", name, err)
	}
	return service, nil
}

func (client KubeClient) CreateService(service pac.Service) error {
	if err := client.kubeClient.Create(context.Background(), &service); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return utils.ErrResourceAlreadyExists
		}
		return fmt.Errorf("failed to create service Error: %v", err)
	}
	return nil
}

func (client KubeClient) DeleteService(name, userId string) error {
	service := pac.Service{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &service); err != nil {
		if apierrors.IsNotFound(err) {
			return utils.ErrResourceNotFound
		}
		return fmt.Errorf("error getting the service with name %s Error: %v", name, err)
	}
	// only allow admin and owner of the service to delete the service
	if userId != "" {
		if service.Spec.UserID != userId {
			return fmt.Errorf("user id: %s is not the owner of serivce %s", userId, service.Name)
		}
	}
	if err := client.kubeClient.Delete(context.Background(), &service); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("service with name %s does not exist", name)
		}
		return fmt.Errorf("failed to delete service with name %s Error: %v", name, err)
	}
	return nil
}

func (client KubeClient) UpdateServiceExpiry(name string, expiry time.Time) error {
	service := pac.Service{}
	if err := client.kubeClient.Get(context.Background(), kClient.ObjectKey{Namespace: DefaultNamespace, Name: name}, &service); err != nil {
		if apierrors.IsNotFound(err) {
			return utils.ErrResourceNotFound
		}
		return fmt.Errorf("failed to get service with name %s Error: %v", name, err)
	}
	service.Spec.Expiry = v1.Time{Time: expiry}
	if err := client.kubeClient.Update(context.Background(), &service); err != nil {
		return fmt.Errorf("failed to update service with name %s Error: %v", name, err)
	}
	return nil
}
