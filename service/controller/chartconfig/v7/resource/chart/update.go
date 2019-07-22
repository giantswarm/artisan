package chart

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartState, err := toChartState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartState.ReleaseName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chart %#q", chartState.ChartName))

		name := key.ChartName(customObject)
		releaseName := chartState.ReleaseName
		channel := chartState.ChannelName

		tarballPath, err := r.apprClient.PullChartTarball(ctx, name, channel)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := r.fs.Remove(tarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %q failed", tarballPath), "stack", fmt.Sprintf("%#q", err))
			}
		}()

		values, err := json.Marshal(chartState.ChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		upgradeForce, err := key.HasForceUpgradeAnnotation(customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.helmClient.UpdateReleaseFromTarball(ctx, releaseName, tarballPath,
			helm.UpdateValueOverrides(values),
			helm.UpgradeForce(upgradeForce))
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated chart %#q", chartState.ChartName))

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chart %#q does not need to be updated", chartState.ChartName))
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
	currentChartState, err := toChartState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredChartState, err := toChartState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the chart has to be updated")

	if isChartInTransitionState(currentChartState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the chart is in status %#q and cannot be updated", currentChartState.ReleaseStatus))
		return nil, nil
	}

	isModified := !currentChartState.IsEmpty() && isChartModified(currentChartState, desiredChartState)
	isWrongStatus := currentChartState.ReleaseStatus != desiredChartState.ReleaseStatus
	if isModified || isWrongStatus {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the chart has to be updated")

		return &desiredChartState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the chart does not have to be updated")
	}

	return nil, nil
}