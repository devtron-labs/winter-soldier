package pkg

import (
	"github.com/tidwall/gjson"
	"reflect"
	"testing"
)

func TestSplitByMathematicalAndLogicalOperator(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "testing ==",
			args:    args{input: "name==prashant"},
			want:    []string{"==", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing !=",
			args:    args{input: "name!=prashant"},
			want:    []string{"!=", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing !",
			args:    args{input: "!name"},
			want:    []string{"!", "name"},
			wantErr: false,
		},
		{
			name:    "testing >=",
			args:    args{input: "name>=prashant"},
			want:    []string{">=", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing <=",
			args:    args{input: "name<=prashant"},
			want:    []string{"<=", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing =>",
			args:    args{input: "name=>prashant"},
			want:    []string{"=>", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing =<",
			args:    args{input: "name=<prashant"},
			want:    []string{"=<", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing >",
			args:    args{input: "name>prashant"},
			want:    []string{">", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing <",
			args:    args{input: "name<prashant"},
			want:    []string{"<", "name", "prashant"},
			wantErr: false,
		},
		{
			name:    "testing ",
			args:    args{input: "name"},
			want:    []string{"", "name"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitByLogicalOperator(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitByLogicalOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitByLogicalOperator() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyLogicalOperator(t *testing.T) {
	type args struct {
		result gjson.Result
		ops    []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "string",
			args: args{
				result: gjson.Result{
					Type: gjson.String,
					Str:  "Prashant",
				},
				ops: []string{"==", "", "Prashant"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: " >= ",
			args: args{
				result: gjson.Result{
					Type: gjson.Number,
					Num:  1.0,
				},
				ops: []string{">=", "", "0.9"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: " exists ",
			args: args{
				result: gjson.Result{
					Raw: "",
				},
				ops: []string{"!", "", "0.9"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: " <= ",
			args: args{
				result: gjson.Result{
					Type: gjson.Number,
					Num:  1.0,
				},
				ops: []string{"<=", "", "2"},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyLogicalOperator(tt.args.result, tt.args.ops)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyLogicalOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ApplyLogicalOperator() got = %v, want %v", got, tt.want)
			}
		})
	}
}
