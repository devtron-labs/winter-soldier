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
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"math"
)

type HistoryUtil interface {
	getLatestHistory(revisionHistories []v1alpha1.RevisionHistory) *v1alpha1.RevisionHistory
	getNewRevisionID(revisionHistories []v1alpha1.RevisionHistory) int64
	addToHistory(history v1alpha1.RevisionHistory, revisionHistories []v1alpha1.RevisionHistory, reSync bool) []v1alpha1.RevisionHistory
}

type HistoryUtilImpl struct {}

func (r *HistoryUtilImpl) getLatestHistory(revisionHistories []v1alpha1.RevisionHistory) *v1alpha1.RevisionHistory {
	maxID := int64(-1)
	found := false
	var latestHistory v1alpha1.RevisionHistory
	for _, history := range revisionHistories {
		if history.ID > maxID {
			maxID = history.ID
			latestHistory = history
			found = true
		}
	}
	if !found {
		return nil
	}
	return &latestHistory
}

func (r *HistoryUtilImpl) getNewRevisionID(revisionHistories []v1alpha1.RevisionHistory) int64 {
	maxID := int64(-1)
	for _, history := range revisionHistories {
		if history.ID > maxID {
			maxID = history.ID
		}
	}
	return maxID + 1
}

//TODO: if last entry in history is same as the new entry for example both has hibernate==true then merge them
func (r *HistoryUtilImpl) addToHistory(history v1alpha1.RevisionHistory, revisionHistories []v1alpha1.RevisionHistory, reSync bool) []v1alpha1.RevisionHistory {
	//if len(revisionHistories) < 10 {
	//	revisionHistories = append(revisionHistories, history)
	//	return revisionHistories
	//}
	minID := int64(math.MaxInt64)
	maxID := int64(-1)
	for _, history := range revisionHistories {
		if history.ID < minID {
			minID = history.ID
		}
		if history.ID > maxID {
			maxID = history.ID
		}
	}
	var finalHistories []v1alpha1.RevisionHistory
	for _, history := range revisionHistories {
		if history.ID != minID {
			finalHistories = append(finalHistories, history)
		}
	}
	finalHistories = append(finalHistories, history)
	return finalHistories
}

