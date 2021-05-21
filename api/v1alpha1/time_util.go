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
	"strconv"
	"strings"
	"time"
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


func (t TimeRangesWithZone) NearestTimeGap(instant time.Time) (int, bool, error) {
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
	for _, tr := range t.TimeRanges {
		timeGap, contains, err := tr.NearestTimeGap(timeWithZone)
		if err != nil {
			return -1, false, err
		}
		if nearestTimeGap > timeGap {
			nearestTimeGap = timeGap
			inRange = contains
		}
	}
	return nearestTimeGap, inRange, nil
}

func (t TimeRange) NearestTimeGap(instant time.Time) (int, bool, error) {
	inRange := false
	timeGap := -1
	instantInSeconds := dayOfWeekToSeconds(int(instant.Weekday())) + hourToSeconds(instant.Hour()) + minToSeconds(instant.Minute()) + instant.Second()
	fromInSeconds, err := t.toSeconds(t.TimeFrom)
	if err != nil {
		return timeGap, inRange, err
	}
	toInSeconds, err := t.toSeconds(t.TimeTo)
	if err != nil {
		return timeGap, inRange, err
	}

	fromWeekdayOrdinal := t.WeekdayFrom.toOrdinal()
	toWeekdayOrdinal := t.WeekdayTo.toOrdinal()

	fromTillInstant := dayOfWeekToSeconds(int(instant.Weekday())) + fromInSeconds
	toTillInstant := dayOfWeekToSeconds(int(instant.Weekday())) + toInSeconds
	fromTillOrdinal := dayOfWeekToSeconds(fromWeekdayOrdinal) + fromInSeconds
	toTillOrdinal := dayOfWeekToSeconds(toWeekdayOrdinal) + toInSeconds
	inRange = fromWeekdayOrdinal <= int(instant.Weekday()) && int(instant.Weekday()) <= toWeekdayOrdinal && fromTillInstant <= instantInSeconds && instantInSeconds <= toTillInstant

	if inRange {
		timeGap = toTillInstant - instantInSeconds
	} else if fromTillOrdinal > instantInSeconds {
		timeGap = fromTillOrdinal - instantInSeconds
	} else if toTillOrdinal < instantInSeconds {
		timeGap  = fromTillOrdinal + dayOfWeekToSeconds(7) - instantInSeconds
	} else if fromTillInstant > instantInSeconds {
		timeGap = fromTillInstant - instantInSeconds
	} else if toTillInstant < instantInSeconds {
		timeGap = fromTillInstant + dayOfWeekToSeconds(1) - instantInSeconds
	}

	return timeGap, inRange, nil
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