package controllers

import (
	"context"
	"fmt"
	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/tidwall/gjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func (r *HibernatorReconciler) handleLabelSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	factory := pkg.NewFactory(r.Mapper)
	resourceMapping, err := factory.MappingFor(rule.Type)
	if err != nil {
		return nil, err
	}
	request := &pkg.ListRequest{
		Namespace:            rule.Namespace,
		GroupVersionResource: resourceMapping.Resource,
		ListOptions: metav1.ListOptions{
			LabelSelector: strings.Join(rule.Labels, ","),
		},
	}
	resp, err := r.Kubectl.ListResources(context.Background(), request)
	if err != nil {
		return nil, err
	}
	for _, manifest := range resp.Manifests {
		fmt.Println(manifest)
	}
	return resp.Manifests, nil
}

func (r *HibernatorReconciler) handleFieldSelector(rule pincherv1alpha1.Selector) ([]unstructured.Unstructured, error) {
	var resp []unstructured.Unstructured
	var err error
	if len(rule.Labels) > 0 {
		resp, err = r.handleLabelSelector(rule)
	} else {
		resp, err = r.handleSelector(rule)
	}
	if err != nil {
		return nil, err
	}
	var matchedObjects []unstructured.Unstructured
	for _, field := range rule.FieldSelector {
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
	resourceMapping, err := factory.MappingFor(rule.Type)
	if err != nil {
		return nil, err
	}
	if len(rule.Name) > 0 {
		request := &pkg.GetRequest{
			Name:             rule.Name,
			Namespace:        rule.Namespace,
			GroupVersionKind: resourceMapping.GroupVersionKind,
		}
		resp, err := r.Kubectl.GetResource(context.Background(), request)
		if err != nil {
			return nil, err
		}
		fmt.Println(resp.Manifest)
		return []unstructured.Unstructured{resp.Manifest}, nil
	} else {
		request := &pkg.ListRequest{
			Namespace:            rule.Namespace,
			GroupVersionResource: resourceMapping.Resource,
			ListOptions:          metav1.ListOptions{},
		}
		resp, err := r.Kubectl.ListResources(context.Background(), request)
		if err != nil {
			return nil, err
		}
		for _, manifest := range resp.Manifests {
			fmt.Println(manifest)
		}
		return resp.Manifests, nil
	}
	return nil, nil
}
