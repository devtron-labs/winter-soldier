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
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
	"math"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
)

// HibernatorReconciler reconciles a Hibernator object
type HibernatorReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	kubectl pkg.KubectlCmd
	mapper  *pkg.Mapper
}

const patch = `[{"op": "replace", "path": "/spec/replicas", "value":%d}]`

// +kubebuilder:rbac:groups=pincher.devtron.ai,resources=hibernators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pincher.devtron.ai,resources=hibernators/status,verbs=get;update;patch

func (r *HibernatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("hibernator", req.NamespacedName)

	// your logic here
	//r.Client.Get()
	hibernator := pincherv1alpha1.Hibernator{}
	r.Client.Get(context.Background(), req.NamespacedName, &hibernator)
	now := time.Now()
	timeRangeWithZone := hibernator.Spec.TimeRangesWithZone
	timeGap, inRange, err := timeRangeWithZone.NearestTimeGap(now)
	if err != nil {
		hibernator.Status.Status = "Failed"
		hibernator.Status.Message = err.Error()
	}
	if !inRange && hibernator.Status.IsHibernating {
		hibernator.Status.IsHibernating = false
		r.Client.Update(context.Background(), &hibernator)
		return ctrl.Result{
			RequeueAfter: time.Duration(timeGap) * time.Minute,
		}, nil
	} else if inRange && !hibernator.Status.IsHibernating {
		finalHibernator, err := r.hibernate(hibernator)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.Client.Update(context.Background(), finalHibernator)
		//hibernator.Status.IsHibernating = true
		//TODO: check if unhibernated then hibernate
		return ctrl.Result{
			RequeueAfter: time.Duration(timeGap) * time.Minute,
		}, nil
	}

	return ctrl.Result{}, nil
}

func (r *HibernatorReconciler) hibernate(hibernator pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, error) {
	hibernator.Status.IsHibernating = true
	//TODO: check if hibernated then unhibernate
	factory := pkg.NewFactory(r.mapper)
	mapping, err := factory.MappingFor("hpa")
	if err != nil {
		fmt.Println("error fetching mapping for hpa")
		return nil, err
	}
	var impactedObjects []pincherv1alpha1.ImpactedObject
	var excludedObjects []pincherv1alpha1.ExcludedObject
	patchZero := fmt.Sprintf(patch, 0)
	for _, rule := range hibernator.Spec.Rules {
		inclusions := r.getMatchingObjects(rule.Inclusions)
		exclusions := r.getMatchingObjects(rule.Exclusions)
		included, excluded := r.getFinalIncludedObjects(inclusions, exclusions)
		hpaTargetObjectPairs := r.fetchHPATargetObjectPairForObjects(included, mapping)
		//TODO: delete HPA, change target object size to 0 and store original size and hpa manifest, store hpa
		for _, ex := range excluded {
			excludedObject := pincherv1alpha1.ExcludedObject{
				Group:     ex.GroupVersionKind().Group,
				Version:   ex.GroupVersionKind().Version,
				Kind:      ex.GroupVersionKind().Kind,
				Name:      ex.GetName(),
				Namespace: ex.GetNamespace(),
			}
			excludedObjects = append(excludedObjects, excludedObject)
		}
		for _, pair := range hpaTargetObjectPairs {
			to, err := pair.TargetObject.MarshalJSON()
			if err != nil {
				continue
			}
			res := gjson.Get(string(to), "spec.replicas")
			impactedObject := pincherv1alpha1.ImpactedObject{
				Group:         pair.TargetObject.GroupVersionKind().Group,
				Version:       pair.TargetObject.GroupVersionKind().Version,
				Kind:          pair.TargetObject.GroupVersionKind().Kind,
				Name:          pair.TargetObject.GetName(),
				Namespace:     pair.TargetObject.GetNamespace(),
				OriginalCount: res.Int(),
			}
			if pair.HPA != nil {
				request := &pkg.DeleteRequest{
					Name:             pair.HPA.GetName(),
					Namespace:        pair.HPA.GetNamespace(),
					GroupVersionKind: pair.HPA.GroupVersionKind(),
					Force:            pointer.BoolPtr(true),
				}
				resp, err := r.kubectl.DeleteResource(context.Background(), request)
				if err != nil {
					//TODO: set event
					fmt.Printf("error err %v while deleting object %v\n", err, *request)
					impactedObject.Status = "error"
					impactedObject.Message = err.Error()
					impactedObjects = append(impactedObjects, impactedObject)
					continue
				}
				deletedObject, err := resp.Manifest.MarshalJSON()
				if err != nil {
					fmt.Printf("error err %v while deleting object %v\n", err, *request)
					impactedObject.Status = "error"
					impactedObject.Message = err.Error()
					impactedObjects = append(impactedObjects, impactedObject)
					continue
				}
				impactedObject.RelatedDeletedObject = string(deletedObject)
			}
			request := &pkg.PatchRequest{
				Name:             pair.TargetObject.GetName(),
				Namespace:        pair.TargetObject.GetNamespace(),
				GroupVersionKind: pair.TargetObject.GroupVersionKind(),
				Patch:            patchZero,
				PatchType:        "application/json-patch+json",
			}
			_, err = r.kubectl.PatchResource(context.Background(), request)
			if err != nil {
				fmt.Printf("error err %v while deleting object %v\n", err, *request)
				impactedObject.Status = "error"
				impactedObject.Message = err.Error()
				impactedObjects = append(impactedObjects, impactedObject)
				continue
			}
			impactedObject.Status = "success"
			impactedObjects = append(impactedObjects, impactedObject)
		}
	}

	history := pincherv1alpha1.RevisionHistory{
		Time:            metav1.Time{Time: time.Now()},
		ID:              r.getRevisionID(hibernator.Status.History),
		ImpactedObjects: impactedObjects,
		ExcludedObjects: excludedObjects,
	}
	hibernator.Status.History = r.addToHistory(history, hibernator.Status.History)
	return &hibernator, nil
}

