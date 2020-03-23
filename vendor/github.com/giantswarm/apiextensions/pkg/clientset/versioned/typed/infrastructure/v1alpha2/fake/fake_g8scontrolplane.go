/*
Copyright 2020 Giant Swarm GmbH.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeG8sControlPlanes implements G8sControlPlaneInterface
type FakeG8sControlPlanes struct {
	Fake *FakeInfrastructureV1alpha2
	ns   string
}

var g8scontrolplanesResource = schema.GroupVersionResource{Group: "infrastructure.giantswarm.io", Version: "v1alpha2", Resource: "g8scontrolplanes"}

var g8scontrolplanesKind = schema.GroupVersionKind{Group: "infrastructure.giantswarm.io", Version: "v1alpha2", Kind: "G8sControlPlane"}

// Get takes name of the g8sControlPlane, and returns the corresponding g8sControlPlane object, and an error if there is any.
func (c *FakeG8sControlPlanes) Get(name string, options v1.GetOptions) (result *v1alpha2.G8sControlPlane, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(g8scontrolplanesResource, c.ns, name), &v1alpha2.G8sControlPlane{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.G8sControlPlane), err
}

// List takes label and field selectors, and returns the list of G8sControlPlanes that match those selectors.
func (c *FakeG8sControlPlanes) List(opts v1.ListOptions) (result *v1alpha2.G8sControlPlaneList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(g8scontrolplanesResource, g8scontrolplanesKind, c.ns, opts), &v1alpha2.G8sControlPlaneList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.G8sControlPlaneList{ListMeta: obj.(*v1alpha2.G8sControlPlaneList).ListMeta}
	for _, item := range obj.(*v1alpha2.G8sControlPlaneList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested g8sControlPlanes.
func (c *FakeG8sControlPlanes) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(g8scontrolplanesResource, c.ns, opts))

}

// Create takes the representation of a g8sControlPlane and creates it.  Returns the server's representation of the g8sControlPlane, and an error, if there is any.
func (c *FakeG8sControlPlanes) Create(g8sControlPlane *v1alpha2.G8sControlPlane) (result *v1alpha2.G8sControlPlane, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(g8scontrolplanesResource, c.ns, g8sControlPlane), &v1alpha2.G8sControlPlane{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.G8sControlPlane), err
}

// Update takes the representation of a g8sControlPlane and updates it. Returns the server's representation of the g8sControlPlane, and an error, if there is any.
func (c *FakeG8sControlPlanes) Update(g8sControlPlane *v1alpha2.G8sControlPlane) (result *v1alpha2.G8sControlPlane, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(g8scontrolplanesResource, c.ns, g8sControlPlane), &v1alpha2.G8sControlPlane{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.G8sControlPlane), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeG8sControlPlanes) UpdateStatus(g8sControlPlane *v1alpha2.G8sControlPlane) (*v1alpha2.G8sControlPlane, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(g8scontrolplanesResource, "status", c.ns, g8sControlPlane), &v1alpha2.G8sControlPlane{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.G8sControlPlane), err
}

// Delete takes name of the g8sControlPlane and deletes it. Returns an error if one occurs.
func (c *FakeG8sControlPlanes) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(g8scontrolplanesResource, c.ns, name), &v1alpha2.G8sControlPlane{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeG8sControlPlanes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(g8scontrolplanesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha2.G8sControlPlaneList{})
	return err
}

// Patch applies the patch and returns the patched g8sControlPlane.
func (c *FakeG8sControlPlanes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha2.G8sControlPlane, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(g8scontrolplanesResource, c.ns, name, pt, data, subresources...), &v1alpha2.G8sControlPlane{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.G8sControlPlane), err
}
