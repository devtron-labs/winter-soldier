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
	"fmt"
	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"strconv"
)

type Execute func(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)

type ResourceAction interface {
	DeleteAction(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)
	HibernateAction(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)
	UnHibernateActionFactory(hibernator *pincherv1alpha1.Hibernator) Execute
}

func NewResourceActionImpl(kubectl pkg.KubectlCmd, historyUtil History) ResourceAction {
	return &ResourceActionImpl{
		Kubectl:     kubectl,
		historyUtil: historyUtil,
	}
}

type ResourceActionImpl struct {
	Kubectl     pkg.KubectlCmd
	historyUtil History
}

func (r *ResourceActionImpl) DeleteAction(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject) {
	impactedObjects := make([]pincherv1alpha1.ImpactedObject, 0)
	excludedObjects := make([]pincherv1alpha1.ExcludedObject, 0)

	for _, inc := range included {

		impactedObject := pincherv1alpha1.ImpactedObject{
			ResourceKey: getResourceKey(inc),
			Status:      "success",
		}

		request := &pkg.DeleteRequest{
			Name:             inc.GetName(),
			Namespace:        inc.GetNamespace(),
			GroupVersionKind: inc.GroupVersionKind(),
			Force:            pointer.BoolPtr(true),
		}
		_, err := r.Kubectl.DeleteResource(context.Background(), request)

		if err != nil {
			//log.Error(err, "action on object %s", requestObj)
			impactedObject.Status = "error"
			impactedObject.Message = err.Error()
		}

		impactedObjects = append(impactedObjects, impactedObject)
	}
	return impactedObjects, excludedObjects
}

func (r *ResourceActionImpl) HibernateAction(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject) {
	impactedObjects := make([]pincherv1alpha1.ImpactedObject, 0)
	excludedObjects := make([]pincherv1alpha1.ExcludedObject, 0)

	for _, inc := range included {

		to, err := inc.MarshalJSON()
		if err != nil {
			continue
		}

		replicaCount := gjson.Get(string(to), "spec.replicas")

		if replicaCount.Int() == 0 {
			continue
		}

		impactedObject := pincherv1alpha1.ImpactedObject{
			ResourceKey:   getResourceKey(inc),
			OriginalCount: int(replicaCount.Int()),
			Status:        "success",
		}

		request := &pkg.PatchRequest{
			Name:             inc.GetName(),
			Namespace:        inc.GetNamespace(),
			GroupVersionKind: inc.GroupVersionKind(),
			Patch:            fmt.Sprintf(fullPatch, 0, replicaAnnotation, replicaCount.Raw),
			PatchType:        string(types.JSONPatchType),
		}
		_, err = r.Kubectl.PatchResource(context.Background(), request)

		if err != nil {
			//log.Error(err, "action on object %s", requestObj)
			impactedObject.Status = "error"
			impactedObject.Message = err.Error()
		}

		impactedObjects = append(impactedObjects, impactedObject)
	}
	return impactedObjects, excludedObjects
}

func (r *ResourceActionImpl) UnHibernateActionFactory(hibernator *pincherv1alpha1.Hibernator) Execute {
	latestHistory := r.historyUtil.getLatestHistory(hibernator.Status.History)

	previousHibernatedObjects := make(map[string]int, 0)

	if latestHistory != nil {
		for _, impactedObject := range latestHistory.ImpactedObjects {
			previousHibernatedObjects[impactedObject.ResourceKey] = impactedObject.OriginalCount
		}
	}

	return func(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject) {
		impactedObjects := make([]pincherv1alpha1.ImpactedObject, 0)
		excludedObjects := make([]pincherv1alpha1.ExcludedObject, 0)

		for _, inc := range included {

			to, err := inc.MarshalJSON()
			if err != nil {
				continue
			}

			annotations := gjson.Get(string(to), "metadata.annotations")
			originalCount := annotations.Map()[replicaAnnotation].Str
			replicaCount, err := strconv.Atoi(originalCount)
			if len(originalCount) == 0 || err != nil {
				resourceKey := getResourceKey(inc)
				ok := false
				if replicaCount, ok = previousHibernatedObjects[resourceKey]; !ok {
					excludedObject := pincherv1alpha1.ExcludedObject{
						ResourceKey: getResourceKey(inc),
						Reason:      "error determining original count",
					}
					excludedObjects = append(excludedObjects, excludedObject)
					continue
				}
			}
			if replicaCount == 0 {
				continue
			}

			impactedObject := pincherv1alpha1.ImpactedObject{
				ResourceKey:   getResourceKey(inc),
				OriginalCount: replicaCount,
				Status:        "success",
			}

			request := &pkg.PatchRequest{
				Name:             inc.GetName(),
				Namespace:        inc.GetNamespace(),
				GroupVersionKind: inc.GroupVersionKind(),
				Patch:            fmt.Sprintf(fullPatch, replicaCount, replicaAnnotation, "0"),
				PatchType:        string(types.JSONPatchType),
			}
			_, err = r.Kubectl.PatchResource(context.Background(), request)

			if err != nil {
				//log.Error(err, "action on object %s", requestObj)
				impactedObject.Status = "error"
				impactedObject.Message = err.Error()
			}

			impactedObjects = append(impactedObjects, impactedObject)
		}
		return impactedObjects, excludedObjects
	}
}
