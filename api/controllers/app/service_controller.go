/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/controllers/app/scope"
	appservice "github.com/IBM/power-access-cloud/api/controllers/app/service"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Debug  bool
}

//+kubebuilder:rbac:groups=app.pac.io,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.pac.io,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.pac.io,resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Service object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Starting service reconciliation ...")

	service := &appv1alpha1.Service{}
	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	catalog := &appv1alpha1.Catalog{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: service.Spec.Catalog.Name}, catalog); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error retrieving catalog with name %s for service %s", service.Spec.Catalog.Name, service.Name)
	}

	scope, err := scope.NewServiceScope(ctx, scope.ServiceScopeParams{
		ControllerScopeParams: scope.ControllerScopeParams{
			Client:  r.Client,
			Logger:  l,
			Debug:   r.Debug,
			Catalog: catalog,
		},
		Service: service,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %v", err)
	}

	defer func() {
		if err := scope.PatchServiceObject(); err != nil {
			l.Error(err, "error updating service status")
		}
	}()

	// If the catalog is retired, we should not allow any new services to be created. Hence, we set the service state to error.
	if service.Status.State != appv1alpha1.ServiceStateCreated && catalog.Spec.Retired {
		service.Status.State = appv1alpha1.ServiceStateError
		service.Status.Message = "catalog is retired"
		return ctrl.Result{}, nil
	}

	if service.Status.State != appv1alpha1.ServiceStateCreated && !catalog.Status.Ready {
		service.Status.State = appv1alpha1.ServiceStateError
		message := fmt.Sprintf("catalog %s not in ready state", service.Spec.Catalog.Name)
		service.Status.Message = message
		return ctrl.Result{}, fmt.Errorf("%s", message)
	}

	service.OwnerReferences = capiutil.EnsureOwnerRef(service.OwnerReferences, metav1.OwnerReference{
		APIVersion: catalog.APIVersion,
		Kind:       catalog.Kind,
		Name:       catalog.Name,
		UID:        catalog.UID,
	})

	var svc appservice.Interface

	switch catalog.Spec.Type {
	case appv1alpha1.CatalogTypeVM:
		svc = appservice.NewVM(scope)
	default:
		return ctrl.Result{}, errors.Errorf("unknown catalog type %s", catalog.Spec.Type)
	}

	if scope.Service.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(scope.Service, appv1alpha1.ServiceFinalizer) {
			controllerutil.AddFinalizer(scope.Service, appv1alpha1.ServiceFinalizer)
			if err = scope.PatchServiceObject(); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to add finalizer to catalog: %w", err)
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(scope.Service, appv1alpha1.ServiceFinalizer) {
			if _, err = svc.Delete(ctx); err != nil {
				return ctrl.Result{}, errors.Wrap(err, "error deleting the service")
			}

			controllerutil.RemoveFinalizer(scope.Service, appv1alpha1.ServiceFinalizer)
			if err = r.Update(ctx, scope.Service); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer from catalog: %w", err)
			}
			return ctrl.Result{}, nil
		}
	}

	if scope.IsExpired() && service.Status.State != appv1alpha1.ServiceStateExpired {
		service.Status.State = appv1alpha1.ServiceStateExpired
		service.Status.Expired = true
		service.Status.Message = "service expired"
		return ctrl.Result{}, nil
	}

	{
		switch scope.Service.Status.State {
		case "":
			scope.Service.Status.State = appv1alpha1.ServiceStateNew
			return ctrl.Result{}, nil
		case appv1alpha1.ServiceStateExpired:
			scope.Logger.Info("service expired", "name", scope.Service.ObjectMeta.Name)
			// Expired service will be deleted after 1 day by reconciler
			if time.Now().After(scope.Service.Spec.Expiry.Time.Add(24 * time.Hour)) {
				if err := scope.ControllerScope.Client.Delete(ctx, scope.Service); err != nil {
					return ctrl.Result{}, errors.Wrap(err, "failed to delete service")
				}
			}
			if _, err := svc.Delete(ctx); err != nil {
				return ctrl.Result{}, errors.Wrap(err, "error cleaning up service")
			}

			scope.Service.Status.AccessInfo = ""
			return ctrl.Result{}, nil
		case appv1alpha1.ServiceStateFailed:
			if scope.Service.Status.Successful {
				scope.Logger.Info("service in error state, but was successful created in the past, hence not taking any action", "name", scope.Service.ObjectMeta.Name)
				return ctrl.Result{}, nil
			}
			scope.Logger.Info("Service is in error state, hence recreating the service", "name", scope.Service.ObjectMeta.Name)
			scope.Service.Status.Message = "Service is in error state, hence recreating the service"
			if _, err := svc.Delete(ctx); err != nil {
				return ctrl.Result{}, errors.Wrap(err, "error cleaning up service")
			}
			scope.Service.Status.AccessInfo = ""
			scope.Service.Status.State = ""
			// give some time for the service to be deleted and requeue for next try
			return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
		}
	}

	if err := svc.Reconcile(ctx); err != nil {
		err = errors.Wrap(err, "error reconciling service")
		scope.Service.Status.State = appv1alpha1.ServiceStateError
		scope.Service.Status.Message = err.Error()

		return ctrl.Result{}, err
	}

	if scope.Service.Status.State == appv1alpha1.ServiceStateInProgress {
		l.Info("Service is in IN_PROGRESS state, requeuing after a min")
		return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Service{}).
		Complete(r)
}
