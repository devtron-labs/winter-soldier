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
	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type HibernatorAction interface {
	hibernate(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) (*pincherv1alpha1.Hibernator, bool)
	delete(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool)
	scale(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) (*pincherv1alpha1.Hibernator, bool)
	executeRules(hibernator *pincherv1alpha1.Hibernator, execute Execute, reSync bool) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)
}

func NewHibernatorActionImpl(kubectl pkg.KubectlCmd, historyUtil History, resourceAction ResourceAction, resourceSelector ResourceSelector, log logr.Logger) HibernatorAction {
	return &HibernatorActionImpl{
		Kubectl:          kubectl,
		historyUtil:      historyUtil,
		resourceAction:   resourceAction,
		resourceSelector: resourceSelector,
		log:              log,
	}
}

type HibernatorActionImpl struct {
	Kubectl          pkg.KubectlCmd
	historyUtil      History
	resourceAction   ResourceAction
	resourceSelector ResourceSelector
	log              logr.Logger
}

func (r *HibernatorActionImpl) unHibernate(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool) {

	reSync := hibernator.Status.Action == hibernator.Spec.Action

	hibernator.Status.Action = pincherv1alpha1.UnHibernate

	impactedObjects, excludedObjects := r.executeRules(hibernator, r.resourceAction.ResetScaleActionFactory(hibernator), reSync)

	if len(impactedObjects) > 0 {
		history := pincherv1alpha1.RevisionHistory{
			Time:            metav1.Time{Time: time.Now()},
			ID:              r.historyUtil.getNewRevisionID(hibernator.Status.History),
			Action:          pincherv1alpha1.UnHibernate,
			ImpactedObjects: impactedObjects,
			ExcludedObjects: excludedObjects,
		}
		hibernator.Status.History = r.historyUtil.addToHistory(history, hibernator.Status.History, reSync)
	}

	return hibernator, len(impactedObjects) > 0
}

func (r *HibernatorActionImpl) hibernate(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) (*pincherv1alpha1.Hibernator, bool) {

	impactedObjects, excludedObjects := make([]pincherv1alpha1.ImpactedObject, 0), make([]pincherv1alpha1.ExcludedObject, 0)
	reSync := false

	shouldHibernate := timeGap.WithinRange
	if hibernator.Spec.UnHibernate {
		shouldHibernate = false
	}
	if hibernator.Spec.Hibernate {
		shouldHibernate = true
	}

	if shouldHibernate {
		reSync = hibernator.Status.Action == pincherv1alpha1.Hibernate || hibernator.Status.Action == pincherv1alpha1.Sleep
		hibernator.Status.Action = pincherv1alpha1.Hibernate
		impactedObjects, excludedObjects = r.executeRules(hibernator, r.resourceAction.ScaleActionFactory(hibernator, timeGap), reSync)
	} else {
		reSync = hibernator.Status.Action == pincherv1alpha1.UnHibernate
		hibernator.Status.Action = pincherv1alpha1.UnHibernate
		impactedObjects, excludedObjects = r.executeRules(hibernator, r.resourceAction.ResetScaleActionFactory(hibernator), reSync)
	}

	if len(impactedObjects) > 0 {
		history := pincherv1alpha1.RevisionHistory{
			Time:            metav1.Time{Time: time.Now()},
			ID:              r.historyUtil.getNewRevisionID(hibernator.Status.History),
			Action:          pincherv1alpha1.Hibernate,
			ImpactedObjects: impactedObjects,
			ExcludedObjects: excludedObjects,
		}
		if !shouldHibernate {
			history.Action = pincherv1alpha1.UnHibernate
		}
		hibernator.Status.History = r.historyUtil.addToHistory(history, hibernator.Status.History, reSync)
	}

	r.log.Info("hibernate Operation - current run within range : next scheduled run at : Impacted Objects : excluded Objects", "withinRange", timeGap.WithinRange, "timeGap", timeGap.TimeGapInSeconds, "impactedObject", impactedObjects, "excludedObject", excludedObjects)

	return hibernator, len(impactedObjects) > 0
}

