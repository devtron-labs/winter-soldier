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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	Kubectl pkg.KubectlCmd
	Mapper  *pkg.Mapper
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
	requeTime := time.Duration(timeGap) * time.Minute
	if err != nil {
		hibernator.Status.Status = "Failed"
		hibernator.Status.Message = err.Error()
	}
	latestHistory := r.getLatestHistory(hibernator.Status.History)
	timeElapsedSinceLastRun := time.Now().Sub(latestHistory.Time.Time)
	if timeElapsedSinceLastRun.Minutes() <= 1 {
		return ctrl.Result{
			RequeueAfter: requeTime,
		}, nil
	}
	if (!inRange || (inRange && timeGap <= 1)) && hibernator.Status.IsHibernating {
		finalHibernator, err := r.unhibernate(hibernator)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.Client.Update(context.Background(), finalHibernator)
	} else if (inRange || (!inRange && timeGap <= 1)) && !hibernator.Status.IsHibernating {
		finalHibernator, err := r.hibernate(hibernator)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.Client.Update(context.Background(), finalHibernator)
	}

	return ctrl.Result{
		RequeueAfter: requeTime,
	}, nil
}

func (r *HibernatorReconciler) unhibernate(hibernator pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, error) {
	hibernator.Status.IsHibernating = false
	latestHistory := r.getLatestHistory(hibernator.Status.History)
	if latestHistory.Hibernate == false {
		return &hibernator, nil
	}
	var impactedObjects []pincherv1alpha1.ImpactedObject
	for _, impactedObject := range latestHistory.ImpactedObjects {
		impactedObject = pincherv1alpha1.ImpactedObject{
			Group:         impactedObject.Group,
			Version:       impactedObject.Version,
			Kind:          impactedObject.Kind,
			Name:          impactedObject.Name,
			Namespace:     impactedObject.Namespace,
			OriginalCount: impactedObject.OriginalCount,
			Status:        "success",
		}
		objectPatch := fmt.Sprintf(patch, impactedObject.OriginalCount)
		request := &pkg.PatchRequest{
			Name:      impactedObject.Name,
			Namespace: impactedObject.Namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   impactedObject.Group,
				Version: impactedObject.Version,
				Kind:    impactedObject.Kind,
			},
			Patch:     objectPatch,
			PatchType: string(types.JSONPatchType),
		}
		_, err := r.Kubectl.PatchResource(context.Background(), request)
		if err != nil {
			fmt.Printf("error err %v while deleting object %v\n", err, *request)
			impactedObject.Status = "error"
			impactedObject.Message = err.Error()
		}
		impactedObjects = append(impactedObjects, impactedObject)
	}
	history := pincherv1alpha1.RevisionHistory{
		Time:            metav1.Time{Time: time.Now()},
		ID:              r.getNewRevisionID(hibernator.Status.History),
		Hibernate:       false,
		ImpactedObjects: impactedObjects,
		ExcludedObjects: []pincherv1alpha1.ExcludedObject{},
	}
	hibernator.Status.History = r.addToHistory(history, hibernator.Status.History)
	return &hibernator, nil
}

func (r *HibernatorReconciler) hibernate(hibernator pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, error) {
	hibernator.Status.IsHibernating = true
	var impactedObjects []pincherv1alpha1.ImpactedObject
	var excludedObjects []pincherv1alpha1.ExcludedObject
	patchZero := fmt.Sprintf(patch, 0)
	for _, rule := range hibernator.Spec.Rules {
		inclusions := r.getMatchingObjects(rule.Inclusions)
		exclusions := r.getMatchingObjects(rule.Exclusions)
		included, excluded := r.getFinalIncludedObjects(inclusions, exclusions)
		//hpaTargetObjectPairs := r.fetchHPATargetObjectPairForObjects(included, mapping)
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
		for _, inc := range included {
			to, err := inc.MarshalJSON()
			if err != nil {
				continue
			}
			res := gjson.Get(string(to), "spec.replicas")
			impactedObject := pincherv1alpha1.ImpactedObject{
				Group:         inc.GroupVersionKind().Group,
				Version:       inc.GroupVersionKind().Version,
				Kind:          inc.GroupVersionKind().Kind,
				Name:          inc.GetName(),
				Namespace:     inc.GetNamespace(),
				OriginalCount: res.Int(),
				Status:        "success",
			}
			if rule.Action == "sleep" {
				request := &pkg.PatchRequest{
					Name:             inc.GetName(),
					Namespace:        inc.GetNamespace(),
					GroupVersionKind: inc.GroupVersionKind(),
					Patch:            patchZero,
					PatchType:        string(types.JSONPatchType),
				}
				_, err = r.Kubectl.PatchResource(context.Background(), request)
				if err != nil {
					fmt.Printf("error err %v while deleting object %v\n", err, *request)
					impactedObject.Status = "error"
					impactedObject.Message = err.Error()
					impactedObjects = append(impactedObjects, impactedObject)
					continue
				}
			} else {
				request := &pkg.DeleteRequest{
					Name:             inc.GetName(),
					Namespace:        inc.GetNamespace(),
					GroupVersionKind: inc.GroupVersionKind(),
					Force:            pointer.BoolPtr(true),
				}
				resp, err := r.Kubectl.DeleteResource(context.Background(), request)
				if err != nil {
					fmt.Printf("error err %v while deleting object %v\n", err, *request)
					impactedObject.Status = "error"
					impactedObject.Message = err.Error()
					impactedObjects = append(impactedObjects, impactedObject)
					continue
				}
				j, _ := resp.Manifest.MarshalJSON()
				impactedObject.RelatedDeletedObject = string(j)
			}

			impactedObject.Status = "success"
			impactedObjects = append(impactedObjects, impactedObject)
		}
	}

	history := pincherv1alpha1.RevisionHistory{
		Time:            metav1.Time{Time: time.Now()},
		ID:              r.getNewRevisionID(hibernator.Status.History),
		Hibernate:       true,
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

func (r *HibernatorReconciler) getLatestHistory(revisionHistories []pincherv1alpha1.RevisionHistory) pincherv1alpha1.RevisionHistory {
	maxID := int64(-1)
	for _, history := range revisionHistories {
		if history.ID > maxID {
			maxID = history.ID
		}
	}
	var latestHistory pincherv1alpha1.RevisionHistory
	for _, history := range revisionHistories {
		if history.ID == maxID {
			latestHistory = history
		}
	}
	return latestHistory
}

func (r *HibernatorReconciler) getNewRevisionID(revisionHistories []pincherv1alpha1.RevisionHistory) int64 {
	maxID := int64(-1)
	for _, history := range revisionHistories {
		if history.ID > maxID {
			maxID = history.ID
		}
	}
	return maxID + 1
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
