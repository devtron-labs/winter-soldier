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
	"strings"
	"time"
)

type TimeUtil interface {
	getPauseUntilDuration(hibernator *v1alpha1.Hibernator, now time.Time) (time.Duration, error)
	timeElapsedSinceLastRunInSeconds(hibernator *v1alpha1.Hibernator) (timeElapsedSinceLastRunInSeconds float64, hasPreviousRun bool)
	getRequeueTimeDuration(timeGap int, hibernator *v1alpha1.Hibernator) time.Duration
}

type TimeUtilImpl struct {
	historyUtil History
}

func (r *TimeUtilImpl) getPauseUntilDuration(hibernator *v1alpha1.Hibernator, now time.Time) (time.Duration, error) {
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

func (r *TimeUtilImpl) timeElapsedSinceLastRunInSeconds(hibernator *v1alpha1.Hibernator) (timeElapsedSinceLastRunInSeconds float64, hasPreviousRun bool) {
	latestHistory := r.historyUtil.getLatestHistory(hibernator.Status.History)
	hasPreviousRun = false
	timeElapsedSinceLastRunInSeconds = 0.0
	if latestHistory != nil {
		timeElapsedSinceLastRunInSeconds = time.Now().Sub(latestHistory.Time.Time).Seconds()
	}
	return timeElapsedSinceLastRunInSeconds, hasPreviousRun
}

func (r *TimeUtilImpl) getRequeueTimeDuration(timeGap int, hibernator *v1alpha1.Hibernator) time.Duration {
	requeueTime := time.Duration(timeGap) * time.Second
	if requeueTime.Seconds() > float64(hibernator.Spec.ReSyncInterval) && hibernator.Spec.ReSyncInterval > 0 {
		requeueTime = time.Duration(hibernator.Spec.ReSyncInterval) * time.Second
	}
	if requeueTime.Seconds() < v1alpha1.MinReSyncIntervalInSeconds {
		requeueTime = time.Duration(61) * time.Second
	}
	return requeueTime
}
