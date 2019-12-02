package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/chart-operator/flag/service/cnr"
	"github.com/giantswarm/chart-operator/flag/service/helm"
	"github.com/giantswarm/chart-operator/flag/service/image"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	CNR        cnr.CNR
	Helm       helm.Helm
	Image      image.Image
	Kubernetes kubernetes.Kubernetes
}
