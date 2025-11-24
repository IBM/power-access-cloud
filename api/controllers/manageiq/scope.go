package manageiq

import (
	"context"
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/ppc64le-cloud/manageiq-client-go"
	appv1alpha1 "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	manageiqv1alpha1 "github.com/IBM/power-access-cloud/api/apis/manageiq/v1alpha1"
	"github.com/IBM/power-access-cloud/api/internal/pkg/client/powervs"
	"github.com/IBM/power-access-cloud/api/internal/pkg/client/vpc"
	corev1 "k8s.io/api/core/v1"
	t "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type ServiceScopeParams struct {
	Logger  logr.Logger
	Client  client.Client
	Service *manageiqv1alpha1.Service
	Config  *appv1alpha1.Config
	Debug   bool
}

type ServiceScope struct {
	logr.Logger
	Client      client.Client
	patchHelper *patch.Helper

	Service       *manageiqv1alpha1.Service
	Config        *appv1alpha1.Config
	MIQClient     *manageiq.Client
	PowerVSClient *powervs.Client
	VPCClient     *vpc.Client
}

func NewServiceScope(ctx context.Context, params ServiceScopeParams) (scope *ServiceScope, err error) {
	scope = &ServiceScope{}

	if params.Client == nil {
		err = errors.New("client is required when creating a ServiceScope")
		return
	}
	scope.Client = params.Client

	if params.Service == nil {
		err = errors.New("service is required when creating a ServiceScope")
		return
	}
	scope.Service = params.Service

	if params.Config == nil {
		err = errors.New("config is required when creating a ServiceScope")
		return
	}
	scope.Config = params.Config

	if params.Logger == (logr.Logger{}) {
		params.Logger = zap.New()
	}
	scope.Logger = params.Logger

	helper, err := patch.NewHelper(params.Service, params.Client)
	if err != nil {
		err = errors.Wrap(err, "failed to init patch helper")
		return
	}
	scope.patchHelper = helper

	config := params.Config
	secret := &corev1.Secret{}
	if err := params.Client.Get(ctx, t.NamespacedName{Namespace: config.Namespace, Name: config.Spec.CredentialSecret.Name}, secret); err != nil {
		scope.Logger.Error(err, "unable to fetch secret: %s", config.Spec.CredentialSecret.Name)
		return nil, err
	}

	auth := &manageiq.KeycloakAuthenticator{
		UserName:        config.Spec.MIQUserName,
		Password:        string(secret.Data["miq-password"]),
		BaseURL:         config.Spec.MIQURL,
		KeycloakBaseURL: config.Spec.KeycloakURL,
		Realm:           config.Spec.KeycloakRealm,
		ClientID:        config.Spec.MIQClientID,
		ClientSecret:    string(secret.Data["miq-client-password"]),
		Debug:           params.Debug,
	}
	scope.MIQClient = manageiq.NewClient(auth, manageiq.ClientParams{})

	client, err := powervs.NewClient(ctx, powervs.Options{
		CloudInstanceID: config.Spec.PowerVS.CloudInstanceID,
		Zone:            config.Spec.PowerVS.Zone,
		Debug:           params.Debug})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create powervs client")
	}
	scope.PowerVSClient = client

	if params.Debug {
		core.SetLoggingLevel(core.LevelDebug)
	}

	vpc, err := vpc.NewClient(ctx, vpc.Options{Region: config.Spec.VPC.Region})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vpc client")
	}
	scope.VPCClient = vpc

	return scope, nil
}

func (m *ServiceScope) Close() error {
	return m.PatchObject()
}

func (m *ServiceScope) PatchObject() error {
	return m.patchHelper.Patch(context.TODO(), m.Service)
}

func (m *ServiceScope) BackendPool(port int) (pools string) {
	return fmt.Sprintf("%s-%d", m.Service.Name, port)
}

func (m *ServiceScope) BackendPools() (pools []string) {
	if m.Service.Spec.VirtualMachine == nil {
		return
	}
	for _, port := range m.Service.Spec.VirtualMachine.Ports {
		pools = append(pools, m.BackendPool(int(port.Number)))
	}
	return
}
