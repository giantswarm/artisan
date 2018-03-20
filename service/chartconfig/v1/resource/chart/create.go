package chart

import (
	"context"
	"fmt"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	chartState, err := toChartState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartState.ChartName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart %s", chartState.ChartName))
		name := key.ChartName(customObject)
		channel := key.ChannelName(customObject)

		tarballPath, err := r.apprClient.PullChartTarball(name, channel)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.helmClient.InstallFromTarball(tarballPath)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart %s", chartState.ChartName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not creating chart %s", chartState.ChartName))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentChartState, err := toChartState(currentState)
	if err != nil {
		return ChartState{}, microerror.Mask(err)
	}
	desiredChartState, err := toChartState(desiredState)
	if err != nil {
		return ChartState{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %s chart has to be created", desiredChartState.ChartName))

	createState := ChartState{}

	if currentChartState.ChartName == "" || desiredChartState.ChartName != currentChartState.ChartName {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s chart needs to be created", desiredChartState.ChartName))

		createState = desiredChartState
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s chart does not need to be created", desiredChartState.ChartName))
	}

	return createState, nil
}
