package release

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/helmclient/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_CurrentState(t *testing.T) {
	testCases := []struct {
		name           string
		obj            *v1alpha1.Chart
		releaseContent *helmclient.ReleaseContent
		releaseHistory *helmclient.ReleaseHistory
		returnedError  error
		expectedState  ReleaseState
		expectedError  bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{
				Name:   "prometheus",
				Status: "DEPLOYED",
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			releaseHistory: &helmclient.ReleaseHistory{
				Name:    "prometheus",
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:   "prometheus",
				Status: "DEPLOYED",
				Values: map[string]interface{}{
					"key": "value",
				},
				Version: "0.1.2",
			},
		},
		{
			name: "case 1: different values",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{
				Name:   "prometheus",
				Status: "FAILED",
				Values: map[string]interface{}{
					"key":     "value",
					"another": "value",
				},
			},
			releaseHistory: &helmclient.ReleaseHistory{
				Name:    "prometheus",
				Version: "1.2.3",
			},
			expectedState: ReleaseState{
				Values: map[string]interface{}{
					"key":     "value",
					"another": "value",
				},
				Name:    "prometheus",
				Status:  "FAILED",
				Version: "1.2.3",
			},
		},
		{
			name: "case 2: empty state when error for no release present",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{},
			releaseHistory: &helmclient.ReleaseHistory{},
			returnedError:  fmt.Errorf("No such release: prometheus"),
			expectedError:  false,
		},
		{
			name: "case 3: unexpected error",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{},
			releaseHistory: &helmclient.ReleaseHistory{},
			returnedError:  fmt.Errorf("Unexpected error"),
			expectedError:  true,
		},
	}

	var err error

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var helmClient helmclient.Interface
			{
				c := helmclienttest.Config{
					DefaultError:          tc.returnedError,
					DefaultReleaseContent: tc.releaseContent,
					DefaultReleaseHistory: tc.releaseHistory,
				}

				helmClient, err = helmclienttest.New(c)
				if err != nil {
					t.Fatalf("error == %#v, want nil", err)
				}
			}

			c := Config{
				Fs:         afero.NewMemMapFs(),
				HelmClient: helmClient,
				K8sClient:  fake.NewSimpleClientset(),
				Logger:     microloggertest.New(),
			}

			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetCurrentState(context.TODO(), tc.obj)
			switch {
			case err != nil && !tc.expectedError:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.expectedError:
				t.Fatalf("error == nil, want non-nil")
			}

			ReleaseState, err := toReleaseState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !reflect.DeepEqual(ReleaseState, tc.expectedState) {
				t.Fatalf("ReleaseState == %q, want %q", ReleaseState, tc.expectedState)
			}
		})
	}

}
