package controllers

import (
	"fmt"
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/go-logr/logr"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestHibernatorReconciler_handleLabelSelector(t *testing.T) {
	rule := v1alpha1.Selector{
		ObjectSelector: v1alpha1.ObjectSelector{
			Labels:        []string{"app=nats-streaming"},
			Name:          "",
			Type:          "sts",
			FieldSelector: nil,
		},
		NamespaceSelector: v1alpha1.NamespaceSelector{
			Name:     "devtroncd",
		},
	}
	type fields struct {
		Client  client.Client
		Log     logr.Logger
		Scheme  *runtime.Scheme
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
	}
	type args struct {
		rule v1alpha1.Selector
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "select sts",
			fields: fields{
				Client:  nil,
				Log:     nil,
				Scheme:  nil,
				kubectl: pkg.NewKubectl(),
				mapper:  pkg.NewMapperFactory(),
			},
			args: args{rule: rule},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorReconciler{
				Client:  tt.fields.Client,
				Log:     tt.fields.Log,
				Scheme:  tt.fields.Scheme,
				Kubectl: tt.fields.kubectl,
				Mapper:  tt.fields.mapper,
			}
			got, err := r.handleLabelSelector(tt.args.rule)
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleLabelSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleLabelSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHibernatorReconciler_handleFieldSelector(t *testing.T) {
	//js := `{"person":{"name": "abcd", "age": 12}}`
	//var obj map[string]interface{}
	//err := json.Unmarshal([]byte(js), &obj)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//uns := unstructured.Unstructured{Object: obj}
	type fields struct {
		Client  client.Client
		Log     logr.Logger
		Scheme  *runtime.Scheme
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
	}
	type args struct {
		rule v1alpha1.Selector
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want   []string
		wantErr bool
	}{
		{
			name: "label with field selector",
			fields: fields{
				kubectl: pkg.NewKubectl(),
				mapper: pkg.NewMapperFactory(),
			},
			args: args{rule: v1alpha1.Selector{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{"app=nats-streaming"},
					Name:          "",
					Type:          "sts",
					FieldSelector: []string{"metadata.name==nats-streaming-demo-devtroncd-nats-streaming"},
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name:     "devtroncd",
				},
			}},
			want:    []string{"nats-streaming-demo-devtroncd-nats-streaming"},
			wantErr: false,
		},
		{
			name: "field selector without label and name",
			fields: fields{
				kubectl: pkg.NewKubectl(),
				mapper: pkg.NewMapperFactory(),
			},
			args: args{rule: v1alpha1.Selector{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{},
					Name:          "",
					Type:          "sts",
					FieldSelector: []string{"metadata.name==nats-streaming-demo-devtroncd-nats-streaming"},
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name:     "devtroncd",
				},
			}},
			want:    []string{"nats-streaming-demo-devtroncd-nats-streaming"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorReconciler{
				Client:  tt.fields.Client,
				Log:     tt.fields.Log,
				Scheme:  tt.fields.Scheme,
				Kubectl: tt.fields.kubectl,
				Mapper:  tt.fields.mapper,
			}
			got, err := r.handleFieldSelector(tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleFieldSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("handleFieldSelector() got = %v, want %v", got, tt.want)
			}
			expt := map[string]bool{}
			for _, w := range tt.want {
				expt[w] = true
			}
			for _, g := range got {
				j, err := g.MarshalJSON()
				if err != nil {
					t.Errorf("%+v",err)
				}
				name := gjson.Get(string(j), "metadata.name")
				if !expt[name.Str] {
					t.Errorf("not found %s is %v", name.Str, tt.want)
				}
			}
		})
	}
}
