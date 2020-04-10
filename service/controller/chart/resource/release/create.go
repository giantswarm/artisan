package release

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
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
		if helmclient.IsPullChartFailedError(err) {
			reason := fmt.Sprintf("pulling chart %#q failed", tarballURL)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if helmclient.IsPullChartNotFound(err) {
			reason := fmt.Sprintf("chart %#q not found", tarballURL)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if helmclient.IsPullChartTimeout(err) {
			reason := fmt.Sprintf("timeout pulling %#q", tarballURL)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		defer func() {
			err := r.fs.Remove(tarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		ch := make(chan error)

		// We create the helm release but with a short timeout so we don't
		// block reconciling other CRs. This gives time to make the port
		// forwarding connection to the Tiller API.
		//
		// If we do timeout the install will continue in the background.
		// We will check the progress in the next reconciliation loop.
		go func() {
			// We need to pass the ValueOverrides option to make the install process
			// use the default values and prevent errors on nested values.
			err = r.helmClient.InstallReleaseFromTarball(ctx, tarballPath, ns, helm.ReleaseName(releaseState.Name), helm.ValueOverrides(releaseState.ValuesYAML))
			close(ch)
		}()

		select {
		case <-ch:
			// Fall through.
		case <-time.After(3 * time.Second):
			// We set the checksum annotation so the update state calculation
			// is accurate when we check in the next reconciliation loop.
			err = r.patchAnnotations(ctx, cr, releaseState)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "release still being created")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if err != nil {
			reason := err.Error()
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("helm release %#q failed", releaseState.Name), "stack", microerror.JSON(err))

			releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseState.Name)
			if helmclient.IsReleaseNotFound(err) {
				reason = fmt.Sprintf("helm error: (%s)", reason)
				addStatusToContext(cc, reason, releaseNotInstalledStatus)

				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}

			// Release is failed so the status resource will check the Helm release.
			if releaseContent.Status == helmFailedStatus {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("failed to create release %#q", releaseContent.Name))
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				return nil
			}
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