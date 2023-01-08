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
	"github.com/devtron-labs/winter-soldier/pkg"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
)

const (
	layout                    = "Jan 2, 2006 3:04pm"
	replicaPatch              = `[{"op": "replace", "path": "/spec/replicas", "value":%d}]`
	replicaAndAnnotationPatch = `[{"op": "replace", "path": "/spec/replicas", "value":%d}, {"op": "add", "path": "/metadata/annotations", "value": {"%s":"%s"}}]`
	replicaAnnotation         = `hibernator.devtron.ai/replicas`
)

// HibernatorReconciler reconciles a Hibernator object
type HibernatorReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	Kubectl          pkg.KubectlCmd
	Mapper           *pkg.Mapper
	HibernatorAction HibernatorAction
	TimeUtil         TimeUtil
}

// +kubebuilder:rbac:groups=pincher.devtron.ai,resources=hibernators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pincher.devtron.ai,resources=hibernators/status,verbs=get;update;patch

func (r *HibernatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = context.Background()
	log := r.Log.WithValues("hibernator", req.NamespacedName)

	// your logic here
	//r.Client.Get()
	hibernator := pincherv1alpha1.Hibernator{}
	err := r.Client.Get(ctx, req.NamespacedName, &hibernator)
	if err != nil {
		log.Error(err, "error while fetching hibernator")
		return ctrl.Result{}, err
	}
	log.Info("initiate processing")

	return r.process(hibernator)
}

func (r *HibernatorReconciler) process(hibernator pincherv1alpha1.Hibernator) (ctrl.Result, error) {
	log := r.Log.WithValues("hibernator", getNamespacedName(&hibernator))
	now := time.Now()

	diff, err := r.TimeUtil.getPauseUntilDuration(&hibernator, now)
	if err != nil {
		log.Error(err, "continue processing as error parsing pause until %s", hibernator.Spec.PauseUntil.DateTime)
	} else if diff.Seconds() > pincherv1alpha1.MinReSyncIntervalInSeconds {
		return ctrl.Result{RequeueAfter: diff}, nil
	}

	if hibernator.Spec.Pause {
		//Re-evaluate when spec changes
		return ctrl.Result{}, nil
	}

	//TODO: calculation may be different for delete
	timeRangeWithZone := hibernator.Spec.When
	nearestTimeGap, err := timeRangeWithZone.NearestTimeGapInSeconds(now)
	//even if timeGap is error if reSyncInterval is set, it should still be able to work in case of delete
	if err != nil {
		log.Error(err, "unable to parse the time interval")
		if hibernator.Spec.ReSyncInterval <= 0 {
			log.Error(err, "unable to parse the time interval and reSyncInterval is <= 0 hence aborting")
			//TODO: generate event
			//TODO: update status if not already failed
			//No point in retrying as it will fail again
			return ctrl.Result{}, nil
		}
	}

	requeueTime := r.TimeUtil.getRequeueTimeDuration(nearestTimeGap.TimeGapInSeconds, &hibernator)

	timeElapsedSinceLastRunInSeconds, hasPreviousRun := r.TimeUtil.timeElapsedSinceLastRunInSeconds(&hibernator)
	if hasPreviousRun && timeElapsedSinceLastRunInSeconds <= pincherv1alpha1.MinReSyncIntervalInSeconds {
		log.Info("skipping reconciliation as time elapsed since last run is less than 1 min")
		return ctrl.Result{RequeueAfter: time.Duration(pincherv1alpha1.MinReSyncIntervalInSeconds) * time.Second}, nil
	}

	finalHibernator := &hibernator
	updated := false
	if hibernator.Spec.Action == pincherv1alpha1.Delete {
		finalHibernator, updated = r.HibernatorAction.delete(&hibernator)
	} else if hibernator.Spec.Action == pincherv1alpha1.Hibernate {
		finalHibernator, updated = r.HibernatorAction.hibernate(&hibernator, nearestTimeGap)
	} else if hibernator.Spec.Action == pincherv1alpha1.Scale {
		finalHibernator, updated = r.HibernatorAction.scale(&hibernator, nearestTimeGap)
	} else {
		log.Info("didnt hibernate or unHibernate -", "action", nearestTimeGap.WithinRange, "timegap", nearestTimeGap.TimeGapInSeconds, "isHibernating", hibernator.Status.IsHibernating)
	}

	if updated {
		err = r.Client.Update(context.Background(), finalHibernator)
		if err != nil {
			log.Error(err, "error while updating hibernator %v")
			return ctrl.Result{}, err
		}
	}

	//log.Info("end processing, processing parameter - start time: %v, timegap: %d, requeueTime:  %s", now, timeGap, requeueTime)
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

func (r *HibernatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pincherv1alpha1.Hibernator{}).
		Complete(r)
}

//type TargetObjectHPAPair struct {
//	TargetObject *unstructured.Unstructured
//	HPA          *unstructured.Unstructured
//}
