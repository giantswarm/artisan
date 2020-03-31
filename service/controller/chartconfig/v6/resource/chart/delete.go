package chart

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
	"k8s.io/helm/pkg/helm"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	chartState, err := toChartState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// chartconfig CR has been migrated to an app CR and can be safely deleted
	// without deleting the related Helm release.
	if chartState.DeleteCustomResourceOnly {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting custom resource %#q but not release", chartState.ReleaseName))
		return nil
	}

	if chartState.ReleaseName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting release %s", chartState.ReleaseName))

		err = r.helmClient.DeleteRelease(ctx, chartState.ReleaseName, helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted release %s", chartState.ReleaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting release %s", chartState.ReleaseName))
	}
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentChartState, err := toChartState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartState, err := toChartState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %s release has to be deleted", desiredChartState.ReleaseName))

	if !currentChartState.IsEmpty() && currentChartState.Equals(desiredChartState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release needs to be deleted", desiredChartState.ReleaseName))

		return &desiredChartState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release does not need to be deleted", desiredChartState.ReleaseName))
	}

	return nil, nil
}