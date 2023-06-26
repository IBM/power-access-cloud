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

package controllers

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

	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
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

	serviceToBePatched := client.MergeFrom(service.DeepCopy())

	defer func() {
		if service.ObjectMeta.DeletionTimestamp.IsZero() {
			if err := r.Status().Patch(ctx, service, serviceToBePatched); err != nil {
				l.Error(err, "error updating service status")
			}
		}
	}()

	catalog := &appv1alpha1.Catalog{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: service.Spec.Catalog.Name}, catalog); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error retrieving catalog with name %s for service %s", service.Spec.Catalog.Name, service.Name)
	}

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
		return ctrl.Result{}, errors.Errorf(message)
	}

	service.OwnerReferences = capiutil.EnsureOwnerRef(service.OwnerReferences, metav1.OwnerReference{
		APIVersion: catalog.APIVersion,
		Kind:       catalog.Kind,
		Name:       catalog.Name,
		UID:        catalog.UID,
	})

	scope, err := NewControllerScope(ctx, ControllerScopeParams{
		Type:    serviceController,
		Client:  r.Client,
		Logger:  l,
		Debug:   r.Debug,
		Service: service,
		Catalog: catalog,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %v", err)
	}

	if scope.Service.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(scope.Service, appv1alpha1.ServiceFinalizer) {
			controllerutil.AddFinalizer(scope.Service, appv1alpha1.ServiceFinalizer)
			if err = scope.Client.Update(ctx, scope.Service); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to add finalizer to catalog: %w", err)
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(scope.Service, appv1alpha1.ServiceFinalizer) {
			var isCleanupDone bool
			switch catalog.Spec.Type {
			case appv1alpha1.CatalogTypeVM:
				isCleanupDone, err = ReconcileDeleteVM(scope)
				if err != nil {
					return ctrl.Result{}, errors.Wrap(err, "error reconciling vm service")
				}
			}

			if isCleanupDone {
				controllerutil.RemoveFinalizer(scope.Service, appv1alpha1.ServiceFinalizer)
				if err = r.Update(ctx, scope.Service); err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to remove finalizer from catalog: %w", err)
				}
				return ctrl.Result{}, nil
			} else {
				// waiting for cleanup to complete before removing the finalizer, requeuing after a min
				return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
			}
		}
	}

	switch catalog.Spec.Type {
	case appv1alpha1.CatalogTypeVM:
		if scope.Service.Status.VM.InstanceID == "" {
			scope.Service.Status.State = appv1alpha1.ServiceStateNew
			if err = r.Status().Update(ctx, scope.Service); err != nil {
				l.Error(err, "error updating service status")
				return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
			}
		}

		if err = ReconcileVM(scope); err != nil {
			err = errors.Wrap(err, "error reconciling vm service")
			scope.Service.Status.State = appv1alpha1.ServiceStateError
			scope.Service.Status.Message = err.Error()

			return ctrl.Result{}, err
		}
		if scope.Service.Status.State == appv1alpha1.ServiceStateInProgress {
			l.Info("VM still in IN_PROGRESS state, requeuing after a min")
			return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Service{}).
		Complete(r)
}
