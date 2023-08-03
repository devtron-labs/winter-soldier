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
	"testing"
	"time"
)

func TestTimeRangesWithZone_Contains(t1 *testing.T) {
	tr1 := TimeRange{
		TimeFrom:    "18:00",
		TimeTo:      "20",
		WeekdayFrom: "Fri",
		WeekdayTo:   "Sat",
	}
	tr2 := TimeRange{
		TimeFrom:    "18:00",
		TimeTo:      "20",
		WeekdayFrom: "Fri",
		WeekdayTo:   "Fri",
	}
	const layout = "Jan 2, 2006 at 3:04pm (IST)"
	tm, _ := time.Parse(layout, "Mar 27, 2021 at 12:45pm (IST)")  //Sat
	tm2, _ := time.Parse(layout, "Mar 27, 2021 at 18:45pm (IST)") //Sat
	tm3, _ := time.Parse(layout, "Mar 22, 2021 at 12:45pm (IST)") //Mon
	type fields struct {
		TimeRanges []TimeRange
		TimeZone   string
	}
	type args struct {
		instant time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm},
			want:    true,
			wantErr: false,
		},
		{
			name: "negative case",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm2},
			want:    false,
			wantErr: false,
		},
		{
			name: "negative case- weekday",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm3},
			want:    false,
			wantErr: false,
		},
		{
			name: "negative case - previous day current time inside",
			fields: fields{
				TimeRanges: []TimeRange{tr2},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm2},
			want:    false,
			wantErr: false,
		},
		{
			name: "negative case - previous day current time outside",
			fields: fields{
				TimeRanges: []TimeRange{tr2},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := TimeRangesWithZone{
				TimeRanges: tt.fields.TimeRanges,
				TimeZone:   tt.fields.TimeZone,
			}
			got, err := t.Contains(tt.args.instant)
			if (err != nil) != tt.wantErr {
				t1.Errorf("Contains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("Contains() got = %v, timeGapInSeconds %v", got, tt.want)
			}
		})
	}
}

