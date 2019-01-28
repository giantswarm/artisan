package release

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/imdario/mergo"
	yaml "gopkg.in/yaml.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)
	tarballURL := key.TarballURL(cr)

	tarballPath, err := r.helmClient.PullChartTarball(ctx, tarballURL)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	defer func() {
		err := r.fs.Remove(tarballPath)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
		}
	}()

	chart, err := r.helmClient.LoadChart(ctx, tarballPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapValues, err := r.getConfigMapValues(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretValues, err := r.getSecretValues(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values, err := r.mergeValues(ctx, configMapValues, secretValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseState := &ReleaseState{
		Name:    releaseName,
		Status:  helmDeployedStatus,
		Values:  values,
		Version: chart.Version,
	}

	return releaseState, nil

	return nil, nil
}

func (r *Resource) getConfigMapValues(ctx context.Context, cr v1alpha1.Chart) (map[string]interface{}, error) {
	values := make(map[string]interface{})

	if key.ConfigMapName(cr) != "" {
		configMapName := key.ConfigMapName(cr)
		configMapNamespace := key.ConfigMapNamespace(cr)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "config map %#q in namespace %#q not found", configMapName, configMapNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		yamlData := configMap.Data[valuesKey]
		if yamlData != "" {
			err = yaml.Unmarshal([]byte(yamlData), &values)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return values, nil
}

func (r *Resource) getSecretValues(ctx context.Context, cr v1alpha1.Chart) (map[string]interface{}, error) {
	values := make(map[string]interface{})

	if key.SecretName(cr) != "" {
		secretName := key.SecretName(cr)
		secretNamespace := key.SecretNamespace(cr)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret %#q in namespace %#q not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		yamlData := secret.Data[valuesKey]
		if yamlData != nil {
			err = yaml.Unmarshal(yamlData, &values)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return values, nil
}

// mergeValues takes in the configmap and secret values and returns a single
// set of values to be passed to Tiller. If both contain data the values are
// merged.
func (r *Resource) mergeValues(ctx context.Context, configMapValues, secretValues map[string]interface{}) (map[string]interface{}, error) {
	var err error

	if !emptyValues(configMapValues) && emptyValues(secretValues) {
		// Return early.
		r.logger.LogCtx(ctx, "level", "debug", "message", "using configmap values")
		return configMapValues, nil
	}

	if emptyValues(configMapValues) && !emptyValues(secretValues) {
		// Return early.
		r.logger.LogCtx(ctx, "level", "debug", "message", "using secret values")
		return secretValues, nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "merging configmap and secret values")

	// Both maps contain values so merge them using mergo.
	values := configMapValues
	err = mergo.Merge(&values, secretValues)
	if err != nil {
		return map[string]interface{}{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "merged configmap and secret values")

	return values, nil
}

func emptyValues(values map[string]interface{}) bool {
	return reflect.DeepEqual(values, map[string]interface{}{})
}
