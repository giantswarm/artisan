// +build k8srequired

package basic

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
)

func TestChartInstalled(t *testing.T) {
	err := installChartOperatorResource(f)
	if err != nil {
		t.Fatalf("could not install chart-operator-resource-chart %v", err)
	}

	var rc *helmclient.ReleaseContent
	operation := func() error {
		rc, err = helmClient.GetReleaseContent("tb-release")
		if err != nil {
			return microerror.Maskf(err, "could not retrieve release content")
		}
		if rc.Status == "PENDING_INSTALL" {
			return microerror.Newf("release still not installed")
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for release %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		t.Fatal("expected nil found", err)
	}

	expectedStatus := "DEPLOYED"
	if rc.Status != expectedStatus {
		t.Fatalf("unexpected chart status, want %q, got %q", expectedStatus, rc.Status)
	}
}

func installChartOperatorResource(f *framework.Host) error {
	const chartOperatorResourceValues = `chart:
  name: "tb-chart"
  channel: "5-5-beta"
  namespace: "default"
  release: "tb-release"
`

	chartOperatorResourceValuesEnv := os.ExpandEnv(chartOperatorResourceValues)
	d := []byte(chartOperatorResourceValuesEnv)

	tmpfile, err := ioutil.TempFile("", "chart-operator-resource-values")
	if err != nil {
		return microerror.Mask(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(d)
	if err != nil {
		return microerror.Mask(err)
	}
	err = tmpfile.Close()
	if err != nil {
		return microerror.Mask(err)
	}

	err = framework.HelmCmd("registry install quay.io/giantswarm/chart-operator-resource-chart:stable -- -n chart-operator-resource --values " + tmpfile.Name())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
