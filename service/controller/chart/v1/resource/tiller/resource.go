package tiller

import (
	"context"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tillerv1"
)

// Config represents the configuration used to create a new tiller resource.
type Config struct {
	HelmClient helmclient.Interface
	Logger     micrologger.Logger
}

// Resource implements the tiller resource.
type Resource struct {
	helmClient helmclient.Interface
	logger     micrologger.Logger
}

// New creates a new configured tiller resource.
func New(config Config) (*Resource, error) {
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.ensureTillerInstalled(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	err := r.ensureTillerInstalled(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureTillerInstalled(ctx context.Context) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring tiller is installed")

	err := r.helmClient.EnsureTillerInstalled(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensured tiller is installed")

	return nil
}