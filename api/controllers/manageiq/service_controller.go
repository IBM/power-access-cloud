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
	"net/url"
	"strings"
	"time"

	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/pkg/errors"
	"github.com/ppc64le-cloud/manageiq-client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	t "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	l := log.FromContext(ctx)

	service := &manageiqv1alpha1.Service{}

	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		l.Error(err, "unable to fetch service")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	config, err := GetOwnerConfig(ctx, r.Client, service.ObjectMeta)
	if err != nil {
		l.Error(err, "errored while getting owner Config")
		return ctrl.Result{}, err
	}

	if config == nil {
		l.Error(err, "config is not present")
		return ctrl.Result{}, errors.New("config is not present")
	}

	scope, err := NewServiceScope(ctx, ServiceScopeParams{
		Client:  r.Client,
		Logger:  l,
		Service: service,
		Config:  config,
		Debug:   r.Debug,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %v", err)
	}

	defer func() {
		if err := scope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	controllerutil.AddFinalizer(service, manageiqv1alpha1.ServiceFinalizer)

	// Handle deleted machines.
	if !service.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, scope)
	}

	if err := r.reconcileMQServiceStatus(scope); err != nil {
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
		if err := r.reconcileVM(ctx, scope); err != nil {
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

func (r *ServiceReconciler) reconcileDelete(ctx context.Context, scope *ServiceScope) (_ ctrl.Result, reterr error) {
	l := log.FromContext(ctx)
	l.Info("Handling deleted Service")

	defer func() {
		if reterr == nil {
			// VSI is deleted so remove the finalizer.
			controllerutil.RemoveFinalizer(scope.Service, manageiqv1alpha1.ServiceFinalizer)
		}
	}()

	lbID := scope.Config.Spec.VPC.LoadBalancerID

	lb, _, err := scope.VPCClient.GetLoadBalancer(&vpcv1.GetLoadBalancerOptions{
		ID: &lbID,
	})
	if err != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, errors.Wrap(err, "failed to GetLoadBalancer ")
	}
	if *lb.ProvisioningStatus != "active" {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, errors.Errorf("loadbalancer state is not active(%s), hence delete later", *lb.ProvisioningStatus)
	}

	listeners, _, err := scope.VPCClient.ListLoadBalancerListeners(&vpcv1.ListLoadBalancerListenersOptions{
		LoadBalancerID: &lbID,
	})
	if err != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, errors.Wrap(err, "failed to list LoadBalancerListeners ")
	}

	for _, pool := range scope.BackendPools() {
		for _, listener := range listeners.Listeners {
			if *listener.DefaultPool.Name == pool {
				// Delete the listener
				if _, err := scope.VPCClient.DeleteLoadBalancerListener(&vpcv1.DeleteLoadBalancerListenerOptions{
					LoadBalancerID: &lbID,
					ID:             listener.ID,
				}); err != nil {
					return ctrl.Result{RequeueAfter: 30 * time.Second}, errors.Wrapf(err, "failed to delete LoadBalancerListener for id: %s", *listener.ID)
				}
			}
		}
	}

	pools, _, err := scope.VPCClient.ListLoadBalancerPools(&vpcv1.ListLoadBalancerPoolsOptions{
		LoadBalancerID: &lbID,
	})
	if err != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, errors.Wrap(err, "failed to list LoadBalancerListeners ")
	}

	for _, pool := range scope.BackendPools() {
		for _, p := range pools.Pools {
			if *p.Name == pool {
				// Delete the pool
				if _, err := scope.VPCClient.DeleteLoadBalancerPool(&vpcv1.DeleteLoadBalancerPoolOptions{
					LoadBalancerID: &lbID,
					ID:             p.ID,
				}); err != nil {
					return ctrl.Result{RequeueAfter: 30 * time.Second}, errors.Wrapf(err, "failed to delete LoadBalancerPool for id: %s", *p.ID)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) reconcileVM(ctx context.Context, scope *ServiceScope) error {
	l := scope.Logger
	l.V(4).Info("in the reconcileVM")

	defer func() {
		if err := r.Status().Update(ctx, scope.Service); err != nil {
			l.Error(err, "unable to update Service status")
		}
	}()

	// This will initialise the status with ports from spec
	if err := reconcilePorts(ctx, scope); err != nil {
		return err
	}

	if err := reconcileInstance(scope); err != nil {
		return err
	}

	// This will update the status with the right mac address
	if err := reconcileNetwork(scope); err != nil {
		return err
	}

	if err := reconcileIngress(scope); err != nil {
		return err
	}

	scope.Service.Status.SetReady()

	return nil
}

// reconcileInstance will reconcile the instance for the service
func reconcileInstance(scope *ServiceScope) error {
	l := scope.Logger
	l.V(4).Info("in the reconcileNetwork")

	in, err := func() (*models.PVMInstanceReference, error) {
		ins, err := scope.PowerVSClient.GetAllInstance()
		if err != nil {
			l.Error(err, "failed to GetAllInstance")
			return nil, err
		}
		for _, in := range ins.PvmInstances {
			if scope.Service.Spec.VirtualMachine.Name == *in.ServerName {
				l.V(3).Info("found the vm!", "instance", in)
				return in, nil
			}
		}
		return nil, fmt.Errorf("vm not found")
	}()

	if err != nil {
		return err
	}

	scope.Service.Status.SetVirtualMachineStatusInstanceID(in)

	return nil
}

func (r *ServiceReconciler) reconcileMQServiceStatus(scope *ServiceScope) error {
	service := scope.Service

	slist, err := scope.MIQClient.ListServices(url.Values{"expand": []string{"resources"}})
	if err != nil {
		scope.Logger.Error(err, "errored listing services")
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
	s, err := scope.MIQClient.GetService(service.Spec.ID, url.Values{})
	if err != nil {
		scope.Logger.Error(err, "failed to get service")
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

func reconcileIngress(scope *ServiceScope) error {
	l := scope.Logger

	l.V(4).Info("in the reconcileIngress")

	service := scope.Service

	lbID := service.Spec.VirtualMachine.VPC.Loadbalancer

	findLB := func(lb string) (*vpcv1.LoadBalancer, error) {
		lbs, _, err := scope.VPCClient.ListLoadBalancers(&vpcv1.ListLoadBalancersOptions{})
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
		pools, _, err := scope.VPCClient.ListLoadBalancerPools(&vpcv1.ListLoadBalancerPoolsOptions{LoadBalancerID: &lbID})
		if err != nil {
			return nil, err
		}
		for _, pool := range pools.Pools {
			if scope.BackendPool(int(port)) == *pool.Name {
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
			opt.SetName(scope.BackendPool(int(port.Number)))
			opt.SetAlgorithm("round_robin")
			opt.SetProtocol("tcp")
			opt.SetHealthMonitor(&vpcv1.LoadBalancerPoolHealthMonitorPrototype{
				Delay:      core.Int64Ptr(20),
				MaxRetries: core.Int64Ptr(2),
				Timeout:    core.Int64Ptr(5),
				Type:       core.StringPtr(port.Type),
			})
			opt.SetLoadBalancerID(lbID)
			opt.SetMembers([]vpcv1.LoadBalancerPoolMemberPrototype{
				{
					Port:   core.Int64Ptr(int64(port.Number)),
					Target: &vpcv1.LoadBalancerPoolMemberTargetPrototypeIP{Address: core.StringPtr(service.Status.VirtualMachine.IPAddress)},
					Weight: core.Int64Ptr(100),
				},
			})
			if _, _, err := scope.VPCClient.CreateLoadBalancerPool(opt); err != nil {
				l.Error(err, "errored while creating loadbalancer pool")
				return err
			}
			l.Info("need to write this code to create pool")
		}
	}

	// Step2: Ensure to create a listener
	findListener := func(name string) (*vpcv1.LoadBalancerListener, error) {
		l, _, err := scope.VPCClient.ListLoadBalancerListeners(&vpcv1.ListLoadBalancerListenersOptions{LoadBalancerID: &lbID})
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
		listener, err := findListener(scope.BackendPool(int(port.Number)))
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
			listener, _, err = scope.VPCClient.CreateLoadBalancerListener(createOpt)
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

func reconcileNetwork(scope *ServiceScope) error {
	l := scope.Logger
	l.V(4).Info("in the reconcileNetwork")

	virtualMachineStatus := &scope.Service.Status.VirtualMachine

	if virtualMachineStatus.InstanceID == "" {
		l.Error(fmt.Errorf("instance id is not yet available"), "retry after sometime")
		return fmt.Errorf("instance id is not yet available")
	}

	in, err := scope.PowerVSClient.GetInstance(virtualMachineStatus.InstanceID)
	if err != nil {
		return err
	}
	// Fetch the network information from the networks attached to the instance and set it in the status
	for _, network := range in.Networks {
		if network.MacAddress == "" {
			return fmt.Errorf("mac address is not yet available")
		}
		if network.Type == "fixed" {
			if network.IPAddress == "" {
				l.Error(fmt.Errorf("ip address is not yet available"), "retry after sometime")
				return fmt.Errorf("ip address is not yet available")
			}
			virtualMachineStatus.MACAddress = network.MacAddress
			virtualMachineStatus.Network = network.NetworkName
			virtualMachineStatus.IPAddress = network.IPAddress
			return nil
		} else if network.Type == "dynamic" {
			virtualMachineStatus.MACAddress = network.MacAddress
			virtualMachineStatus.Network = network.NetworkName
			dhcpservers, err := scope.PowerVSClient.GetAllDHCPServers()
			if err != nil {
				return fmt.Errorf("failed to get all dhcp servers: %w", err)
			}

			if err := func() error {
				for _, dhcpserver := range dhcpservers {
					if *dhcpserver.Network.Name == virtualMachineStatus.Network {
						s, err := scope.PowerVSClient.GetDHCPServer(*dhcpserver.ID)
						if err != nil {
							return err
						}
						for _, lease := range s.Leases {
							if *lease.InstanceMacAddress == virtualMachineStatus.MACAddress && *lease.InstanceIP != "" {
								virtualMachineStatus.IPAddress = *lease.InstanceIP
								return nil
							}
						}
					}
				}
				return fmt.Errorf("no lease found for the assigned mac address")
			}(); err != nil {
				return err
			}
		}
	}

	return nil
}

func reconcilePorts(ctx context.Context, scope *ServiceScope) error {
	l := log.FromContext(ctx)
	l.V(4).Info("in the reconcilePorts")

	addPorts := func(port manageiqv1alpha1.Port) {
		for _, p := range scope.Service.Status.VirtualMachine.Ports {
			if p.Number == port.Number {
				return
			}
		}
		scope.Service.Status.VirtualMachine.Ports = append(scope.Service.Status.VirtualMachine.Ports, manageiqv1alpha1.Port{Number: port.Number, Type: port.Type})
	}

	for _, port := range scope.Service.Spec.VirtualMachine.Ports {
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
