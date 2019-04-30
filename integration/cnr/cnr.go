// +build k8srequired

package cnr

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/chart-operator/integration/setup"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

type Chart struct {
	Channel string
	Release string
	Tarball string
	Name    string
}

func Push(ctx context.Context, config setup.Config, charts []Chart) error {
	var err error

	var forwarder *k8sportforward.Forwarder
	{
		c := k8sportforward.ForwarderConfig{
			RestConfig: config.CPK8sClients.RestConfig(),
		}

		forwarder, err = k8sportforward.NewForwarder(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = config.Release.Condition().PodExists(ctx, "giantswarm", "app=cnr-server")
	if err != nil {
		return microerror.Mask(err)
	}

	tunnel, err := forwarder.ForwardPort("giantswarm", podName, 5000)
	if err != nil {
		return microerror.Mask(err)
	}

	err = waitForServer("http://" + tunnel.LocalAddress() + "/cnr/api/v1/packages")
	if err != nil {
		return microerror.Mask(err)
	}

	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      "http://" + tunnel.LocalAddress(),
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}
	for _, chart := range charts {
		err = a.PushChartTarball(ctx, chart.Name, chart.Release, chart.Tarball)
		if err != nil {
			return microerror.Mask(err)
		}

		err = a.PromoteChart(ctx, chart.Name, chart.Release, chart.Channel)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func waitForServer(url string) error {
	var err error

	operation := func() error {
		_, err := http.Get(url)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for server at %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
