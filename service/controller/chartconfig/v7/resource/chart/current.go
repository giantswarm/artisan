package chart

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cordon annotation is set to prevent any changes to the chartconfig CR.
	// Delete annotation is set to indicate the chartconfig CR has been
	// migrated to an app CR and can be deleted.
	if key.IsCordoned(customObject) && !key.HasDeleteCROnlyAnnotation(customObject) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chart %#q has been cordoned until %#q due to reason %#q ", key.ChartName(customObject), key.CordonUntil(customObject), key.CordonReason(customObject)))

		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	releaseName := key.ReleaseName(customObject)
	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		// Return early as release is not installed.
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseHistory, err := r.helmClient.GetReleaseHistory(ctx, releaseName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ChartState{
		ChannelName:    key.ChannelName(customObject),
		ChartName:      key.ChartName(customObject),
		ChartValues:    releaseContent.Values,
		ReleaseName:    releaseName,
		ReleaseStatus:  releaseContent.Status,
		ReleaseVersion: releaseHistory.Version,
	}

	return chartState, nil
}
