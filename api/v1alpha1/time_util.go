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

package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	firstDayOfWeek             = 0
	lastDayOfWeek              = 6
	MinReSyncIntervalInSeconds = 60
)

func (t TimeRangesWithZone) Contains(instant time.Time) (bool, error) {
	zone := "UTC"
	if len(t.TimeZone) != 0 {
		zone = t.TimeZone
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return false, err
	}
	timeWithZone := instant.In(loc)
	for _, tr := range t.TimeRanges {
		contains, err := tr.Contains(timeWithZone)
		if err != nil {
			return false, err
		}
		if contains {
			return contains, nil
		}
	}
	return false, nil
}

func (t TimeRange) Contains(instant time.Time) (bool, error) {
	weekday := instant.Weekday()
	instantInSeconds := hourToSeconds(instant.Hour()) + minToSeconds(instant.Minute()) + instant.Second()
	from, err := t.toSeconds(t.TimeFrom)
	if err != nil {
		return false, err
	}
	to, err := t.toSeconds(t.TimeTo)
	if err != nil {
		return false, err
	}
	if from > instantInSeconds || to < instantInSeconds {
		return false, nil
	}
	fromWeekdayOrdinal := t.WeekdayFrom.toOrdinal()
	toWeekdayOrdinal := t.WeekdayTo.toOrdinal()
	if fromWeekdayOrdinal > int(weekday) || toWeekdayOrdinal < int(weekday) {
		return false, nil
	}
	return true, nil
}

func (t TimeRange) toSeconds(time string) (int, error) {
	timeParts := strings.Split(time, ":")
	hr, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return 0, err
	}
	timeInSeconds := hourToSeconds(hr)
	if len(timeParts) >= 2 {
		min, err := strconv.Atoi(timeParts[1])
		if err != nil {
			return 0, err
		}
		timeInSeconds += minToSeconds(min)
	}
	if len(timeParts) >= 3 {
		sec, err := strconv.Atoi(timeParts[2])
		if err != nil {
			return 0, err
		}
		timeInSeconds += sec
	}
	return timeInSeconds, nil
}

func (w Weekday) toOrdinal() int {
	switch strings.ToLower(string(w)) {
	case "sun":
		return 0
	case "mon":
		return 1
	case "tue":
		return 2
	case "wed":
		return 3
	case "thu":
		return 4
	case "fri":
		return 5
	case "sat":
		return 6
	default:
		return -1
	}
}

func (t TimeRangesWithZone) NearestTimeGapInSeconds(instant time.Time) (int, bool, error) {
	zone := "UTC"
	if len(t.TimeZone) != 0 {
		zone = t.TimeZone
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return -1, false, err
	}
	timeWithZone := instant.In(loc)
	nearestTimeGap := 2147483647
	inRange := false
	normalizedTimeRanges := t.normalizeTimeRange()
	var matchedTimeRange *TimeRange
	for _, tr := range normalizedTimeRanges {
		timeGap, contains, err := tr.NearestTimeGapInSeconds(timeWithZone)
		if err != nil {
			return -1, false, err
		}
		if contains && inRange && nearestTimeGap > timeGap {
			nearestTimeGap = timeGap
			inRange = contains
			t := cloneTimeRange(tr)
			matchedTimeRange = &t
		} else if contains && !inRange {
			nearestTimeGap = timeGap
			inRange = contains
			t := cloneTimeRange(tr)
			matchedTimeRange = &t
		} else if !inRange && nearestTimeGap > timeGap {
			nearestTimeGap = timeGap
			inRange = contains
			t := cloneTimeRange(tr)
			matchedTimeRange = &t
		}
	}
	fmt.Printf("matched timeRange %v\n", matchedTimeRange)
	return nearestTimeGap, inRange, nil
}

func (t TimeRangesWithZone) normalizeTimeRange() []TimeRange {
	var normalizedTimeRanges []TimeRange
	for _, tr := range t.TimeRanges {
		if tr.WeekdayFrom.toOrdinal() < tr.WeekdayTo.toOrdinal() {
			normalizedTimeRanges = append(normalizedTimeRanges, tr)
		} else {
			tr1 := cloneTimeRange(tr)
			tr1.WeekdayTo = Sat
			tr2 := cloneTimeRange(tr)
			tr2.WeekdayFrom = Sun
			normalizedTimeRanges = append(normalizedTimeRanges, tr1)
			normalizedTimeRanges = append(normalizedTimeRanges, tr2)
		}
	}
	return normalizedTimeRanges
}

func (t TimeRange) NearestTimeGapInSeconds(instant time.Time) (int, bool, error) {
	inRange := false
	timeGap := -1
	instantInSeconds := hourToSeconds(instant.Hour()) + minToSeconds(instant.Minute()) + instant.Second()
	startTimeInSeconds, err := t.toSeconds(t.TimeFrom)
	if err != nil {
		return timeGap, inRange, err
	}
	endTimeInSeconds, err := t.toSeconds(t.TimeTo)
	if err != nil {
		return timeGap, inRange, err
	}

	fromWeekdayOrdinal := t.WeekdayFrom.toOrdinal()
	toWeekdayOrdinal := t.WeekdayTo.toOrdinal()

	instantWeekdayInSeconds := dayOfWeekToSeconds(int(instant.Weekday()))
	startTimeOnInstantInSeconds := instantWeekdayInSeconds + startTimeInSeconds
	endTimeOnInstantInSeconds := instantWeekdayInSeconds + endTimeInSeconds
	startTimeOnStartDayInSeconds := dayOfWeekToSeconds(fromWeekdayOrdinal) + startTimeInSeconds
	endTimeOnEndDayInSeconds := dayOfWeekToSeconds(toWeekdayOrdinal) + endTimeInSeconds
	inRange = fromWeekdayOrdinal <= int(instant.Weekday()) && int(instant.Weekday()) <= toWeekdayOrdinal && startTimeInSeconds <= instantInSeconds && instantInSeconds <= endTimeInSeconds

	if inRange {
		timeGap = endTimeInSeconds - instantInSeconds
	} else if startTimeOnStartDayInSeconds > instantInSeconds+instantWeekdayInSeconds {
		timeGap = startTimeOnStartDayInSeconds - instantWeekdayInSeconds - instantInSeconds
	} else if endTimeOnEndDayInSeconds < instantInSeconds+instantWeekdayInSeconds {
		timeGap = startTimeOnStartDayInSeconds + dayOfWeekToSeconds(7) - instantInSeconds - instantWeekdayInSeconds
	} else if startTimeOnInstantInSeconds > instantInSeconds+instantWeekdayInSeconds {
		timeGap = startTimeOnInstantInSeconds - instantInSeconds - instantWeekdayInSeconds
	} else if endTimeOnInstantInSeconds < instantInSeconds+instantWeekdayInSeconds {
		timeGap = startTimeOnInstantInSeconds + dayOfWeekToSeconds(1) - instantInSeconds
	}

	return timeGap, inRange, nil
}

func cloneTimeRange(timeRange TimeRange) TimeRange {
	return TimeRange{
		TimeZone:    timeRange.TimeZone,
		TimeFrom:    timeRange.TimeFrom,
		TimeTo:      timeRange.TimeTo,
		WeekdayFrom: timeRange.WeekdayFrom,
		WeekdayTo:   timeRange.WeekdayTo,
	}
}

func dayOfWeekToSeconds(weekday int) int {
	return weekday * 24 * 60 * 60
}

func hourToSeconds(hours int) int {
	return hours * 60 * 60
}

func minToSeconds(min int) int {
	return min * 60
}
