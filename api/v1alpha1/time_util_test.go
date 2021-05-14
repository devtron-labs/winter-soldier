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
		TimeFrom:       "18:00",
		TimeTo:         "20",
		CronExpression: "",
		WeekdayFrom:    "Fri",
		WeekdayTo:      "Sat",
	}
	const layout = "Jan 2, 2006 at 3:04pm (IST)"
	tm, _ := time.Parse(layout, "Mar 27, 2021 at 12:45pm (IST)")
	tm2, _ := time.Parse(layout, "Mar 27, 2021 at 18:45pm (IST)")
	tm3, _ := time.Parse(layout, "Mar 22, 2021 at 12:45pm (IST)")
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
				t1.Errorf("Contains() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeRangesWithZone_NearestTimeGap(t1 *testing.T) {
	tr1 := TimeRange{
		TimeFrom:       "18:00",
		TimeTo:         "20",
		CronExpression: "",
		WeekdayFrom:    "Fri",
		WeekdayTo:      "Sat",
	}
	tr2 := TimeRange{
		TimeFrom:       "15:00",
		TimeTo:         "21",
		CronExpression: "",
		WeekdayFrom:    "Tue",
		WeekdayTo:      "Fri",
	}
	tr3 := TimeRange{
		TimeFrom:       "15:00",
		TimeTo:         "21",
		CronExpression: "",
		WeekdayFrom:    "Mon",
		WeekdayTo:      "Wed",
	}
	const layout = "Jan 2, 2006 at 3:04pm (IST)"
	tm, _ := time.Parse(layout, "Mar 27, 2021 at 12:45pm (IST)")
	tm2, _ := time.Parse(layout, "Mar 27, 2021 at 12:15pm (IST)")
	tm3, _ := time.Parse(layout, "Mar 22, 2021 at 12:45pm (IST)")
	tm4, _ := time.Parse(layout, "Mar 27, 2021 at 3:15pm (IST)")
	tm5, _ := time.Parse(layout, "Mar 25, 2021 at 3:15pm (IST)")
	tm6, _ := time.Parse(layout, "Mar 25, 2021 at 3:15pm (IST)")
	tm7, _ := time.Parse(layout, "Mar 23, 2021 at 12:45pm (IST)")
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
		want    int
		want1   bool
		wantErr bool
	}{
		{
			name: "inRange",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm},
			want:    105,
			want1:   true,
			wantErr: false,
		},
		{
			name: "before",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm2},
			want:    15,
			want1:   false,
			wantErr: false,
		},
		{
			name: "before days",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm3},
			want:    5745,
			want1:   false,
			wantErr: false,
		},
		{
			name: "after",
			fields: fields{
				TimeRanges: []TimeRange{tr1},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm4},
			want:    8475,
			want1:   false,
			wantErr: false,
		},
		{
			name: "after",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr2},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm4},
			want:    3975,
			want1:   false,
			wantErr: false,
		},
		{
			name: "after - in between",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr3},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm5},
			want:    1275,
			want1:   false,
			wantErr: false,
		},
		{
			name: "in range for 3",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr3},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm6},
			want:    1275,
			want1:   false,
			wantErr: false,
		},
		{
			name: "in range for 3",
			fields: fields{
				TimeRanges: []TimeRange{tr1, tr3},
				TimeZone:   "Asia/Kolkata",
			},
			args:    args{instant: tm7},
			want:    165,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := TimeRangesWithZone{
				TimeRanges: tt.fields.TimeRanges,
				TimeZone:   tt.fields.TimeZone,
			}
			got, got1, err := t.NearestTimeGap(tt.args.instant)
			if (err != nil) != tt.wantErr {
				t1.Errorf("NearestTimeGap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("NearestTimeGap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t1.Errorf("NearestTimeGap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
