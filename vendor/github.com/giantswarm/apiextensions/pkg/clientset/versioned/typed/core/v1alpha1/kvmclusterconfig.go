/*
Copyright 2018 Giant Swarm GmbH.

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

package v1alpha1

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	scheme "github.com/giantswarm/apiextensions/pkg/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KVMClusterConfigsGetter has a method to return a KVMClusterConfigInterface.
// A group's client should implement this interface.
type KVMClusterConfigsGetter interface {
	KVMClusterConfigs(namespace string) KVMClusterConfigInterface
}

// KVMClusterConfigInterface has methods to work with KVMClusterConfig resources.
type KVMClusterConfigInterface interface {
	Create(*v1alpha1.KVMClusterConfig) (*v1alpha1.KVMClusterConfig, error)
	Update(*v1alpha1.KVMClusterConfig) (*v1alpha1.KVMClusterConfig, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.KVMClusterConfig, error)
	List(opts v1.ListOptions) (*v1alpha1.KVMClusterConfigList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KVMClusterConfig, err error)
	KVMClusterConfigExpansion
}

// kVMClusterConfigs implements KVMClusterConfigInterface
type kVMClusterConfigs struct {
	client rest.Interface
	ns     string
}

// newKVMClusterConfigs returns a KVMClusterConfigs
func newKVMClusterConfigs(c *CoreV1alpha1Client, namespace string) *kVMClusterConfigs {
	return &kVMClusterConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kVMClusterConfig, and returns the corresponding kVMClusterConfig object, and an error if there is any.
func (c *kVMClusterConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.KVMClusterConfig, err error) {
	result = &v1alpha1.KVMClusterConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KVMClusterConfigs that match those selectors.
func (c *kVMClusterConfigs) List(opts v1.ListOptions) (result *v1alpha1.KVMClusterConfigList, err error) {
	result = &v1alpha1.KVMClusterConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kVMClusterConfigs.
func (c *kVMClusterConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a kVMClusterConfig and creates it.  Returns the server's representation of the kVMClusterConfig, and an error, if there is any.
func (c *kVMClusterConfigs) Create(kVMClusterConfig *v1alpha1.KVMClusterConfig) (result *v1alpha1.KVMClusterConfig, err error) {
	result = &v1alpha1.KVMClusterConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		Body(kVMClusterConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kVMClusterConfig and updates it. Returns the server's representation of the kVMClusterConfig, and an error, if there is any.
func (c *kVMClusterConfigs) Update(kVMClusterConfig *v1alpha1.KVMClusterConfig) (result *v1alpha1.KVMClusterConfig, err error) {
	result = &v1alpha1.KVMClusterConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		Name(kVMClusterConfig.Name).
		Body(kVMClusterConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the kVMClusterConfig and deletes it. Returns an error if one occurs.
func (c *kVMClusterConfigs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kVMClusterConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kVMClusterConfig.
func (c *kVMClusterConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KVMClusterConfig, err error) {
	result = &v1alpha1.KVMClusterConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kvmclusterconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
