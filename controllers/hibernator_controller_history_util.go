package controllers

import (
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"math"
)

func (r *HibernatorReconciler) getLatestHistory(revisionHistories []v1alpha1.RevisionHistory) *v1alpha1.RevisionHistory {
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

func (r *HibernatorReconciler) getNewRevisionID(revisionHistories []v1alpha1.RevisionHistory) int64 {
	maxID := int64(-1)
	for _, history := range revisionHistories {
		if history.ID > maxID {
			maxID = history.ID
		}
	}
	return maxID + 1
}

func (r *HibernatorReconciler) addToHistory(history v1alpha1.RevisionHistory, revisionHistories []v1alpha1.RevisionHistory) []v1alpha1.RevisionHistory {
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
	var finalHistories []v1alpha1.RevisionHistory
	for _, history := range revisionHistories {
		if history.ID != minID {
			finalHistories = append(finalHistories, history)
		}
	}
	finalHistories = append(finalHistories, history)
	return finalHistories

}

