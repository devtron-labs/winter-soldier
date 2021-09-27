package controllers

import (
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"strings"
	"time"
)

func (r *HibernatorReconciler) getPauseUntilDuration(hibernator *v1alpha1.Hibernator, now time.Time) (time.Duration, error) {
	diff := time.Duration(0) * time.Minute
	if len(strings.Trim(hibernator.Spec.PauseUntil.DateTime, " 	")) > 0 {
		tm, err := time.Parse(layout, strings.Trim(hibernator.Spec.PauseUntil.DateTime, " 	"))
		if err != nil {
			return diff, err
		} else {
			diff = tm.Sub(now)
		}
	}
	return diff, nil
}

func (r *HibernatorReconciler) timeElapsedSinceLastRunInMinutes(hibernator *v1alpha1.Hibernator) (timeElapsedSinceLastRunInMinutes float64, hasPreviousRun bool) {
	latestHistory := r.getLatestHistory(hibernator.Status.History)
	hasPreviousRun = false
	timeElapsedSinceLastRunInMinutes = 0.0
	if latestHistory != nil {
		timeElapsedSinceLastRunInMinutes = time.Now().Sub(latestHistory.Time.Time).Minutes()
	}
	return timeElapsedSinceLastRunInMinutes, hasPreviousRun
}

func (r *HibernatorReconciler) getRequeueTimeDuration(timeGap int, hibernator *v1alpha1.Hibernator) time.Duration {
	requeueTime := time.Duration(timeGap) * time.Second
	if requeueTime.Seconds() > float64(hibernator.Spec.ReSyncInterval) && hibernator.Spec.ReSyncInterval > 0 {
		requeueTime = time.Duration(hibernator.Spec.ReSyncInterval) * time.Second
	}
	if requeueTime.Seconds() < 60 {
		requeueTime = time.Duration(61) * time.Second
	}
	return requeueTime
}
