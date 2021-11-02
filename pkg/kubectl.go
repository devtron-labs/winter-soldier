/*
Copyright 2021 Devtron Labs Pvt Ltd.

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

package pkg

import (
	"context"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type KubectlCmd interface {
	ListResources(ctx context.Context, r *ListRequest) (*ListResponse, error)
	GetResource(ctx context.Context, r *GetRequest) (*ManifestResponse, error)
	DeleteResource(ctx context.Context, r *DeleteRequest) (*ManifestResponse, error)
	PatchResource(ctx context.Context, r *PatchRequest) (*ManifestResponse, error)
}

// FromKubeConfig creates a Cluster from a kubeConfig chain.
func FromKubeConfig() (*rest.Config, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return config, err
}

type kubectl struct {
	restConfig          *rest.Config
	extensionsClientset *clientset.Clientset
}

func NewKubectl() KubectlCmd {
	restConfig := ctrl.GetConfigOrDie()
	extensionsClientset, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil
	}
	return &kubectl{
		restConfig,
		extensionsClientset,
	}
}

func (k *kubectl) ListResources(ctx context.Context, r *ListRequest) (*ListResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	resourceList, err := dynamicIf.Resource(r.GroupVersionResource).Namespace(r.Namespace).List(ctx, r.ListOptions)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Manifests: resourceList.Items}, nil
}

func (k *kubectl) GetResource(ctx context.Context, r *GetRequest) (*ManifestResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	disco, err := discovery.NewDiscoveryClientForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(disco, r.GroupVersionKind)
	if err != nil {
		return nil, err
	}
	resource := r.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	obj, err := dynamicIf.Resource(resource).Namespace(r.Namespace).Get(ctx, r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (k *kubectl) DeleteResource(ctx context.Context, r *DeleteRequest) (*ManifestResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	disco, err := discovery.NewDiscoveryClientForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(disco, r.GroupVersionKind)
	if err != nil {
		return nil, err
	}
	resource := r.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	obj, err := dynamicIf.Resource(resource).Namespace(r.Namespace).Get(ctx, r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = dynamicIf.Resource(resource).Namespace(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (k *kubectl) PatchResource(ctx context.Context, r *PatchRequest) (*ManifestResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	disco, err := discovery.NewDiscoveryClientForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(disco, r.GroupVersionKind)
	if err != nil {
		return nil, err
	}
	resource := r.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	obj, err := dynamicIf.Resource(resource).Namespace(r.Namespace).Patch(ctx, r.Name, types.PatchType(r.PatchType), []byte(r.Patch), metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

// See: https://github.com/ksonnet/ksonnet/blob/master/utils/client.go
func ServerResourceForGroupVersionKind(disco discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	resources, err := disco.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}
	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			//log.Debugf("Chose API '%s' for %s", r.Name, gvk)
			return &r, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, "")
}
