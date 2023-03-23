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
	manageiqv1alpha1 "github.com/PDeXchange/pac/apis/manageiq/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"

	"github.com/ppc64le-cloud/manageiq-client-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	t "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
)

var (
	MQClientSecret = "mq-client"
)

// ConfigReconciler reconciles a Config object
type ConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.pac.io,resources=configs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.pac.io,resources=configs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.pac.io,resources=configs/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Config object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	config := &appv1alpha1.Config{}

	if err := r.Get(ctx, req.NamespacedName, config); err != nil {
		l.Error(err, "unable to fetch Config")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret, err := r.reconcileSecret(ctx, req, config)
	if err != nil {
		return ctrl.Result{}, err
	}

	auth := &manageiq.KeycloakAuthenticator{
		UserName:        config.Spec.MIQUserName,
		Password:        string(secret.Data["miq-password"]),
		BaseURL:         config.Spec.MIQURL,
		KeycloakBaseURL: config.Spec.KeycloakURL,
		Realm:           config.Spec.KeycloakRealm,
		ClientID:        config.Spec.MIQClientID,
		ClientSecret:    string(secret.Data["miq-client-password"]),
	}

	mq := manageiq.NewClient(auth, manageiq.ClientParams{})

	if err := r.reconcileServices(ctx, mq, req, config); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ConfigReconciler) reconcileServices(ctx context.Context, mq *manageiq.Client, req ctrl.Request, config *appv1alpha1.Config) error {
	l := log.FromContext(ctx)
	slist, err := mq.ListServices(url.Values{"expand": []string{"resources"}})
	if err != nil {
		l.Error(err, "errored while listing the services")
		return err
	}
	for _, service := range slist.Resources {
		l.V(4).Info("service information", "ID", service.ID, "name", service.Name)
		s, err := mq.GetService(service.ID, url.Values{"attributes": []string{"vms"}})
		if err != nil {
			l.Error(err, "errored while fetching the services")
			return err
		}
		if s.Retired {
			l.V(3).Info("service retire, hence skipping", "name", s.Name)
			continue
		}
		//spew.Dump(s)
		constructService := func() (*manageiqv1alpha1.Service, error) {
			service := &manageiqv1alpha1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.Name,
					Namespace: req.Namespace,
				},
				Spec: manageiqv1alpha1.ServiceSpec{
					ID:        s.ID,
					CreatedAt: s.CreatedAt,
					Type:      manageiqv1alpha1.ServiceTypeVM,
				},
			}

			if len(s.VMs) >= 1 {
				service.Spec.VirtualMachine = &manageiqv1alpha1.VirtualMachine{
					Name: s.VMs[0].Name,
					ID:   s.VMs[0].UIDEMS,
					Ports: []manageiqv1alpha1.Port{
						{
							Number: 22,
							Type:   "tcp",
						},
					},
					Zone:            config.Spec.PowerVS.Zone,
					CloudInstanceID: config.Spec.PowerVS.CloudInstanceID,
					VPC: manageiqv1alpha1.VPC{
						ID:           config.Spec.VPC.ID,
						Region:       config.Spec.VPC.Region,
						Loadbalancer: config.Spec.VPC.LoadBalancerID,
					},
				}
			}
			if err := ctrl.SetControllerReference(config, service, r.Scheme); err != nil {
				return nil, err
			}
			return service, nil
		}

		ss, err := constructService()
		if err != nil {
			l.Error(err, "unable to construct k8s service")
			return err
		}

		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, ss, func() error {
			return nil
		}); err != nil {
			l.Error(err, "unable to create Service", "service", ss)
			return err
		}
	}

	return nil
}

// reconcileSecret populate the secret for the manageiq client
func (r *ConfigReconciler) reconcileSecret(ctx context.Context, req ctrl.Request, config *appv1alpha1.Config) (s *corev1.Secret, err error) {
	l := log.FromContext(ctx)

	secret := &corev1.Secret{}
	if err := r.Get(ctx, t.NamespacedName{Namespace: req.Namespace, Name: config.Spec.CredentialSecret.Name}, secret); err != nil {
		l.Error(err, "unable to fetch secret: %s", config.Spec.CredentialSecret.Name)
		return nil, err
	}

	defer func() {
		err = r.Update(ctx, secret)
	}()

	if _, ok := secret.Data["miq-password"]; !ok {
		return nil, fmt.Errorf("miq-password not found in the secret: %s", config.Spec.CredentialSecret.Name)
	}

	if _, ok := secret.Data["miq-client-password"]; !ok {
		return nil, fmt.Errorf("miq-client-password not found in the secret: %s", config.Spec.CredentialSecret.Name)
	}

	// TODO: Add code to generate the miq-client-password

	return secret, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Config{}).
		Complete(r)
}
