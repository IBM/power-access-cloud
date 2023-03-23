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

package manageiq

import (
	"context"
	"fmt"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/internal/pkg/client/iam"
	"github.com/PDeXchange/pac/internal/pkg/client/powervs"
	"github.com/PDeXchange/pac/internal/pkg/client/vpc"
	"github.com/pkg/errors"
	"github.com/ppc64le-cloud/manageiq-client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	t "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/url"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"time"

	manageiqv1alpha1 "github.com/PDeXchange/pac/apis/manageiq/v1alpha1"
)

var (
	ErrorPoolNotFound     = fmt.Errorf("pool not found")
	ErrorListenerNotFound = fmt.Errorf("listener not found")
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Debug  bool
}

//+kubebuilder:rbac:groups=manageiq.pac.io,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=manageiq.pac.io,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=manageiq.pac.io,resources=services/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

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

	service := &manageiqv1alpha1.Service{}

	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		l.Error(err, "unable to fetch CronJob")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.reconcileMQServiceStatus(ctx, service); err != nil {
		l.Error(err, "unable to reconcileMQService")
		return ctrl.Result{}, err
	}

	// TODO: Add code to remove the entry from the LB
	// If retired, then no further action is required
	if service.IsRetired() || service.IsDeleted() {
		l.V(4).Info("service is either retired or deleted", "retired", service.IsRetired(), "deleted", service.IsDeleted())
		return ctrl.Result{}, nil
	}

	switch service.Spec.Type {
	case manageiqv1alpha1.ServiceTypeVM:
		if service.Spec.VirtualMachine == nil {
			l.V(4).Info("VirtualMachine doesn't exist yet, hence reconcile after 2 mins")
			return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
		}
		auth, err := iam.GetIAMAuth()
		if err != nil {
			l.Error(err, "failed to authenticate")
			return ctrl.Result{}, err
		}
		if err := r.reconcileVM(ctx, auth, service); err != nil {
			l.Error(err, "unable to reconcileVM")
			return ctrl.Result{}, err
		}
	default:
		l.Error(fmt.Errorf("unknown service type"), string(service.Spec.Type))
		return ctrl.Result{}, fmt.Errorf("unknown service type")
	}

	if err := r.reconcileMQService(ctx, service); err != nil {
		l.Error(err, "unable to reconcileMQService")
		return ctrl.Result{}, err
	}

	l.Info("Reconcile completed")

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) reconcileVM(ctx context.Context, auth core.Authenticator, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)
	l.V(4).Info("in the reconcileVM")

	defer func() {
		if err := r.Status().Update(ctx, service); err != nil {
			l.Error(err, "unable to update Service status")
		}
	}()

	// This will initialise the status with ports from spec
	if err := reconcilePorts(ctx, r, service); err != nil {
		return err
	}

	// This will update the status with the right mac address
	if err := reconcileNetwork(ctx, r, service); err != nil {
		return err
	}

	// reconcileIPAddress reconcile the IP addresses from the dhcp server's leases and assign to status
	if err := reconcileIPAddress(ctx, r, service); err != nil {
		return err
	}

	if err := reconcileIngress(ctx, service); err != nil {
		return err
	}

	service.Status.SetReady()

	return nil
}

func (r *ServiceReconciler) reconcileMQServiceStatus(ctx context.Context, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)

	config, err := GetOwnerConfig(ctx, r.Client, service.ObjectMeta)
	if err != nil {
		l.Error(err, "errored while getting owner Config")
		return err
	}

	defer func() {
		if err := r.Status().Update(ctx, service); err != nil {
			l.Error(err, "unable to update Service status")
		}
	}()

	secret := &corev1.Secret{}
	if err := r.Get(ctx, t.NamespacedName{Namespace: config.Namespace, Name: config.Spec.CredentialSecret.Name}, secret); err != nil {
		l.Error(err, "unable to fetch secret: %s", config.Spec.CredentialSecret.Name)
		return err
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

	slist, err := mq.ListServices(url.Values{"expand": []string{"resources"}})
	if err != nil {
		l.Error(err, "errored listing services")
		return err
	}
	var serviceIDFound bool
	for _, s := range slist.Resources {
		if s.ID == service.Spec.ID {
			serviceIDFound = true
			break
		}
	}
	if !serviceIDFound {
		service.SetDeleted()
		service.SetNotReady()
		return nil
	}
	s, err := mq.GetService(service.Spec.ID, url.Values{})
	if err != nil {
		l.Error(err, "failed to get service")
		return err
	}
	if s.Retired {
		service.SetRetired()
		service.SetNotReady()
	}
	return nil
}

