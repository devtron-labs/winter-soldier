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
	instantInMinutes := instant.Hour() * 60 + instant.Minute()
	from, err := t.toMinutes(t.TimeFrom)
	if err != nil {
		return false, err
	}
	to, err := t.toMinutes(t.TimeTo)
	if err != nil {
		return false, err
	}
	if from > instantInMinutes || to < instantInMinutes {
		return false, nil
	}
	fromWeekdayOrdinal := t.WeekdayFrom.toOrdinal()
	toWeekdayOrdinal := t.WeekdayTo.toOrdinal()
	if fromWeekdayOrdinal > int(weekday) || toWeekdayOrdinal < int(weekday) {
		return false, nil
	}
	return true, nil
}

func (t TimeRange) toMinutes(time string) (int, error) {
	timeParts := strings.Split(time, ":")
	hr, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return 0, err
	}
	timeInMinutes := hr * 60
	if len(timeParts) == 2 {
		min, err := strconv.Atoi(timeParts[1])
		if err != nil {
			return 0, err
		}
		timeInMinutes += min
	}
	return timeInMinutes, nil
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
	instantInMinutes := int(instant.Weekday()) * 24 * 60 + instant.Hour() * 60 + instant.Minute()
	fromInMinutes, err := t.toMinutes(t.TimeFrom)
	if err != nil {
		return timeGap, inRange, err
	}
	toInMinutes, err := t.toMinutes(t.TimeTo)
	if err != nil {
		return timeGap, inRange, err
	}

	fromWeekdayOrdinal := t.WeekdayFrom.toOrdinal()
	toWeekdayOrdinal := t.WeekdayTo.toOrdinal()

	fromTillInstant := int(instant.Weekday()) * 24 * 60 + fromInMinutes
	toTillInstant := int(instant.Weekday()) * 24 * 60 + toInMinutes
	fromTillOrdinal := fromWeekdayOrdinal * 24 * 60 + fromInMinutes
	toTillOrdinal := toWeekdayOrdinal * 24 * 60 + toInMinutes
	inRange = fromWeekdayOrdinal <= int(instant.Weekday()) && int(instant.Weekday()) <= toWeekdayOrdinal && fromTillInstant <= instantInMinutes && instantInMinutes <= toTillInstant

	if inRange {
		timeGap = toTillInstant - instantInMinutes
	} else if fromTillOrdinal > instantInMinutes  {
		timeGap = fromTillOrdinal - instantInMinutes
	} else if toTillOrdinal < instantInMinutes  {
		timeGap  = fromTillOrdinal + 7 * 24 * 60 - instantInMinutes
	} else if fromTillInstant > instantInMinutes {
		timeGap = fromTillInstant - instantInMinutes
	} else if toTillInstant < instantInMinutes {
		timeGap = fromTillInstant + 1 * 24 * 60 - instantInMinutes
	}

	return timeGap, inRange, nil
}