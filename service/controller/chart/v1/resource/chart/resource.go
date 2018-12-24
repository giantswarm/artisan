package chart

import (
	"context"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "chartv1"
)

// Config represents the configuration used to create a new chart resource.
type Config struct {
	// Dependencies.
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	return nil
}

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func toChartState(v interface{}) (ChartState, error) {
	if v == nil {
		return ChartState{}, nil
	}

	chartState, ok := v.(*ChartState)
	if !ok {
		return ChartState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", chartState, v)
	}

	return *chartState, nil
}