func (r *ServiceReconciler) reconcileMQService(ctx context.Context, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)
	if !service.Status.Ready {
		return fmt.Errorf("service is not ready yet")
	}
	config, err := GetOwnerConfig(ctx, r.Client, service.ObjectMeta)
	if err != nil {
		l.Error(err, "errored while getting owner Config")
		return err
	}

	secret := &corev1.Secret{}
	if err := r.Get(ctx, t.NamespacedName{Namespace: config.Namespace, Name: config.Spec.CredentialSecret.Name}, secret); err != nil {
		l.Error(err, "unable to fetch secret: %s", config.Spec.CredentialSecret.Name)
		return err
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

	var desc []string
	//TODO: Enhance the code to print the loadbalancer hostname only once
	for _, net := range service.Status.VirtualMachine.Ports {
		desc = append(desc, fmt.Sprintf("%s:%d(%d)", service.Status.VirtualMachine.Loadbalancer, net.Target, net.Number))
	}

	mq := manageiq.NewClient(auth, manageiq.ClientParams{})
	body := make(map[string]interface{})
	body["action"] = "edit"
	body["resource"] = map[string]string{
		"description": strings.Join(desc[:], ","),
	}
	if _, err := mq.UpdateService(service.Spec.ID, body); err != nil {
		l.Error(err, "failed to update the service with ingress information")
		return err
	}
	return nil
}

