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
	"encoding/json"
	"fmt"
	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"strconv"
	"strings"
	"time"
)

type Execute func(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)

type ResourceAction interface {
	DeleteAction(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)
	ScaleActionFactory(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) Execute
	ScaleResourceActionFactory(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) Execute
	ResetScaleActionFactory(hibernator *pincherv1alpha1.Hibernator) Execute
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
			impactedObject.Status = "error"
			impactedObject.Message = err.Error()
		}

		impactedObjects = append(impactedObjects, impactedObject)
	}
	return impactedObjects, excludedObjects
}

func (r *ResourceActionImpl) ScaleActionFactory(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) Execute {
	fmt.Printf("entering ScaleActionFactory %s \n", time.Now().Format(time.RFC1123Z))
	targetReplicaCount := 0
	if hibernator.Spec.TargetReplicas != nil && len(*hibernator.Spec.TargetReplicas) > timeGap.MatchedIndex {
		targetReplicaCount = (*hibernator.Spec.TargetReplicas)[timeGap.MatchedIndex]
	} else if hibernator.Spec.TargetReplicas != nil && len(*hibernator.Spec.TargetReplicas) != 0 && len(*hibernator.Spec.TargetReplicas) <= timeGap.MatchedIndex {
		targetReplicaCount = (*hibernator.Spec.TargetReplicas)[len(*hibernator.Spec.TargetReplicas)-1]
	}
	if hibernator.Spec.Action == pincherv1alpha1.Hibernate || hibernator.Spec.Action == pincherv1alpha1.Sleep {
		targetReplicaCount = 0
	}

	fmt.Printf("entering ScaleActionFactory %d \n", targetReplicaCount)
	return func(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject) {

		impactedObjects := make([]pincherv1alpha1.ImpactedObject, 0)
		excludedObjects := make([]pincherv1alpha1.ExcludedObject, 0)

		for _, inc := range included {

			to, err := inc.MarshalJSON()
			if err != nil {
				continue
			}

			replicaCount := gjson.Get(string(to), "spec.replicas")

			if int(replicaCount.Int()) == targetReplicaCount {
				continue
			}

			patch := fmt.Sprintf(replicaPatch, targetReplicaCount)
			if !r.hasReplicaAnnotation(inc) {
				fmt.Println("annotation missing in ScaleActionFactory")
				patch = fmt.Sprintf(replicaAndAnnotationPatch, targetReplicaCount, replicaAnnotation, replicaCount.Raw)
			} else {
				fmt.Println("annotation found in ScaleActionFactory")
			}

			if inc.GetKind() == "HorizontalPodAutoscaler" {
				replicaCount = gjson.Get(string(to), "spec.minReplicas")

				if int(replicaCount.Int()) == targetReplicaCount {
					continue
				}
				patch = fmt.Sprintf(minReplicaPatch, targetReplicaCount)
				if !r.hasReplicaAnnotation(inc) {
					patch = fmt.Sprintf(minReplicaAndAnnotationPatch, targetReplicaCount, replicaAnnotation, replicaCount.Raw)
				}
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
				Patch:            patch,
				PatchType:        string(types.JSONPatchType),
			}
			_, err = r.Kubectl.PatchResource(context.Background(), request)

			if err != nil {
				impactedObject.Status = "error"
				impactedObject.Message = err.Error()
			}

			impactedObjects = append(impactedObjects, impactedObject)
		}
		return impactedObjects, excludedObjects
	}
}

func (r *ResourceActionImpl) ResetScaleActionFactory(hibernator *pincherv1alpha1.Hibernator) Execute {
	fmt.Printf("entering ResetScaleActionFactory %s \n", time.Now().Format(time.RFC1123Z))
	previousHibernatedObjects := make(map[string]int, 0)
	latestHistory := r.historyUtil.getLatestHistory(hibernator.Status.History)
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

			currentReplicaCount := gjson.Get(string(to), "spec.replicas")

			replicaCount, err := r.getOriginalReplicaCount(inc)
			if err != nil {
				continue
			}

			if replicaCount == 0 && (hibernator.Spec.Action == pincherv1alpha1.Hibernate || hibernator.Spec.Action == pincherv1alpha1.Sleep) {
				continue
			}

			if replicaCount == int(currentReplicaCount.Int()) {
				continue
			}

			if err != nil {
				excludedObject := pincherv1alpha1.ExcludedObject{
					ResourceKey: getResourceKey(inc),
					Reason:      "error determining original count",
				}
				excludedObjects = append(excludedObjects, excludedObject)
				continue
			}

			impactedObject := pincherv1alpha1.ImpactedObject{
				ResourceKey:   getResourceKey(inc),
				OriginalCount: replicaCount,
				Status:        "success",
			}

			patch := fmt.Sprintf(replicaPatch, replicaCount)

			if inc.GetKind() == "HorizontalPodAutoscaler" {
				patch = fmt.Sprintf(minReplicaPatch, replicaCount)
			}

			request := &pkg.PatchRequest{
				Name:             inc.GetName(),
				Namespace:        inc.GetNamespace(),
				GroupVersionKind: inc.GroupVersionKind(),
				Patch:            patch,
				PatchType:        string(types.JSONPatchType),
			}
			_, err = r.Kubectl.PatchResource(context.Background(), request)

			if err != nil {
				impactedObject.Status = "error"
				impactedObject.Message = err.Error()
			}

			impactedObjects = append(impactedObjects, impactedObject)
		}
		return impactedObjects, excludedObjects
	}
}

