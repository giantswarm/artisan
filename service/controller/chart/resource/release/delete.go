package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient/v2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v2/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v2/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/operatorkit/v2/resource/crud"

	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseState, err := toReleaseState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if releaseState.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting release %#q", releaseState.Name))

		err = r.helmClient.DeleteRelease(ctx, key.Namespace(cr), releaseState.Name)
		if helmclient.IsReleaseNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q already deleted", releaseState.Name))
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		rel, err := r.helmClient.GetReleaseContent(ctx, key.Namespace(cr), releaseState.Name)
		if rel != nil {
			// Release still exists. We cancel the resource and keep the finalizer.
			// We will retry the delete in the next reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q still exists", releaseState.Name))

			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")

			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if helmclient.IsReleaseNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted release %#q", releaseState.Name))
		} else if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting release %#q", releaseState.Name))
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
	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %#q release has to be deleted", desiredReleaseState.Name))

	if !isEmpty(currentReleaseState) && currentReleaseState.Name == desiredReleaseState.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release needs to be deleted", desiredReleaseState.Name))

		return &desiredReleaseState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release does not need to be deleted", desiredReleaseState.Name))
	}

	return nil, nil
}
