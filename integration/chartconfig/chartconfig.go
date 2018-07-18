package chartconfig

import (
	"bytes"
	"html/template"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
)

type ChartConfigValues struct {
	Channel                  string
	ConfigMapName            string
	ConfigMapNamespace       string
	ConfigMapResourceVersion string
	Name                     string
	Namespace                string
	Release                  string
	SecretName               string
	SecretNamespace          string
	SecretResourceVersion    string
	VersionBundleVersion     string
}

func (ccv ChartConfigValues) ExecuteChartValuesTemplate() (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
