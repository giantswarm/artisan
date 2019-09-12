package chart

import (
	"context"
	"encoding/json"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v5/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	name := key.ChartName(customObject)
	channel := key.ChannelName(customObject)

	// Values configmap contains settings managed by the controlling operator.
	chartConfigmapValues, err := r.getConfigMapValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// User configmap contains settings overridden by the user.
	userConfigmapValues, err := r.getUserConfigMapValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Merge configmap values. Custom values from the user override values
	// managed by the controlling operator.
	chartConfigmapValues, err = mergeValuesConfigMaps(chartConfigmapValues, userConfigmapValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartSecretValues, err := r.getSecretValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	chartValues, err := union(chartConfigmapValues, chartSecretValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	releaseVersion, err := r.apprClient.GetReleaseVersion(ctx, name, channel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ChartState{
		ChannelName: key.ChannelName(customObject),
		ChartName:   key.ChartName(customObject),
		ChartValues: chartValues,
		// DeleteCustomResourceOnly is set when the chartconfig CR has been
		// migrated to an app CR and can be safely deleted.
		DeleteCustomResourceOnly: key.HasDeleteCROnlyAnnotation(customObject),
		ReleaseName:              key.ReleaseName(customObject),
		ReleaseStatus:            releaseStatusDeployed,
		ReleaseVersion:           releaseVersion,
	}

	return chartState, nil
}

func (r *Resource) getConfigMapValues(ctx context.Context, customObject v1alpha1.ChartConfig) (map[string]interface{}, error) {
	chartValues := make(map[string]interface{})

	configMapName := key.ConfigMapName(customObject)
	configMapNamespace := key.ConfigMapNamespace(customObject)

	if configMapName != "" {
		configMap, err := r.getConfigMap(ctx, configMapName, configMapNamespace)
		if err != nil {
			return chartValues, microerror.Mask(err)
		}

		jsonData := configMap.Data["values.json"]
		if jsonData != "" {
			err = json.Unmarshal([]byte(jsonData), &chartValues)
			if err != nil {
				return chartValues, microerror.Mask(err)
			}
		}
	}

	return chartValues, nil
}

func (r *Resource) getUserConfigMapValues(ctx context.Context, customObject v1alpha1.ChartConfig) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	userValues := make(map[string]interface{})

	configMapName := key.UserConfigMapName(customObject)
	configMapNamespace := key.UserConfigMapNamespace(customObject)

	if configMapName != "" {
		configMap, err := r.getConfigMap(ctx, configMapName, configMapNamespace)
		if err != nil {
			return userValues, microerror.Mask(err)
		}

		for k, v := range configMap.Data {
			userValues[k] = v
		}
	}

	if len(userValues) > 0 {
		values["configmap"] = userValues
	}

	return values, nil
}

func (r *Resource) getConfigMap(ctx context.Context, configMapName, configMapNamespace string) (*corev1.ConfigMap, error) {
	configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, microerror.Maskf(notFoundError, "config map '%s' in namespace '%s' not found", configMapName, configMapNamespace)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return configMap, nil
}

func (r *Resource) getSecretValues(ctx context.Context, customObject v1alpha1.ChartConfig) (map[string]interface{}, error) {
	secretValues := make(map[string]interface{})

	if key.SecretName(customObject) != "" {
		secretName := key.SecretName(customObject)
		secretNamespace := key.SecretNamespace(customObject)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret '%s' in namespace '%s' not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		// TODO: fix this "secret.json" name somewhere and access it in release-operator.
		secretData := secret.Data["secret.json"]
		if secretData != nil {
			err = json.Unmarshal(secretData, &secretValues)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return secretValues, nil
}

// mergeValuesConfigMaps merges values generated by the controlling operator
// with userValues overriden by the user.
func mergeValuesConfigMaps(values, userValues map[string]interface{}) (map[string]interface{}, error) {
	if values == nil || len(values) == 0 {
		return userValues, nil
	}
	if userValues == nil || len(userValues) == 0 {
		return values, nil
	}

	// Add any top level user values not present in generated values.
	for userKey, userVals := range userValues {
		_, ok := values[userKey]
		if !ok {
			values[userKey] = userVals
		}
	}

	for key, raw := range values {
		vals, ok := raw.(map[string]interface{})
		if !ok {
			// Not a map. Nothing to merge.
			continue
		}

		userRaw, ok := userValues[key]
		if !ok {
			// No user values. Nothing to merge.
			continue
		}

		userVals, ok := userRaw.(map[string]interface{})
		if !ok {
			// User values should always be a map.
			return values, microerror.Maskf(invalidTypeError, "expected %T got %T", map[string]interface{}{}, userVals)
		}

		// Override with user value if there is a matching generated value.
		for k, v := range userVals {
			vals[k] = v
		}

		values[key] = vals
	}

	return values, nil
}

func union(a, b map[string]interface{}) (map[string]interface{}, error) {
	if a == nil {
		return b, nil
	}

	for k, v := range b {
		_, ok := a[k]
		if ok {
			// The secret and config map we use have at least one shared key. We can not
			// decide which value is supposed to be applied.
			return nil, microerror.Maskf(invalidConfigError, "secret and config map share the same key %s", k)
		}
		a[k] = v
	}
	return a, nil
}