func (r *HibernatorReconciler) addToHistory(history pincherv1alpha1.RevisionHistory, revisionHistories []pincherv1alpha1.RevisionHistory) []pincherv1alpha1.RevisionHistory {
	if len(revisionHistories) < 10 {
		revisionHistories = append(revisionHistories, history)
		return revisionHistories
	}
	minID := int64(math.MaxInt64)
	for _, history := range revisionHistories {
		if history.ID < minID {
			minID = history.ID
		}
	}
	var finalHistories []pincherv1alpha1.RevisionHistory
	for _, history := range revisionHistories {
		if history.ID != minID {
			finalHistories = append(finalHistories, history)
		}
	}
	finalHistories = append(finalHistories, history)
	return finalHistories

}

func (r *HibernatorReconciler) getRevisionID(revisionHistories []pincherv1alpha1.RevisionHistory) int64 {
	maxID := int64(-1)
	for _, history := range revisionHistories {
		if history.ID > maxID {
			maxID = history.ID
		}
	}
	return maxID + 1
}

func (r *HibernatorReconciler) fetchHPATargetObjectPairForObjects(included []unstructured.Unstructured, mapping *meta.RESTMapping) []TargetObjectHPAPair {
	var hpas []unstructured.Unstructured
	namespaces := map[string]bool{}

	for _, inc := range included {
		namespaces[inc.GetNamespace()] = true
	}
	for k, _ := range namespaces {
		request := &pkg.ListRequest{
			Namespace:            k,
			GroupVersionResource: mapping.Resource,
			ListOptions:          metav1.ListOptions{},
		}
		namespaceHPA, err := r.kubectl.ListResources(context.Background(), request)
		if err != nil {
			continue
		}
		hpas = append(hpas, namespaceHPA.Manifests...)
	}
	return r.createHPAAndTargetObjectMapping(hpas, included)

}

func (r *HibernatorReconciler) createHPAAndTargetObjectMapping(hpas, targetObjects []unstructured.Unstructured) []TargetObjectHPAPair {
	targetObjectsKeys := map[string]*unstructured.Unstructured{}
	matchedHPA := map[string]bool{}
	for i := 0; i < len(targetObjects); i++ {
		inc := targetObjects[i]
		key := getKeyForHPATargetObjectMapping(inc.GetNamespace(), inc.GetAPIVersion(), inc.GetKind(), inc.GetName())
		targetObjectsKeys[key] = &inc
	}
	var targetObjectHPAPairs []TargetObjectHPAPair
	for _, manifest := range hpas {
		targetKey := r.getHPATargetRefKey(manifest)
		if to, ok := targetObjectsKeys[targetKey]; ok {
			matchedHPA[targetKey] = true
			toHPA := TargetObjectHPAPair{
				TargetObject: to,
				HPA:          &manifest,
			}
			targetObjectHPAPairs = append(targetObjectHPAPairs, toHPA)
		}
	}
	for k, v := range targetObjectsKeys {
		if !matchedHPA[k] {
			toHPA := TargetObjectHPAPair{
				TargetObject: v,
			}
			targetObjectHPAPairs = append(targetObjectHPAPairs, toHPA)
		}
	}
	return targetObjectHPAPairs
}

func getKeyForHPATargetObjectMapping(namespace, apiVersion, kind, name string) string {
	return fmt.Sprintf("/%s/%s/%s/%s", namespace, apiVersion, kind, name)
}

func (r *HibernatorReconciler) getHPATargetRefKey(hpa unstructured.Unstructured) string {
	specInterface := hpa.Object["spec"]
	spec := specInterface.(map[string]interface{})
	scaleTargetRefInterface := spec["scaleTargetRef"]
	scaleTargetRef := scaleTargetRefInterface.(map[string]interface{})
	apiVersion, kind, name := "", "", ""
	if av, ok := scaleTargetRef["apiVersion"]; ok {
		apiVersion = av.(string)
	}
	if k, ok := scaleTargetRef["kind"]; ok {
		kind = k.(string)
	}
	if n, ok := scaleTargetRef["name"]; ok {
		name = n.(string)
	}
	namespace := hpa.GetNamespace()
	return getKeyForHPATargetObjectMapping(namespace, apiVersion, kind, name)
}

func (r *HibernatorReconciler) getMatchingObjects(selectors []pincherv1alpha1.Selector) []unstructured.Unstructured {
	var allMatches []unstructured.Unstructured
	for _, selector := range selectors {
		var err error
		var matches []unstructured.Unstructured
		if len(selector.FieldSelector) != 0 {
			matches, err = r.handleFieldSelector(selector)
		} else if len(selector.Labels) != 0 {
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

func (r *HibernatorReconciler) getFinalIncludedObjects(inclusions, exclusions []unstructured.Unstructured) (included []unstructured.Unstructured, excluded []unstructured.Unstructured) {
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

func (r *HibernatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pincherv1alpha1.Hibernator{}).
		Complete(r)
}

type TargetObjectHPAPair struct {
	TargetObject *unstructured.Unstructured
	HPA          *unstructured.Unstructured
}
