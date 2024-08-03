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
	"fmt"
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestHibernatorReconciler_handleLabelSelector(t *testing.T) {
	rule := v1alpha1.Selector{
		ObjectSelector: v1alpha1.ObjectSelector{
			Labels:        []string{"app=web"},
			Name:          "",
			Type:          "deployment",
			FieldSelector: nil,
		},
		NamespaceSelector: v1alpha1.NamespaceSelector{
			Name: "pras",
		},
	}
	type fields struct {
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
		factory func(mapper *pkg.Mapper) pkg.ArgsProcessor
		Log     logr.Logger
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
			name: "select deployment",
			fields: fields{
				kubectl: pkg.NewKubectlMock(pkg.DeploymentObjectsMock),
				mapper:  pkg.NewMockMapperFactory(),
				factory: pkg.NewMockFactory,
				Log:     ctrl.Log.WithName("test").WithName("Hibernator"),
			},
			args: args{rule: rule},
			want: []*unstructured.Unstructured{
				{Object: map[string]interface{}{"metadata": map[string]interface{}{"name": "rss-site"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceSelectorImpl{
				Kubectl: tt.fields.kubectl,
				Mapper:  tt.fields.mapper,
				factory: tt.fields.factory,
				Log:     tt.fields.Log,
			}
			got, err := r.handleLabelSelector(tt.args.rule)
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleLabelSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("handleLabelSelector() got = %v, want %v", got, tt.want)
			}
			gotKeys := make(map[string]bool, 0)
			wantKeys := make(map[string]bool, 0)
			for _, g := range got {
				gotKeys[g.GetName()] = true
			}
			for _, w := range tt.want {
				wantKeys[w.GetName()] = true
			}
			for k := range gotKeys {
				if !wantKeys[k] {
					t.Errorf("handleLabelSelector() excess key %s", k)
				}
			}
			for k := range wantKeys {
				if !gotKeys[k] {
					t.Errorf("handleLabelSelector() missing key %s", k)
				}
			}
		})
	}
}

func TestHibernatorReconciler_handleFieldSelector(t *testing.T) {
	type fields struct {
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
		factory func(mapper *pkg.Mapper) pkg.ArgsProcessor
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
			name: "label with field selector",
			fields: fields{
				kubectl: pkg.NewKubectlMock(pkg.DeploymentObjectsMock),
				mapper:  pkg.NewMockMapperFactory(),
				factory: pkg.NewMockFactory,
			},
			args: args{rule: v1alpha1.Selector{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{"app=nginx"},
					Name:          "",
					Type:          "deployment",
					FieldSelector: []string{"{{spec.replicas}}==3"},
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
			want: []*unstructured.Unstructured{
				{Object: map[string]interface{}{"metadata": map[string]interface{}{"name": "nginx"}}},
			},
			wantErr: false,
		},
		{
			name: "field selector without label and name",
			fields: fields{
				kubectl: pkg.NewKubectlMock(pkg.DeploymentObjectsMock),
				mapper:  pkg.NewMockMapperFactory(),
				factory: pkg.NewMockFactory,
			},
			args: args{rule: v1alpha1.Selector{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{},
					Name:          "",
					Type:          "deployment",
					FieldSelector: []string{"{{spec.replicas}}==3"},
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
			want: []*unstructured.Unstructured{
				{Object: map[string]interface{}{"metadata": map[string]interface{}{"name": "nginx-deployment"}}},
				{Object: map[string]interface{}{"metadata": map[string]interface{}{"name": "nginx"}}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceSelectorImpl{
				Kubectl: tt.fields.kubectl,
				Mapper:  tt.fields.mapper,
				factory: tt.fields.factory,
			}
			got, err := r.handleFieldSelector(tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleFieldSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("handleFieldSelector() got = %v, want %v", got, tt.want)
			}
			gotKeys := make(map[string]bool, 0)
			wantKeys := make(map[string]bool, 0)
			for _, g := range got {
				gotKeys[g.GetName()] = true
			}
			for _, w := range tt.want {
				wantKeys[w.GetName()] = true
			}
			for k := range gotKeys {
				if !wantKeys[k] {
					t.Errorf("handleLabelSelector() excess key %s", k)
				}
			}
		})
	}
}