// GetOwnerConfig returns the Config object owning the current resource.
func GetOwnerConfig(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*appv1alpha1.Config, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "Config" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if gv.Group == appv1alpha1.GroupVersion.Group {
			return GetConfigByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// GetConfigByName finds and return a Cluster object using the specified params.
func GetConfigByName(ctx context.Context, c client.Client, namespace, name string) (*appv1alpha1.Config, error) {
	cluster := &appv1alpha1.Config{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, cluster); err != nil {
		return nil, errors.Wrapf(err, "failed to get Config/%s", name)
	}

	return cluster, nil
}

func reconcileIngress(ctx context.Context, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)

	l.V(4).Info("in the reconcileIngress")

	client, err := vpc.NewClient(ctx, vpc.Options{Region: service.Spec.VirtualMachine.VPC.Region})
	if err != nil {
		return err
	}

	lbID := service.Spec.VirtualMachine.VPC.Loadbalancer

	findLB := func(lb string) (*vpcv1.LoadBalancer, error) {
		lbs, _, err := client.ListLoadBalancers(&vpcv1.ListLoadBalancersOptions{})
		if err != nil {
			l.Error(err, "errored while listing the loadbalancers")
			return nil, err
		}
		// TODO: Handle the paging
		for _, lb := range lbs.LoadBalancers {
			if *lb.ID == lbID {
				return &lb, nil
			}
		}
		return nil, fmt.Errorf("loadbalancer not found: %s", lb)
	}

	lb, err := findLB(lbID)
	if err != nil {
		return err
	}

	// Step1: Ensure to create a pool
	findPool := func(port uint) (*vpcv1.LoadBalancerPool, error) {
		pools, _, err := client.ListLoadBalancerPools(&vpcv1.ListLoadBalancerPoolsOptions{LoadBalancerID: &lbID})
		if err != nil {
			return nil, err
		}
		for _, pool := range pools.Pools {
			if fmt.Sprintf("%s-%d", service.Spec.VirtualMachine.Name, port) == *pool.Name {
				return &pool, nil
			}
		}
		return nil, ErrorPoolNotFound
	}
	for _, port := range service.Status.VirtualMachine.Ports {
		if _, err := findPool(port.Number); err != nil && err != ErrorPoolNotFound {
			l.Error(err, "errored while finding pools")
			return err
		} else if err == ErrorPoolNotFound {
			opt := &vpcv1.CreateLoadBalancerPoolOptions{}
			opt.SetName(fmt.Sprintf("%s-%d", service.Spec.VirtualMachine.Name, port.Number))
			opt.SetAlgorithm("round_robin")
			opt.SetProtocol("tcp")
			opt.SetHealthMonitor(&vpcv1.LoadBalancerPoolHealthMonitorPrototype{
				Delay:      core.Int64Ptr(20),
				MaxRetries: core.Int64Ptr(2),
				Timeout:    core.Int64Ptr(5),
				Type:       core.StringPtr(string(port.Type)),
			})
			opt.SetLoadBalancerID(lbID)
			opt.SetMembers([]vpcv1.LoadBalancerPoolMemberPrototype{
				{
					Port:   core.Int64Ptr(int64(port.Number)),
					Target: &vpcv1.LoadBalancerPoolMemberTargetPrototypeIP{Address: core.StringPtr(service.Status.VirtualMachine.IPAddress)},
					Weight: core.Int64Ptr(100),
				},
			})
			if _, _, err := client.CreateLoadBalancerPool(opt); err != nil {
				l.Error(err, "errored while creating loadbalancer pool")
				return err
			}
			l.Info("need to write this code to create pool")
		}
	}

	// Step2: Ensure to create a listener
	findListener := func(name string) (*vpcv1.LoadBalancerListener, error) {
		l, _, err := client.ListLoadBalancerListeners(&vpcv1.ListLoadBalancerListenersOptions{LoadBalancerID: &lbID})
		if err != nil {
			return nil, err
		}

		for _, listener := range l.Listeners {
			if *listener.DefaultPool.Name == name {
				return &listener, nil
			}
		}
		return nil, ErrorListenerNotFound
	}
	for i, port := range service.Status.VirtualMachine.Ports {
		service.Status.VirtualMachine.Loadbalancer = *lb.Hostname
		listener, err := findListener(fmt.Sprintf("%s-%d", service.Spec.VirtualMachine.Name, port.Number))
		if err != nil && err != ErrorListenerNotFound {
			continue
		} else if err == ErrorListenerNotFound {
			pool, err := findPool(port.Number)
			if err != nil {
				l.Error(err, "backend pool not found", "pool", fmt.Sprintf("%s-%d", service.Spec.VirtualMachine.Name, port.Number))
				return err
			}

			generateRandPort := func() int64 {
				var (
					max int64 = 50000
					min int64 = 40000
				)
				rand.Seed(time.Now().UnixNano())
				return rand.Int63nRange(min, max)
			}

			createOpt := &vpcv1.CreateLoadBalancerListenerOptions{
				LoadBalancerID:  core.StringPtr(lbID),
				ConnectionLimit: core.Int64Ptr(15000),
				Port:            core.Int64Ptr(generateRandPort()),
				Protocol:        core.StringPtr("tcp"),
				DefaultPool:     &vpcv1.LoadBalancerPoolIdentity{ID: pool.ID},
			}
			listener, _, err = client.CreateLoadBalancerListener(createOpt)
			if err != nil {
				l.Error(err, "failed to CreateLoadBalancerListener")
				continue
			}
		}
		if listener != nil && listener.Port != nil {
			service.Status.VirtualMachine.Ports[i].Target = uint(*listener.Port)
		}
	}

	return nil
}

func reconcileIPAddress(ctx context.Context, r *ServiceReconciler, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)

	l.V(4).Info("in the reconcileIPAddress")

	vmStatus := &service.Status.VirtualMachine

	if vmStatus.Network == "" {
		l.Error(fmt.Errorf("network information is not yet available"), "retry after sometime")
		return fmt.Errorf("network information is not yet available")
	}
	client, err := powervs.NewClient(ctx, powervs.Options{
		CloudInstanceID: service.Spec.VirtualMachine.CloudInstanceID,
		Zone:            service.Spec.VirtualMachine.Zone,
		Debug:           r.Debug})
	if err != nil {
		return err
	}

	dhcpservers, err := client.GetAllDHCPServers()
	if err != nil {
		return err
	}

	for _, dhcpserver := range dhcpservers {
		if *dhcpserver.Network.Name == vmStatus.Network {
			s, err := client.GetDHCPServer(*dhcpserver.ID)
			if err != nil {
				return err
			}
			for _, lease := range s.Leases {
				if *lease.InstanceMacAddress == vmStatus.MACAddress && *lease.InstanceIP != "" {
					vmStatus.IPAddress = *lease.InstanceIP
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no lease found for the assigned mac address")
}

func reconcileNetwork(ctx context.Context, r *ServiceReconciler, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)
	l.V(4).Info("in the reconcileNetwork")
	client, err := powervs.NewClient(ctx, powervs.Options{
		CloudInstanceID: service.Spec.VirtualMachine.CloudInstanceID,
		Zone:            service.Spec.VirtualMachine.Zone,
		Debug:           r.Debug})
	if err != nil {
		l.Error(err, "failed to create powervs client")
		return err
	}

	in, err := func() (*models.PVMInstanceReference, error) {
		ins, err := client.GetAllInstance()
		if err != nil {
			l.Error(err, "failed to GetAllInstance")
			return nil, err
		}
		for _, in := range ins.PvmInstances {
			if service.Spec.VirtualMachine.Name == *in.ServerName {
				l.V(3).Info("found the vm!", "instance", in)
				return in, nil
			}
		}
		return nil, fmt.Errorf("vm not found")
	}()

	if err != nil {
		return err
	}

	service.Status.SetVirtualMachineStatusInstanceID(in)
	service.Status.SetVirtualMachineStatusMACAddress(in)

	return nil
}

func reconcilePorts(ctx context.Context, r *ServiceReconciler, service *manageiqv1alpha1.Service) error {
	l := log.FromContext(ctx)
	l.V(4).Info("in the reconcilePorts")

	addPorts := func(port manageiqv1alpha1.Port) {
		for _, p := range service.Status.VirtualMachine.Ports {
			if p.Number == port.Number {
				return
			}
		}
		service.Status.VirtualMachine.Ports = append(service.Status.VirtualMachine.Ports, manageiqv1alpha1.Port{Number: port.Number, Type: port.Type})
	}

	for _, port := range service.Spec.VirtualMachine.Ports {
		addPorts(port)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&manageiqv1alpha1.Service{}).
		Complete(r)
}
