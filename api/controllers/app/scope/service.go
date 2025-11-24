package scope

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"sigs.k8s.io/cluster-api/util/patch"
)

type ServiceScopeParams struct {
	ControllerScopeParams
	Service *v1alpha1.Service
}

type ServiceScope struct {
	ControllerScope
	servicePatchHelper *patch.Helper
	Service            *v1alpha1.Service
}

func (s *ServiceScope) IsExpired() bool {
	currentTime := time.Now()
	return currentTime.After(s.Service.Spec.Expiry.Time)
}

// PatchObject persists the catalog/service configuration and status.
func (m *ServiceScope) PatchServiceObject() error {
	return m.servicePatchHelper.Patch(context.TODO(), m.Service)
}

func NewServiceScope(ctx context.Context, params ServiceScopeParams) (*ServiceScope, error) {
	scope := &ServiceScope{}

	ctrlScope, err := NewControllerScope(ctx, params.ControllerScopeParams)
	if err != nil {
		err = errors.Wrap(err, "failed to init controller scope")
		return scope, err
	}
	scope.ControllerScope = *ctrlScope

	if params.Service == nil {
		err = errors.New("service is required when creating a ServiceScope")
		return scope, err
	}
	scope.Service = params.Service

	serviceHelper, err := patch.NewHelper(params.Service, params.Client)
	if err != nil {
		err = errors.Wrap(err, "failed to init patch helper")
		return scope, err
	}
	scope.servicePatchHelper = serviceHelper

	return scope, nil
}