func TestTimeRangesWithZone_NearestTimeGap(t1 *testing.T) {
	tr1 := TimeRange{
		TimeFrom:    "18:00",
		TimeTo:      "20",
		WeekdayFrom: "Fri",
		WeekdayTo:   "Sat",
	}
	tr2 := TimeRange{
		TimeFrom:    "15:00",
		TimeTo:      "21",
		WeekdayFrom: "Tue",
		WeekdayTo:   "Fri",
	}
	tr3 := TimeRange{
		TimeFrom:    "15:00",
		TimeTo:      "21",
		WeekdayFrom: "Mon",
		WeekdayTo:   "Wed",
	}
	tr4 := TimeRange{
		TimeFrom:    "18:00",
		TimeTo:      "20",
		WeekdayFrom: "Fri",
		WeekdayTo:   "Sun",
	}
	tr5 := TimeRange{
		TimeFrom:    "19:00",
		TimeTo:      "22",
		WeekdayFrom: "Sun",
		WeekdayTo:   "Tue",
	}
	tr6 := TimeRange{
		TimeFrom:    "18:35",
		TimeTo:      "18:38:00",
		WeekdayFrom: "Mon",
		WeekdayTo:   "Sun",
	}
	tr7 := TimeRange{
		TimeFrom:    "18:38",
		TimeTo:      "18:40:00",
		WeekdayFrom: "Mon",
		WeekdayTo:   "Sun",
	}
	tr8 := TimeRange{
		TimeFrom:    "18:38",
		TimeTo:      "18:40:00",
		WeekdayFrom: "Mon",
		WeekdayTo:   "Mon",
	}
	tm, _ := time.Parse(time.RFC1123, "Sat, 27 Mar 2021 18:15:00 IST")   //Sat
	tm2, _ := time.Parse(time.RFC1123, "Sat, 27 Mar 2021 17:45:00 IST")  //Sat
	tm3, _ := time.Parse(time.RFC1123, "Mon, 22 Mar 2021 18:15:00 IST")  //Mon
	tm4, _ := time.Parse(time.RFC1123, "Sat, 27 Mar 2021 20:45:00 IST")  //Sat
	tm5, _ := time.Parse(time.RFC1123, "Thu, 25 Mar 2021 20:45:00 IST")  //Thu
	tm6, _ := time.Parse(time.RFC1123, "Thu, 25 Mar 2021 20:45:00 IST")  //Thu
	tm7, _ := time.Parse(time.RFC1123, "Tue, 23 Mar 2021 18:15:00 IST")  //Tue
	tm8, _ := time.Parse(time.RFC1123, "Sun, 28 Mar 2021 18:15:00 IST")  //Sun
	tm9, _ := time.Parse(time.RFC1123, "Sat, 14 Jan 2023 18:36:00 IST")  //Sat
	tm10, _ := time.Parse(time.RFC1123, "Sat, 14 Jan 2023 18:39:00 IST") //Sat
	tm11, _ := time.Parse(time.RFC1123, "Sun, 15 Jan 2023 18:36:00 IST") //Sun
	tm12, _ := time.Parse(time.RFC1123, "Sun, 15 Jan 2023 18:39:00 IST") //Sun
	tm13, _ := time.Parse(time.RFC1123, "Tue, 23 Mar 2021 18:39:00 IST") //Tue
	type fields struct {
		TimeRanges []TimeRange
		TimeZone   string
	}
	type args struct {
		instant time.Time
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		timeGapInSeconds int
		matchedIndex     int
		want1            bool
		wantErr          bool
	}{
		{
			name: "inRange",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm},
			timeGapInSeconds: 105 * 60,
			matchedIndex:     0,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "before",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm2},
			timeGapInSeconds: 15 * 60,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
		{
			name: "before days",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm3},
			timeGapInSeconds: 5745 * 60,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
		{
			name: "after",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm4},
			timeGapInSeconds: 8475 * 60,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
		{
			name: "after",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr2},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm4},
			timeGapInSeconds: 3975 * 60,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
		{
			name: "after - in between",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr3},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm5},
			timeGapInSeconds: 1275 * 60,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
		{
			name: "in range for 3",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr3},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm6},
			timeGapInSeconds: 1275 * 60,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
		{
			name: "in range for 3",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr3},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm7},
			timeGapInSeconds: 165 * 60,
			matchedIndex:     1,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "inRange circular",
			fields: fields{
				TimeRanges: []TimeRange{tr4},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm8},
			timeGapInSeconds: 105 * 60,
			matchedIndex:     0,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "inRange circular with overlapping ranges",
			fields: fields{
				TimeRanges: []TimeRange{tr4, tr5},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm8},
			timeGapInSeconds: 105 * 60,
			matchedIndex:     0,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "inRange circular with nonoverlaping ranges",
			fields: fields{
				TimeRanges: []TimeRange{tr6, tr7},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm9},
			timeGapInSeconds: 2 * 60,
			matchedIndex:     0,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "inRange circular with nonoverlaping ranges",
			fields: fields{
				TimeRanges: []TimeRange{tr6, tr7},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm10},
			timeGapInSeconds: 1 * 60,
			matchedIndex:     1,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "inRange circular with nonoverlaping ranges",
			fields: fields{
				TimeRanges: []TimeRange{tr6, tr7},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm11},
			timeGapInSeconds: 2 * 60,
			matchedIndex:     0,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "inRange circular with nonoverlaping ranges",
			fields: fields{
				TimeRanges: []TimeRange{tr6, tr7},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm12},
			timeGapInSeconds: 1 * 60,
			matchedIndex:     1,
			want1:            true,
			wantErr:          false,
		},
		{
			name: "previous date matching time check",
			fields: fields{
				TimeRanges: []TimeRange{tr8},
				TimeZone:   "Asia/Kolkata",
			},
			args:             args{instant: tm13},
			timeGapInSeconds: 518340,
			matchedIndex:     -1,
			want1:            false,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := TimeRangesWithZone{
				TimeRanges: tt.fields.TimeRanges,
				TimeZone:   tt.fields.TimeZone,
			}
			nearestTimeGap, err := t.NearestTimeGapInSeconds(tt.args.instant)
			if (err != nil) != tt.wantErr {
				t1.Errorf("NearestTimeGapInSeconds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if nearestTimeGap.TimeGapInSeconds != tt.timeGapInSeconds {
				t1.Errorf("NearestTimeGapInSeconds() timeGapinSeconds got = %v, timeGapInSeconds %v", nearestTimeGap.TimeGapInSeconds, tt.timeGapInSeconds)
			}
			if nearestTimeGap.MatchedIndex != tt.matchedIndex {
				t1.Errorf("NearestTimeGapInSeconds() matchedIndex got = %v, timeGapInSeconds %v", nearestTimeGap.MatchedIndex, tt.matchedIndex)
			}
			if nearestTimeGap.WithinRange != tt.want1 {
				t1.Errorf("NearestTimeGapInSeconds() WithinRange got1 = %t, timeGapInSeconds %t", nearestTimeGap.WithinRange, tt.want1)
			}
		})
	}
}
