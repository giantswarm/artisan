package release

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/helmclient/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/chart-operator/pkg/project"
)

func Test_DesiredState(t *testing.T) {
	testCases := []struct {
		name          string
		obj           *v1alpha1.Chart
		configMap     *apiv1.ConfigMap
		helmChart     helmclient.Chart
		secret        *apiv1.Secret
		expectedState ReleaseState
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "",
				ValuesYAML:        []byte("{}"),
				Version:           "0.1.2",
			},
		},
		{
			name: "case 1: basic match with empty config map",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{},
			},
			helmChart: helmclient.Chart{
				Version: "1.2.3",
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "",
				ValuesYAML:        []byte("{}"),
				Version:           "1.2.3",
			},
		},
		{
			name: "case 2: basic match with config map value",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `test: test`,
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "d27213d2ae2b24e8d1be0806469c564c",
				ValuesYAML:        []byte("test: test\n"),
				Version:           "0.1.2",
			},
		},
		{
			name: "case 3: config map with multiple values",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `"provider": "azure"
"replicas": 2`},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "dead3edde0c0c861d8bf4d83e2e4847a",
				ValuesYAML:        []byte("provider: azure\nreplicas: 2\n"),
				Version:           "0.1.2",
			},
		},
		{
			name: "case 4: config map not found",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-values-configmap",
					Namespace: "giantswarm",
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			errorMatcher: IsNotFound,
		},
		{
			name: "case 5: basic match with secret value",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"values": []byte(`"test": "test"`),
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "d27213d2ae2b24e8d1be0806469c564c",
				ValuesYAML:        []byte("test: test\n"),
				Version:           "0.1.2",
			},
		},
		{
			name: "case 6: secret with multiple values",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"values": []byte(`"secretpassword": "admin"
"secretnumber": 2`),
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "8ccfa2ed7f5cb9a125b5f53254c296a8",
				ValuesYAML:        []byte("secretnumber: 2\nsecretpassword: admin\n"),
				Version:           "0.1.2",
			},
		},
		{
			name: "case 7: secret not found",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-values-secret",
					Namespace: "giantswarm",
				},
			},
			errorMatcher: IsNotFound,
		},
		{
			name: "case 8: secret and configmap clash",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `"username": "admin"
"replicas": 2`,
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"values": []byte(`"username": "admin"
"secretnumber": 2`),
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmDeployedStatus,
				ValuesMD5Checksum: "80c4411b068b4b415a94b2b775797891",
				ValuesYAML:        []byte("replicas: 2\nsecretnumber: 2\nusername: admin\n"),
				Version:           "0.1.2",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			objs := make([]runtime.Object, 0, 0)
			if tc.configMap != nil {
				objs = append(objs, tc.configMap)
			}
			if tc.secret != nil {
				objs = append(objs, tc.secret)
			}

			var helmClient helmclient.Interface
			{
				c := helmclienttest.Config{
					LoadChartResponse: tc.helmChart,
				}
				helmClient = helmclienttest.New(c)
			}

			c := Config{
				Fs:         afero.NewMemMapFs(),
				G8sClient:  fake.NewSimpleClientset(),
				HelmClient: helmClient,
				K8sClient:  k8sfake.NewSimpleClientset(objs...),
				Logger:     microloggertest.New(),

				ProjectName: project.Name(),
			}
			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetDesiredState(context.TODO(), tc.obj)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			releaseState, err := toReleaseState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !cmp.Equal(releaseState.ValuesYAML, tc.expectedState.ValuesYAML) {
				desiredYAML := string(releaseState.ValuesYAML)
				expectedYAML := string(tc.expectedState.ValuesYAML)

				t.Fatalf("want matching ValuesYAML \n %s", cmp.Diff(desiredYAML, expectedYAML))
			}

			if !cmp.Equal(releaseState, tc.expectedState) {
				t.Fatalf("want matching ReleaseState \n %s", cmp.Diff(releaseState, tc.expectedState))
			}
		})
	}
}
