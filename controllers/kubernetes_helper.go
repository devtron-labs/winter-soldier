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
	"github.com/tidwall/gjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func (r *HibernatorReconciler) handleLabelSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	factory := pkg.NewFactory(r.Mapper)
	types := strings.Split(rule.ObjectSelector.Type, ",")
	namespaces, err := r.getNamespaces(rule, factory)
	if err != nil {
		return nil, err
	}
	var manifests []unstructured.Unstructured
	for _, namespace := range namespaces {
		for _, t := range types {
			resourceMapping, err := factory.MappingFor(t)
			if err != nil {
				return nil, err
			}
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
			for _, m := range resp.Manifests {
				manifests = append(manifests, m)
			}
		}
	}

	return manifests, nil
}

func (r *HibernatorReconciler) handleFieldSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
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
	for _, field := range rule.ObjectSelector.FieldSelector {
		ops, err := pkg.SplitByLogicalOperator(field)
		if err != nil {
			continue
		}
		for _, r := range resp {
			j, err := r.MarshalJSON()
			if err != nil {
				continue
			}
			res := gjson.Get(string(j), ops[1])
			match, err := pkg.ApplyLogicalOperator(res, ops)
			if err != nil {
				continue
			}
			if match {
				matchedObjects = append(matchedObjects, r)
			}
		}
	}
	return matchedObjects, nil
}

func (r *HibernatorReconciler) handleSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	factory := pkg.NewFactory(r.Mapper)
	types := strings.Split(rule.ObjectSelector.Type, ",")
	namespaces, err := r.getNamespaces(rule, factory)
	if err != nil {
		return nil, err
	}
	if len(rule.ObjectSelector.Name) > 0 {
		names := strings.Split(rule.ObjectSelector.Name, ",")
		var manifests []unstructured.Unstructured
		for _, namespace := range namespaces {
			for _, t := range types {
				resourceMapping, err := factory.MappingFor(t)
				if err != nil {
					return nil, err
				}
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
		for _, namespace := range namespaces {
			for _, t := range types {
				resourceMapping, err := factory.MappingFor(t)
				if err != nil {
					return nil, err
				}
				request := &pkg.ListRequest{
					Namespace:            namespace,
					GroupVersionResource: resourceMapping.Resource,
					ListOptions:          metav1.ListOptions{},
				}
				resp, err := r.Kubectl.ListResources(context.Background(), request)
				if err != nil {
					continue
				}
				for _, m := range resp.Manifests {
					manifests = append(manifests, m)
				}
			}
		}
		return manifests, nil
	}
	return nil, nil
}

func (r *HibernatorReconciler) getNamespaces(rule pincherv1alpha1.Selector, factory *pkg.ArgsProcessor) ([]string, error) {
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
