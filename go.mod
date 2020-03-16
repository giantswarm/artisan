module github.com/giantswarm/chart-operator

go 1.13

require (
	github.com/giantswarm/apiextensions v0.1.1
	github.com/giantswarm/appcatalog v0.1.11
	github.com/giantswarm/backoff v0.0.0-20200209120535-b7cb1852522d
	github.com/giantswarm/e2esetup v0.0.0-20191209131007-01b9f9061692
	github.com/giantswarm/exporterkit v0.0.0-20190619131829-9749deade60f
	github.com/giantswarm/helmclient v0.0.0-20200316174225-0acb4df43c6f
	github.com/giantswarm/k8sclient v0.0.0-20200120104955-1542917096d6
	github.com/giantswarm/microendpoint v0.0.0-20200205204116-c2c5b3af4bdb
	github.com/giantswarm/microerror v0.2.0
	github.com/giantswarm/microkit v0.0.0-20191023091504-429e22e73d3e
	github.com/giantswarm/micrologger v0.2.0
	github.com/giantswarm/operatorkit v0.0.0-20200205163802-6b6e6b2c208b
	github.com/giantswarm/versionbundle v0.0.0-20200205145509-6772c2bc7b34
	github.com/google/go-cmp v0.4.0
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/viper v1.6.2
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	sigs.k8s.io/yaml v1.1.0
)
