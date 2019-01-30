package chartconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/templates"
)

func InstallResources(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger) error {
	err := initializeCNR(ctx, h, helmClient, l)
	if err != nil {
		return microerror.Mask(err)
	}

	version := fmt.Sprintf(":%s", env.CircleSHA())
	err = h.InstallOperator("chart-operator", "chartconfig", templates.ChartOperatorValues, version)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func DeleteResources(ctx context.Context, helmClient *helmclient.Client, l micrologger.Logger) error {
	// Clean chartconfig related components.
	items := []string{"cnr-server", "apiextensions-chart-config-e2e"}

	for _, item := range items {
		err := helmClient.DeleteRelease(ctx, item, helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func VersionBundleVersion(githubToken, testedVersion string) string {
	if githubToken == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", env.EnvVarGithubBotToken))
	}
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", env.EnvVarTestedVersion))
	}

	params := &framework.VBVParams{
		Component: "chart-operator",
		Provider:  "aws",
		Token:     githubToken,
		VType:     testedVersion,
	}
	versionBundleVersion, err := framework.GetVersionBundleVersion(params)
	if err != nil {
		panic(err.Error())
	}

	if versionBundleVersion == "" {
		if strings.ToLower(testedVersion) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}
		panic("version bundle version  must not be empty")
	}

	return versionBundleVersion
}

func initializeCNR(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger) error {
	err := installCNR(ctx, h, helmClient, l)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCNR(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger) error {
	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	tarball, err := a.PullChartTarball(ctx, "cnr-server-chart", "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	err = helmClient.InstallReleaseFromTarball(ctx, tarball, "giantswarm", helm.ReleaseName("cnr-server"), helm.ValueOverrides([]byte("{}")), helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