func (r *ResourceActionImpl) getOriginalReplicaCount(res unstructured.Unstructured /*, previousHibernatedObjects map[string]int*/) (int, error) {
	to, err := res.MarshalJSON()
	if err != nil {
		return 0, err
	}
	annotations := gjson.Get(string(to), "metadata.annotations")
	originalCount := annotations.Map()[replicaAnnotation].Str
	replicaCount, err := strconv.Atoi(originalCount)
	//if len(originalCount) == 0 || err != nil {
	//	resourceKey := getResourceKey(res)
	//	ok := false
	//	if replicaCount, ok = previousHibernatedObjects[resourceKey]; !ok {
	//		return 0, nil
	//	}
	//}
	return replicaCount, err
}

func (r *ResourceActionImpl) hasReplicaAnnotation(res unstructured.Unstructured) bool {
	to, err := res.MarshalJSON()
	if err != nil {
		return false
	}
	annotations := gjson.Get(string(to), "metadata.annotations")
	originalCount := annotations.Map()[replicaAnnotation].Str
	return len(originalCount) != 0
}

func (r *ResourceActionImpl) ScaleResourceActionFactory(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) Execute {
	fmt.Printf("entering ScaleResourceActionFactory %s \n", time.Now().Format(time.RFC1123Z))
	var resourceRequirements map[string]v1.ResourceRequirements
	if hibernator.Spec.TargetResources != nil && len(*hibernator.Spec.TargetResources) > timeGap.MatchedIndex {
		resourceRequirements = (*hibernator.Spec.TargetResources)[timeGap.MatchedIndex]
	} else if len(*hibernator.Spec.TargetResources) != 0 && len(*hibernator.Spec.TargetReplicas) <= timeGap.MatchedIndex {
		resourceRequirements = (*hibernator.Spec.TargetResources)[len(*hibernator.Spec.TargetResources)-1]
	}

	return func(included []unstructured.Unstructured) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject) {

		impactedObjects := make([]pincherv1alpha1.ImpactedObject, 0)
		excludedObjects := make([]pincherv1alpha1.ExcludedObject, 0)

		if resourceRequirements == nil {
			return impactedObjects, excludedObjects
		}
		resourceRequirementsAsSpec := make([]map[string]interface{}, 0)
		for k, v := range resourceRequirements {
			resourceRequirementsAsSpec = append(resourceRequirementsAsSpec, map[string]interface{}{"name": k, "resources": v})
		}
		var err error

		for _, inc := range included {
			if !strings.EqualFold(inc.GetKind(), "Pod") {
				continue
			}
			var pod v1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(inc.UnstructuredContent(), &pod)
			if err != nil {
				continue
			}
			needUpdate := false
			for _, container := range pod.Spec.Containers {
				rr, ok := resourceRequirements[container.Name]
				if !ok {
					continue
				}
				cpuRequestEqual := container.Resources.Requests.Cpu().Cmp(*rr.Requests.Cpu()) == 0
				memoryRequestEqual := container.Resources.Requests.Memory().Cmp(*rr.Requests.Memory()) == 0
				cpuLimitEqual := container.Resources.Limits.Cpu().Cmp(*rr.Limits.Cpu()) == 0
				memoryLimitEqual := container.Resources.Limits.Memory().Cmp(*rr.Limits.Memory()) == 0
				allEqual := cpuRequestEqual && memoryRequestEqual && cpuLimitEqual && memoryLimitEqual
				if allEqual {
					continue
				}
				needUpdate = true
			}
			if !needUpdate {
				continue
			}
			resourcePatch := map[string]map[string][]map[string]interface{}{"spec": {"containers": resourceRequirementsAsSpec}}

			patch := ""
			if requirements, err := json.Marshal(resourcePatch); err == nil {
				patch = string(requirements)
			}
			if len(patch) == 0 {
				continue
			}

			impactedObject := pincherv1alpha1.ImpactedObject{
				ResourceKey: getResourceKey(inc),
				Status:      "success",
			}

			request := &pkg.PatchRequest{
				Name:             inc.GetName(),
				Namespace:        inc.GetNamespace(),
				GroupVersionKind: inc.GroupVersionKind(),
				Patch:            patch,
				PatchType:        string(types.StrategicMergePatchType),
			}
			_, err = r.Kubectl.PatchResource(context.Background(), request)

			if err != nil {
				impactedObject.Status = "error"
				impactedObject.Message = err.Error()
			}

			impactedObjects = append(impactedObjects, impactedObject)
		}
		return impactedObjects, excludedObjects
	}
}
