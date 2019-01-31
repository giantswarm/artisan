package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	releaseState, err := toReleaseState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if releaseState.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating release %#q", releaseState.Name))

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

		yamlValues, err := yaml.Marshal(releaseState.Values)
		if err != nil {
			return microerror.Mask(err)
		}

		// We need to pass the ValueOverrides option to make the update process
		// use the default values and prevent errors on nested values.
		err = r.helmClient.UpdateReleaseFromTarball(ctx, releaseState.Name, tarballPath, helm.UpdateValueOverrides(yamlValues))
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated release %#q", releaseState.Name))

	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %#q release has to be updated", desiredReleaseState.Name))

	if isReleaseInTransitionState(currentReleaseState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release is in status %#q and cannot be updated", desiredReleaseState.Name, currentReleaseState.Status))
		return nil, nil
	}

	isModified := !isEmpty(currentReleaseState) && isReleaseModified(currentReleaseState, desiredReleaseState)
	isWrongStatus := currentReleaseState.Status != desiredReleaseState.Status
	if isModified || isWrongStatus {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release has to be updated", desiredReleaseState.Name))

		return &desiredReleaseState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release does not have to be updated", desiredReleaseState.Name))
	}

	return nil, nil
}
