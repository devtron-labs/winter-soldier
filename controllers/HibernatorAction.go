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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type HibernatorAction interface {
	unHibernate(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool)
	hibernate(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool)
	delete(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool)
	executeRules(hibernator *pincherv1alpha1.Hibernator, execute Execute, reSync bool) ([]pincherv1alpha1.ImpactedObject, []pincherv1alpha1.ExcludedObject)
}

func NewHibernatorActionImpl(kubectl pkg.KubectlCmd, historyUtil History, resourceAction ResourceAction, resourceSelector ResourceSelector) HibernatorAction {
	return &HibernatorActionImpl{
		Kubectl:          kubectl,
		historyUtil:      historyUtil,
		resourceAction:   resourceAction,
		resourceSelector: resourceSelector,
	}
}

type HibernatorActionImpl struct {
	Kubectl        pkg.KubectlCmd
	historyUtil    History
	resourceAction ResourceAction
	resourceSelector ResourceSelector
}

func (r *HibernatorActionImpl) unHibernate(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool) {
	//log := r.Log.WithValues("hibernator", getNamespacedName(hibernator))
	//
	//log.Info("initiating unHibernate")

	reSync := hibernator.Status.Action == pincherv1alpha1.UnHibernate || hibernator.Status.Action == pincherv1alpha1.Delete

	hibernator.Status.Action = pincherv1alpha1.UnHibernate
	impactedObjects, excludedObjects := r.executeRules(hibernator, r.resourceAction.UnHibernateActionFactory(hibernator), reSync)

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
	//log.Info("ending unHibernate")

	return hibernator, len(impactedObjects) > 0
}

func (r *HibernatorActionImpl) hibernate(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool) {
	//log := r.Log.WithValues("hibernator", getNamespacedName(hibernator))
	//log.Info("starting hibernate")

	reSync := hibernator.Spec.Action == hibernator.Status.Action

	hibernator.Status.Action = pincherv1alpha1.Hibernate

	//patchZero := fmt.Sprintf(patch, 0)

	impactedObjects, excludedObjects := r.executeRules(hibernator, r.resourceAction.HibernateAction, reSync)

	if len(impactedObjects) > 0 {
		history := pincherv1alpha1.RevisionHistory{
			Time:            metav1.Time{Time: time.Now()},
			ID:              r.historyUtil.getNewRevisionID(hibernator.Status.History),
			Action:          pincherv1alpha1.Hibernate,
			ImpactedObjects: impactedObjects,
			ExcludedObjects: excludedObjects,
		}
		hibernator.Status.History = r.historyUtil.addToHistory(history, hibernator.Status.History, reSync)
	}

	//log.Info("ending hibernate")
	return hibernator, len(impactedObjects) > 0
}

func (r *HibernatorActionImpl) delete(hibernator *pincherv1alpha1.Hibernator) (*pincherv1alpha1.Hibernator, bool) {
	//log := r.Log.WithValues("hibernator", getNamespacedName(hibernator))
	//log.Info("starting delete")

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

	//log.Info("ending delete")
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
