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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	capiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/controllers/util"
)

// CatalogReconciler reconciles a Catalog object
type CatalogReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Debug  bool
}

func filterOwnedServices(ctx context.Context, scope *ControllerScope) ([]client.Object, error) {
	var ownedServices []client.Object
	eachFunc := func(o runtime.Object) error {
		obj := o.(client.Object)
		acc, err := meta.Accessor(obj)
		if err != nil {
			return nil
		}

		if capiutil.IsOwnedByObject(acc, scope.Catalog) {
			ownedServices = append(ownedServices, obj)
		}

		return nil
	}

	var serviceList appv1alpha1.ServiceList
	if err := scope.Client.List(ctx, &serviceList); err != nil {
		return nil, errors.Wrap(err, "error listing services")
	}

	lists := []client.ObjectList{
		&serviceList,
	}

	for _, list := range lists {
		if err := meta.EachListItem(list, eachFunc); err != nil {
			return nil, errors.Wrapf(err, "error finding owned services of catalog %s/%s", scope.Catalog.Namespace, scope.Catalog.Name)
		}
	}

	return ownedServices, nil
}

func reconcileVMCatalog(ctx context.Context, scope *ControllerScope) error {
	scope.Logger.Info("Starting VM catalog reconciliation ...", "name", scope.Catalog.Name)

	vm := &scope.Catalog.Spec.VM

	if err := util.ValidateVMCapacity(&scope.Catalog.Spec.Capacity, &vm.Capacity); err != nil {
		return errors.Wrap(err, "error validating vm capacity")
	}

	powerVSGUID, _, _, _ := util.ParsePowerVSCRN(vm.CRN)

	powerVSInstance, err := scope.PlatformClient.GetResourceInstance(ctx, powerVSGUID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving powervs instance with id %s", powerVSGUID)
	}
	if *powerVSInstance.State != "active" {
		return errors.Errorf("powervs instance not in active state, current state: %s", *powerVSInstance.State)
	}

	image, err := scope.PowerVSClient.GetImageByName(vm.Image)
	if err != nil {
		return err
	}
	if *image.State != "active" {
		return errors.Errorf("image '%s' not in active state, current state: %s", vm.Image, *image.State)
	}

	if _, err = scope.PowerVSClient.GetNetworkByName(vm.Network); err != nil {
		return err
	}

	if err = util.ValidateSysType(vm.SystemType); err != nil {
		return err
	}

	if err = util.ValidateProcType(vm.ProcessorType); err != nil {
		return err
	}

	scope.Catalog.Status.Ready = true
	scope.Catalog.Status.Message = "catalog ready to use"

	scope.Logger.Info("Reconciled VM catalog", "name", scope.Catalog.Name)
	return nil
}

//+kubebuilder:rbac:groups=app.pac.io,resources=catalogs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.pac.io,resources=catalogs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.pac.io,resources=catalogs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Catalog object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CatalogReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Starting catalog reconciliation ...")

	catalog := &appv1alpha1.Catalog{}
	if err := r.Get(ctx, req.NamespacedName, catalog); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	scope, err := NewControllerScope(ctx, ControllerScopeParams{
		Type:    catalogController,
		Client:  r.Client,
		Logger:  l,
		Debug:   r.Debug,
		Catalog: catalog,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %v", err)
	}

	defer func() {
		if err := scope.Close(); err != nil && err == nil {
			l.Error(err, "error updating catalog status")
		}
	}()

	// Set ready as false if catalog is retired
	if catalog.Spec.Retired {
		catalog.Status.Message = "catalog is retired"
		return ctrl.Result{}, nil
	}

	if catalog.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(catalog, appv1alpha1.CatalogFinalizer) {
			controllerutil.AddFinalizer(catalog, appv1alpha1.CatalogFinalizer)
			if err = r.Update(ctx, catalog); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to add finalizer to catalog: %w", err)
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(catalog, appv1alpha1.CatalogFinalizer) {
			ownedServices, err := filterOwnedServices(ctx, scope)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(err, "error filtering services owned by catalog")
			}

			if len(ownedServices) > 0 {
				return ctrl.Result{RequeueAfter: time.Minute * 1}, errors.New("finalizer not removed since catalog still owning services")
			} else {
				controllerutil.RemoveFinalizer(catalog, appv1alpha1.CatalogFinalizer)
				if err = r.Update(ctx, catalog); err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to remove finalizer from catalog: %w", err)
				}
				l.Info("removed finalizer on catalog")
				return ctrl.Result{}, nil
			}
		}
	}

	switch catalog.Spec.Type {
	case appv1alpha1.CatalogTypeVM:
		if err = reconcileVMCatalog(ctx, scope); err != nil {
			catalog.Status.Ready = false
			catalog.Status.Message = err.Error()
			return ctrl.Result{}, errors.Wrap(err, "error reconciling vm catalog")
		}
	default:
		catalog.Status.Ready = false
		catalog.Status.Message = fmt.Sprintf("not able to idenitfy catalog type %s", catalog.Spec.Type)
	}

	l.Info("Reconciled catalog")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Catalog{}).
		Complete(r)
}
