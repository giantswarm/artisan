package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	releaseState, err := toReleaseState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if releaseState.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating release %#q", releaseState.Name))

		ns := key.Namespace(cr)
		tarballURL := key.TarballURL(cr)

		tarballPath, err := r.helmClient.PullChartTarball(ctx, tarballURL)
		if err != nil {
			return microerror.Mask(err)
		}

		defer func() {
			err := r.fs.Remove(tarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		// We need to pass the ValueOverrides option to make the install process
		// use the default values and prevent errors on nested values.
		err = r.helmClient.InstallReleaseFromTarball(ctx, tarballPath, ns, helm.ReleaseName(releaseState.Name), helm.ValueOverrides(releaseState.ValuesYAML))
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.patchAnnotations(ctx, cr, releaseState)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created release %#q", releaseState.Name))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %#q release has to be created", desiredReleaseState.Name))

	createState := &ReleaseState{}

	if isEmpty(currentReleaseState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release needs to be created", desiredReleaseState.Name))

		createState = &desiredReleaseState
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release does not need to be created", desiredReleaseState.Name))
	}

	return createState, nil
}
