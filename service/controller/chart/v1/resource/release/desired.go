package release

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(customResource)
	tarballURL := key.TarballURL(customResource)

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

	configMapValues, err := r.getConfigMapValues(ctx, customResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretValues, err := r.getSecretValues(ctx, customResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values, err := union(configMapValues, secretValues)
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
}

func (r *Resource) getConfigMapValues(ctx context.Context, customResource v1alpha1.Chart) (map[string]interface{}, error) {
	configMapValues := make(map[string]interface{})

	if key.ConfigMapName(customResource) != "" {
		configMapName := key.ConfigMapName(customResource)
		configMapNamespace := key.ConfigMapNamespace(customResource)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "config map %#q in namespace %#q not found", configMapName, configMapNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		jsonData := configMap.Data[controller.ConfigMapValuesKey]
		if jsonData != "" {
			err = json.Unmarshal([]byte(jsonData), &configMapValues)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return configMapValues, nil
}

func (r *Resource) getSecretValues(ctx context.Context, customResource v1alpha1.Chart) (map[string]interface{}, error) {
	secretValues := make(map[string]interface{})

	if key.SecretName(customResource) != "" {
		secretName := key.SecretName(customResource)
		secretNamespace := key.SecretNamespace(customResource)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret %#q in namespace %#q not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		jsonData := secret.Data[controller.SecretValuesKey]
		if jsonData != nil {
			err = json.Unmarshal(jsonData, &secretValues)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return secretValues, nil
}

func union(a, b map[string]interface{}) (map[string]interface{}, error) {
	if a == nil {
		return b, nil
	}

	for k, v := range b {
		_, ok := a[k]
		if ok {
			// The configmap and secret have at least one shared key. We cannot
			// decide which value should be applied.
			return nil, microerror.Maskf(invalidExecutionError, "configmap and secret share the same key %#q", k)
		}
		a[k] = v
	}
	return a, nil
}
