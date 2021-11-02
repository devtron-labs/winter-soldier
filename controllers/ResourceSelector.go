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

package controllers

import (
	"context"
	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

type ResourceSelector interface {
	handleLabelSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error)
	handleFieldSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error)
	handleSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error)
	getNamespaces(rule pincherv1alpha1.Selector, factory pkg.ArgsProcessor) ([]string, error)
	getMatchingObjects(selectors []pincherv1alpha1.Selector) []unstructured.Unstructured
	getIncludedExcludedObjects(inclusions, exclusions []unstructured.Unstructured) (included []unstructured.Unstructured, excluded []unstructured.Unstructured)
}

func NewResourceSelectorImpl(Kubectl pkg.KubectlCmd, Mapper *pkg.Mapper, factory func(mapper *pkg.Mapper) pkg.ArgsProcessor) ResourceSelector {
	return &ResourceSelectorImpl{
		Kubectl: Kubectl,
		Mapper:  Mapper,
		factory: factory,
	}
}

type ResourceSelectorImpl struct {
	Kubectl pkg.KubectlCmd
	Mapper  *pkg.Mapper
	factory func(mapper *pkg.Mapper) pkg.ArgsProcessor
}

func (r *ResourceSelectorImpl) handleLabelSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	factory := r.factory(r.Mapper)
	types := strings.Split(rule.ObjectSelector.Type, ",")
	namespaces, err := r.getNamespaces(rule, factory)
	if err != nil {
		return nil, err
	}
	var manifests []unstructured.Unstructured
	for _, t := range types {
		resourceMapping, err := factory.MappingFor(t)
		if err != nil {
			return nil, err
		}
		for _, namespace := range namespaces {
			request := &pkg.ListRequest{
				Namespace:            namespace,
				GroupVersionResource: resourceMapping.Resource,
				ListOptions: metav1.ListOptions{
					LabelSelector: strings.Join(rule.ObjectSelector.Labels, ","),
				},
			}
			resp, err := r.Kubectl.ListResources(context.Background(), request)
			if err != nil {
				continue
			}
			manifests = append(manifests, resp.Manifests...)
		}
	}

	return manifests, nil
}

func (r *ResourceSelectorImpl) handleFieldSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	var resp []unstructured.Unstructured
	var err error
	if len(rule.ObjectSelector.Labels) > 0 {
		resp, err = r.handleLabelSelector(rule)
	} else {
		resp, err = r.handleSelector(rule)
	}
	if err != nil {
		return nil, err
	}
	var matchedObjects []unstructured.Unstructured
	for _, r := range resp {
		j, err := r.MarshalJSON()
		if err != nil {
			continue
		}
		found := true
		for _, field := range rule.ObjectSelector.FieldSelector {
			if !pkg.ExpressionEvaluator(field, string(j)) {
				found = false
				break
			}
		}
		if found {
			matchedObjects = append(matchedObjects, r)
		}
	}
	return matchedObjects, nil
}

func (r *ResourceSelectorImpl) handleSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	factory := r.factory(r.Mapper)
	types := strings.Split(rule.ObjectSelector.Type, ",")
	namespaces, err := r.getNamespaces(rule, factory)
	if err != nil {
		return nil, err
	}
	if len(rule.ObjectSelector.Name) > 0 {
		names := strings.Split(rule.ObjectSelector.Name, ",")
		var manifests []unstructured.Unstructured
		for _, t := range types {
			resourceMapping, err := factory.MappingFor(t)
			if err != nil {
				return nil, err
			}
			for _, namespace := range namespaces {
				for _, name := range names {
					request := &pkg.GetRequest{
						Name:             name,
						Namespace:        namespace,
						GroupVersionKind: resourceMapping.GroupVersionKind,
					}
					resp, err := r.Kubectl.GetResource(context.Background(), request)
					if err != nil {
						continue
					}
					manifests = append(manifests, resp.Manifest)
				}
			}
		}
		return manifests, nil
	} else {
		var manifests []unstructured.Unstructured
		for _, t := range types {
			resourceMapping, err := factory.MappingFor(t)
			if err != nil {
				return nil, err
			}
			for _, namespace := range namespaces {
				request := &pkg.ListRequest{
					Namespace:            namespace,
					GroupVersionResource: resourceMapping.Resource,
					ListOptions:          metav1.ListOptions{},
				}
				resp, err := r.Kubectl.ListResources(context.Background(), request)
				if err != nil {
					continue
				}
				manifests = append(manifests, resp.Manifests...)
			}
		}
		return manifests, nil
	}
}

func (r *ResourceSelectorImpl) getNamespaces(rule pincherv1alpha1.Selector, factory pkg.ArgsProcessor) ([]string, error) {
	var namespaces []string
	if rule.NamespaceSelector.Name == "all" || len(rule.NamespaceSelector.Name) == 0 {
		resourceMapping, _ := factory.MappingFor("ns")
		listOptions := metav1.ListOptions{}
		if len(rule.NamespaceSelector.Labels) > 0 {
			listOptions = metav1.ListOptions{
				LabelSelector: strings.Join(rule.ObjectSelector.Labels, ","),
			}
		}
		request := &pkg.ListRequest{
			GroupVersionResource: resourceMapping.Resource,
			ListOptions:          listOptions,
		}
		resp, err := r.Kubectl.ListResources(context.Background(), request)
		if err != nil {
			return nil, err
		}
		for _, manifest := range resp.Manifests {
			namespaces = append(namespaces, manifest.GetName())
		}
	} else {
		namespaces = strings.Split(rule.NamespaceSelector.Name, ",")
	}
	return namespaces, nil
}

func (r *ResourceSelectorImpl) getMatchingObjects(selectors []pincherv1alpha1.Selector) []unstructured.Unstructured {
	var allMatches []unstructured.Unstructured
	for _, selector := range selectors {
		var err error
		var matches []unstructured.Unstructured
		if len(selector.ObjectSelector.FieldSelector) != 0 {
			matches, err = r.handleFieldSelector(selector)
		} else if len(selector.ObjectSelector.Labels) != 0 {
			matches, err = r.handleLabelSelector(selector)
		} else {
			matches, err = r.handleSelector(selector)
		}
		if err != nil {
			continue
		}
		allMatches = append(allMatches, matches...)
	}
	return allMatches
}

func (r *ResourceSelectorImpl) getIncludedExcludedObjects(inclusions, exclusions []unstructured.Unstructured) (included []unstructured.Unstructured, excluded []unstructured.Unstructured) {
	excludedKey := map[string]bool{}
	for _, exclusion := range exclusions {
		key := pkg.GetResourceKey(&exclusion)
		excludedKey[key.String()] = true
	}
	for _, inclusion := range inclusions {
		key := pkg.GetResourceKey(&inclusion)
		if excludedKey[key.String()] {
			excluded = append(excluded, inclusion)
		} else {
			included = append(included, inclusion)
		}
	}
	return included, excluded
}
