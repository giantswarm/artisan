package status

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient/v2/pkg/helmclient"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/pkg/annotation"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting status for release %#q", releaseName))

	// If something goes wrong outside of Helm we add that to the
	// controller context in the release resource. So we include this
	// information in the CR status.
	if cc.Status.Reason != "" {
		status := v1alpha1.ChartStatus{
			Reason: cc.Status.Reason,
			Release: v1alpha1.ChartStatusRelease{
				Status: cc.Status.Release.Status,
			},
		}

		err = r.setStatus(ctx, cr, status)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	releaseContent, err := r.helmClient.GetReleaseContent(ctx, key.Namespace(cr), releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q not found", releaseName))

		// There is no Helm release for this chart CR so its likely that
		// something has gone wrong. This could be for a reason outside
		// of Helm like the tarball URL is incorrect.
		//
		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var status, reason string
	{
		if key.IsCordoned(cr) {
			status = releaseStatusCordoned
			reason = key.CordonReason(cr)
		} else {
			status = releaseContent.Status
			if releaseContent.Status != helmclient.StatusDeployed {
				reason = releaseContent.Description
			}
		}
	}

	desiredStatus := v1alpha1.ChartStatus{
		AppVersion: releaseContent.AppVersion,
		Reason:     reason,
		Release: v1alpha1.ChartStatusRelease{
			LastDeployed: metav1.Time{Time: releaseContent.LastDeployed},
			Revision:     releaseContent.Revision,
			Status:       status,
		},
		Version: releaseContent.Version,
	}

	if !equals(desiredStatus, key.ChartStatus(cr)) {
		err = r.setStatus(ctx, cr, desiredStatus)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status for release %#q already set to %#q", releaseName, releaseContent.Status))
	}

	return nil
}

func (r *Resource) setStatus(ctx context.Context, cr v1alpha1.Chart, status v1alpha1.ChartStatus) error {
	if url, ok := cr.GetAnnotations()[annotation.WebhookURL]; ok {
		token := cr.GetAnnotations()[annotation.WebhookToken]

		err := updateAppStatus(url, token, status, r.httpClientTimeout)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("sending webhook to %#q failed", url), "stack", fmt.Sprintf("%#v", err))
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release %#q status to %#q", key.ReleaseName(cr), status.Release.Status))

	// Get chart CR again to ensure the resource version is correct.
	currentCR, err := r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Get(ctx, cr.Name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	currentCR.Status = status

	_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).UpdateStatus(ctx, currentCR, metav1.UpdateOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set status for release %#q", key.ReleaseName(cr)))

	return nil
}

func updateAppStatus(webhookURL, token string, status v1alpha1.ChartStatus, timeout time.Duration) error {
	request := Request{
		AppVersion:   status.AppVersion,
		LastDeployed: status.Release.LastDeployed,
		Reason:       status.Reason,
		Status:       status.Release.Status,
		Version:      status.Version,
	}

	if token != "" {
		request.Token = token
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return microerror.Mask(err)
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodPatch, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return microerror.Mask(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return microerror.Maskf(wrongStatusError, "expected http status '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}

	return nil
}
