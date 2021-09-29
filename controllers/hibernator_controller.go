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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
)

const layout = "Jan 2, 2006 3:04pm"

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
	//_ = context.Background()
	log := r.Log.WithValues("hibernator", req.NamespacedName)

	// your logic here
	//r.Client.Get()
	hibernator := pincherv1alpha1.Hibernator{}
	err := r.Client.Get(context.Background(), req.NamespacedName, &hibernator)
	if err != nil {
		log.Error(err, "error while fetching hibernator")
		return ctrl.Result{}, err
	}
	log.Info("initiate processing")

	return r.process(hibernator)
}

func (r *HibernatorReconciler) process(hibernator pincherv1alpha1.Hibernator) (ctrl.Result, error) {
	log := r.Log.WithValues("hibernator", r.getNamespacedName(&hibernator))
	now := time.Now()

	diff, err := r.getPauseUntilDuration(&hibernator, now)
	if err != nil {
		log.Error(err, "continue processing as error parsing pause until %s", hibernator.Spec.PauseUntil.DateTime)
	} else if diff.Seconds() > 60 {
		return ctrl.Result{RequeueAfter: diff}, nil
	}

	if hibernator.Spec.Pause {
		return ctrl.Result{}, nil
		//return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	timeRangeWithZone := hibernator.Spec.When
	timeGap, shouldHibernate, err := timeRangeWithZone.NearestTimeGapInSeconds(now)
	//TODO: even if timeGap is error if reSyncInterval is set, it should still be able to work
	if err != nil {
		log.Error(err, "unable to parse the time interval")
		//hibernator.Status.Status = "Failed"
		//hibernator.Status.Message = err.Error()
		//TODO: should it return
		if hibernator.Spec.ReSyncInterval <= 0 {
			log.Error(err, "unable to parse the time interval and reSyncInterval is <= 0 hence aborting")
			//TODO: generate event
			//TODO: update status if not already failed
			return ctrl.Result{}, nil
		}
	}

	if hibernator.Spec.Hibernate {
		shouldHibernate = true
	}
	if hibernator.Spec.UnHibernate {
		shouldHibernate = false
	}

	requeueTime := r.getRequeueTimeDuration(timeGap, &hibernator)

	timeElapsedSinceLastRunInMinutes, hasPreviousRun := r.timeElapsedSinceLastRunInMinutes(&hibernator)
	if hasPreviousRun && timeElapsedSinceLastRunInMinutes <= 1 {
		log.Info("skipping reconciliation as time elapsed since last run is less than 1 min")
		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	//TODO: handle the case of resync, in case of hibernating
	finalHibernator := &hibernator
	updated := false
	if !shouldHibernate /*&& hibernator.Status.IsHibernating*/ {
		finalHibernator = r.unhibernate(&hibernator)
		updated = true
	} else if shouldHibernate /*&& !hibernator.Status.IsHibernating*/ {
		finalHibernator = r.hibernate(&hibernator)
		updated = true
	} else {
		log.Info("didnt hibernate or unhibernate - shouldHibernate: %t, timegap: %d, hibernating: %t", shouldHibernate, timeGap, hibernator.Status.IsHibernating)
	}

	if updated {
		err = r.Client.Update(context.Background(), finalHibernator)
		if err != nil {
			log.Error(err, "error while updating hibernator %v")
			return ctrl.Result{}, err
		}
	}

	log.Info("end processing, processing parameter - start time: %v, timegap: %d, requeueTime:  %s", now, timeGap, requeueTime)
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

//TODO: handle sync
func (r *HibernatorReconciler) unhibernate(hibernator *pincherv1alpha1.Hibernator) *pincherv1alpha1.Hibernator {
	log := r.Log.WithValues("hibernator", r.getNamespacedName(hibernator))

	log.Info("initiating unhibernate")
	//TODO: remove this check as it would be difficult to unhibernate all together
	latestHistory := r.getLatestHistory(hibernator.Status.History)
	if latestHistory.Hibernate == false {
		return hibernator
	}
	//TODO: change logic to get all objects and process those whose replicaCount is 0 and have been hibernated by us earlier
	hibernator.Status.IsHibernating = false

	var impactedObjects []pincherv1alpha1.ImpactedObject
	for _, impactedObject := range latestHistory.ImpactedObjects {

		impactedObject = pincherv1alpha1.ImpactedObject{
			ResourceKey:   impactedObject.ResourceKey,
			OriginalCount: impactedObject.OriginalCount,
			Status:        "success",
		}

		//objectPatch := fmt.Sprintf(patch, impactedObject.OriginalCount)
		namespace, group, version, kind, name := r.componentsOfResourceKey(impactedObject.ResourceKey)
		request := &pkg.PatchRequest{
			Name:      name,
			Namespace: namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   group,
				Version: version,
				Kind:    kind,
			},
			Patch:     fmt.Sprintf(patch, impactedObject.OriginalCount),
			PatchType: string(types.JSONPatchType),
		}
		_, err := r.Kubectl.PatchResource(context.Background(), request)
		if err != nil {
			log.Error(err, "error while deleting object %v\n", *request)
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
	log.Info("ending unhibernate")

	return hibernator
}

//TODO: handle sync
func (r *HibernatorReconciler) hibernate(hibernator *pincherv1alpha1.Hibernator) *pincherv1alpha1.Hibernator {
	log := r.Log.WithValues("hibernator", r.getNamespacedName(hibernator))
	log.Info("starting hibernate")
	hibernator.Status.IsHibernating = true
	var impactedObjects []pincherv1alpha1.ImpactedObject
	var excludedObjects []pincherv1alpha1.ExcludedObject
	patchZero := fmt.Sprintf(patch, 0)
	for _, rule := range hibernator.Spec.Rules {
		inclusions := r.getMatchingObjects(rule.Inclusions)
		exclusions := r.getMatchingObjects(rule.Exclusions)
		included, excluded := r.getIncludedExcludedObjects(inclusions, exclusions)
		//hpaTargetObjectPairs := r.fetchHPATargetObjectPairForObjects(included, mapping)

		for _, ex := range excluded {
			excludedObject := pincherv1alpha1.ExcludedObject{
				ResourceKey: r.getResourceKey(ex),
			}
			excludedObjects = append(excludedObjects, excludedObject)
		}

		for _, inc := range included {
			to, err := inc.MarshalJSON()
			if err != nil {
				continue
			}

			replicaCount := gjson.Get(string(to), "spec.replicas")

			if rule.Action == "sleep" && replicaCount.Int() == 0 {
				excludedObject := pincherv1alpha1.ExcludedObject{
					ResourceKey: r.getResourceKey(inc),
					Reason:      "existing replica count is zero",
				}
				excludedObjects = append(excludedObjects, excludedObject)
				continue
			}

			impactedObject := pincherv1alpha1.ImpactedObject{
				ResourceKey:   r.getResourceKey(inc),
				OriginalCount: replicaCount.Int(),
				Status:        "success",
			}

			var requestObj string
			if rule.Action == "sleep" {
				request := &pkg.PatchRequest{
					Name:             inc.GetName(),
					Namespace:        inc.GetNamespace(),
					GroupVersionKind: inc.GroupVersionKind(),
					Patch:            patchZero,
					PatchType:        string(types.JSONPatchType),
				}
				_, err = r.Kubectl.PatchResource(context.Background(), request)
				requestObj = fmt.Sprintf("%v", request)
			} else {
				request := &pkg.DeleteRequest{
					Name:             inc.GetName(),
					Namespace:        inc.GetNamespace(),
					GroupVersionKind: inc.GroupVersionKind(),
					Force:            pointer.BoolPtr(true),
				}
				_, err = r.Kubectl.DeleteResource(context.Background(), request)
				requestObj = fmt.Sprintf("%v", request)
			}

			if err != nil {
				log.Error(err, "action on object %s", requestObj)
				impactedObject.Status = "error"
				impactedObject.Message = err.Error()
			}

			impactedObjects = append(impactedObjects, impactedObject)
		}
	}
	if impactedObjects == nil {
		impactedObjects = make([]pincherv1alpha1.ImpactedObject, 0)
	}
	if excludedObjects == nil {
		excludedObjects = make([]pincherv1alpha1.ExcludedObject, 0)
	}

	history := pincherv1alpha1.RevisionHistory{
		Time:            metav1.Time{Time: time.Now()},
		ID:              r.getNewRevisionID(hibernator.Status.History),
		Hibernate:       true,
		ImpactedObjects: impactedObjects,
		ExcludedObjects: excludedObjects,
	}
	hibernator.Status.History = r.addToHistory(history, hibernator.Status.History)

	log.Info("ending hibernate")
	return hibernator
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