func (r *HibernatorActionImpl) delete(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool) {

	reSync := hibernator.Spec.Action == hibernator.Status.Action

	hibernator.Status.Action = pincherv1alpha1.Delete

	impactedObjects, excludedObjects := r.executeRules(hibernator, r.resourceAction.DeleteAction, reSync)

	if len(impactedObjects) > 0 {
		history := pincherv1alpha1.RevisionHistory{
			Time:            metav1.Time{Time: time.Now()},
			ID:              r.historyUtil.getNewRevisionID(hibernator.Status.History),
			Action:          pincherv1alpha1.Delete,
			ImpactedObjects: impactedObjects,
			ExcludedObjects: excludedObjects,
		}
		hibernator.Status.History = r.historyUtil.addToHistory(history, hibernator.Status.History, reSync)
	}

	r.log.Info("delete Operation - current run within range : next scheduled run at : Impacted Objects : excluded Objects", "impactedObjects", impactedObjects, "excludedObjects", excludedObjects)
	return hibernator, len(impactedObjects) > 0
}

func (r *HibernatorActionImpl) scale(hibernator *pincherv1alpha1.Hibernator, timeGap pincherv1alpha1.NearestTimeGap) (*pincherv1alpha1.Hibernator, bool) {

	reSync := hibernator.Spec.Action == hibernator.Status.Action

	hibernator.Status.Action = pincherv1alpha1.Scale

	impactedObjects, excludedObjects := make([]pincherv1alpha1.ImpactedObject, 0), make([]pincherv1alpha1.ExcludedObject, 0)
	if timeGap.WithinRange {
		impactedObjects, excludedObjects = r.executeRules(hibernator, r.resourceAction.ScaleActionFactory(hibernator, timeGap), reSync)
	} else {
		impactedObjects, excludedObjects = r.executeRules(hibernator, r.resourceAction.ResetScaleActionFactory(hibernator), reSync)
	}

	if len(impactedObjects) > 0 {
		history := pincherv1alpha1.RevisionHistory{
			Time:            metav1.Time{Time: time.Now()},
			ID:              r.historyUtil.getNewRevisionID(hibernator.Status.History),
			Action:          pincherv1alpha1.Scale,
			ImpactedObjects: impactedObjects,
			ExcludedObjects: excludedObjects,
		}
		hibernator.Status.History = r.historyUtil.addToHistory(history, hibernator.Status.History, reSync)
	}

	r.log.Info("Scale Operation - current run within range : next scheduled run at : Impacted Objects : excluded Objects", "withinRange", timeGap.WithinRange, "timeGap", timeGap.TimeGapInSeconds, "impactedObjects", impactedObjects, "excludedObjects", excludedObjects)

	return hibernator, len(impactedObjects) > 0
}

func (r *HibernatorActionImpl) executeRules(hibernator *pincherv1alpha1.Hibernator, execute Execute, reSync bool) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject) {
	//log := r.Log.WithValues("hibernator", r.getNamespacedName(hibernator))

	impactedObjects := make([]pincherv1alpha1.ImpactedObject, 0)
	excludedObjects := make([]pincherv1alpha1.ExcludedObject, 0)

	for _, rule := range hibernator.Spec.Selectors {
		inclusions := r.resourceSelector.getMatchingObjects(rule.Inclusions)
		exclusions := r.resourceSelector.getMatchingObjects(rule.Exclusions)
		included, excluded := r.resourceSelector.getIncludedExcludedObjects(inclusions, exclusions)

		impactedObjects, excludedObjects = execute(included)

		for _, ex := range excluded {
			excludedObject := pincherv1alpha1.ExcludedObject{
				ResourceKey: getResourceKey(ex),
			}
			excludedObjects = append(excludedObjects, excludedObject)
		}
	}

	return impactedObjects, excludedObjects
}
