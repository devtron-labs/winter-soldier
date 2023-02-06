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
	"context"
	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	"testing"
)

func TestHibernatorActionImpl_unHibernate(t *testing.T) {
	kubectl := pkg.NewKubectlMock(pkg.DeploymentObjectsMock)

	type Response struct {
		impactedObjects []pincherv1alpha1.ImpactedObject
		excludedObjects []pincherv1alpha1.ExcludedObject
	}

	type fields struct {
		Kubectl          pkg.KubectlCmd
		historyUtil      History
		resourceAction   ResourceAction
		resourceSelector ResourceSelector
	}
	type args struct {
		hibernator *pincherv1alpha1.Hibernator
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
		want1  bool
	}{
		{
			name: "base case",
			args: args{hibernator: &pkg.HibernateTest},
			fields: fields{
				Kubectl:     kubectl,
				historyUtil: &HistoryImpl{},
				resourceAction: &ResourceActionImpl{
					Kubectl:     kubectl,
					historyUtil: &HistoryImpl{},
				},
				resourceSelector: &ResourceSelectorImpl{
					Kubectl: kubectl,
					Mapper:  pkg.NewMockMapperFactory(),
					factory: pkg.NewMockFactory,
				},
			},
			want1: false,
			want: Response{
				impactedObjects: []pincherv1alpha1.ImpactedObject{
					{ResourceKey: "/pras/apps/v1/Deployment/rss-site", OriginalCount: 2, Status: "success"},
				},
				excludedObjects: []pincherv1alpha1.ExcludedObject{
					{ResourceKey: "/pras/extensions/v1beta1/Deployment/nginx-deployment", Reason: "error determining original count"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorActionImpl{
				Kubectl:          tt.fields.Kubectl,
				historyUtil:      tt.fields.historyUtil,
				resourceAction:   tt.fields.resourceAction,
				resourceSelector: tt.fields.resourceSelector,
			}
			got, got1 := r.unHibernate(tt.args.hibernator)
			if got1 != tt.want1 {
				t.Errorf("hibernate() got = %t, want %t", got1, tt.want1)
			}
			if tt.want1 == false {
				return
			}
			if got.Status.History != nil {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			if len(got.Status.History[0].ImpactedObjects) != len(tt.want.impactedObjects) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			if len(got.Status.History[0].ExcludedObjects) != len(tt.want.excludedObjects) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			impactedObjects := make(map[string]int, 0)
			for _, object := range tt.want.impactedObjects {
				impactedObjects[object.ResourceKey] = object.OriginalCount
			}
			excludedObjects := make(map[string]bool, 0)
			for _, object := range tt.want.excludedObjects {
				excludedObjects[object.ResourceKey] = true
			}
			for _, object := range got.Status.History[0].ImpactedObjects {
				if v, ok := impactedObjects[object.ResourceKey]; !ok || v != object.OriginalCount {
					t.Errorf("hibernate() got = %v, want %v", got, tt.want)
				}
			}
			for _, object := range got.Status.History[0].ExcludedObjects {
				if ok := excludedObjects[object.ResourceKey]; !ok {
					t.Errorf("hibernate() got = %v, want %v", got, tt.want)
				}
			}
			for _, object := range got.Status.History[0].ImpactedObjects {
				parts := strings.Split(object.ResourceKey, "/")
				name := parts[len(parts)-1]
				namespace := parts[1]
				kind := parts[len(parts)-2]
				r := &pkg.GetRequest{
					Name:             name,
					Namespace:        namespace,
					GroupVersionKind: schema.GroupVersionKind{Kind: kind},
				}
				o, err := tt.fields.Kubectl.GetResource(context.Background(), r)
				if err != nil {
					t.Errorf("hibernate() error fetching to check %s", object.ResourceKey)
				}
				j, _ := o.Manifest.MarshalJSON()
				replicaCount := gjson.Get(string(j), "specs.replicas")
				if replicaCount.Int() == 0 {
					t.Errorf("hibernate() replica count 0 for %s", object.ResourceKey)
				}
				ann := o.Manifest.GetAnnotations()
				if ann["hibernator.devtron.ai/replicas"] != "0" {
					t.Errorf("hibernate() annotations not set to 0 for %s", object.ResourceKey)
				}
			}
			if got1 != tt.want1 {
				t.Errorf("unHibernate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestHibernatorActionImpl_hibernate(t *testing.T) {
	kubectl := pkg.NewKubectlMock(pkg.DeploymentObjectsMock)

	type Response struct {
		impactedObjects []pincherv1alpha1.ImpactedObject
		excludedObjects []pincherv1alpha1.ExcludedObject
	}

	type fields struct {
		Kubectl          pkg.KubectlCmd
		historyUtil      History
		resourceAction   ResourceAction
		resourceSelector ResourceSelector
	}
	type args struct {
		hibernator *pincherv1alpha1.Hibernator
		timeGap    pincherv1alpha1.NearestTimeGap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
		want1  bool
	}{
		{
			name: "base case",
			args: args{
				hibernator: &pkg.HibernateTest,
				timeGap:    pincherv1alpha1.NearestTimeGap{WithinRange: true},
			},
			fields: fields{
				Kubectl:     kubectl,
				historyUtil: &HistoryImpl{},
				resourceAction: &ResourceActionImpl{
					Kubectl: kubectl,
				},
				resourceSelector: &ResourceSelectorImpl{
					Kubectl: kubectl,
					Mapper:  pkg.NewMockMapperFactory(),
					factory: pkg.NewMockFactory,
				},
			},
			want1: true,
			want: Response{
				impactedObjects: []pincherv1alpha1.ImpactedObject{
					{ResourceKey: "/pras/extensions/v1beta1/Deployment/nginx-deployment", OriginalCount: 3, Status: "success"},
				},
				excludedObjects: []pincherv1alpha1.ExcludedObject{
					{ResourceKey: "/pras/apps/v1/Deployment/rss-site"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorActionImpl{
				Kubectl:          tt.fields.Kubectl,
				historyUtil:      tt.fields.historyUtil,
				resourceAction:   tt.fields.resourceAction,
				resourceSelector: tt.fields.resourceSelector,
			}
			got, got1 := r.hibernate(tt.args.hibernator, tt.args.timeGap)
			if len(got.Status.History) != 1 {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			if len(got.Status.History[0].ImpactedObjects) != len(tt.want.impactedObjects) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			if len(got.Status.History[0].ExcludedObjects) != len(tt.want.excludedObjects) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			impactedObjects := make(map[string]int, 0)
			for _, object := range tt.want.impactedObjects {
				impactedObjects[object.ResourceKey] = object.OriginalCount
			}
			excludedObjects := make(map[string]bool, 0)
			for _, object := range tt.want.excludedObjects {
				excludedObjects[object.ResourceKey] = true
			}
			for _, object := range got.Status.History[0].ImpactedObjects {
				if v, ok := impactedObjects[object.ResourceKey]; !ok || v != object.OriginalCount {
					t.Errorf("hibernate() got = %v, want %v", got, tt.want)
				}
			}
			for _, object := range got.Status.History[0].ExcludedObjects {
				if ok := excludedObjects[object.ResourceKey]; !ok {
					t.Errorf("hibernate() got = %v, want %v", got, tt.want)
				}
			}
			for _, object := range got.Status.History[0].ImpactedObjects {
				parts := strings.Split(object.ResourceKey, "/")
				name := parts[len(parts)-1]
				namespace := parts[1]
				kind := parts[len(parts)-2]
				r := &pkg.GetRequest{
					Name:             name,
					Namespace:        namespace,
					GroupVersionKind: schema.GroupVersionKind{Kind: kind},
				}
				o, err := tt.fields.Kubectl.GetResource(context.Background(), r)
				if err != nil {
					t.Errorf("hibernate() error fetching to check %s", object.ResourceKey)
				}
				j, _ := o.Manifest.MarshalJSON()
				replicaCount := gjson.Get(string(j), "specs.replicas")
				if replicaCount.Int() != 0 {
					t.Errorf("hibernate() replica count not 0 for %s", object.ResourceKey)
				}
				ann := o.Manifest.GetAnnotations()
				if len(ann["hibernator.devtron.ai/replicas"]) == 0 || ann["hibernator.devtron.ai/replicas"] == "0" {
					t.Errorf("hibernate() annotations not set for %s", object.ResourceKey)
				}
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			//}
			if got1 != tt.want1 {
				t.Errorf("hibernate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestHibernatorActionImpl_delete(t *testing.T) {
	kubectl := pkg.NewKubectlMock(pkg.DeploymentObjectsMock)

	type Response struct {
		impactedObjects []pincherv1alpha1.ImpactedObject
		excludedObjects []pincherv1alpha1.ExcludedObject
	}

	type fields struct {
		Kubectl          pkg.KubectlCmd
		historyUtil      History
		resourceAction   ResourceAction
		resourceSelector ResourceSelector
	}
	type args struct {
		hibernator *pincherv1alpha1.Hibernator
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
		want1  bool
	}{
		{
			name: "base case",
			args: args{hibernator: &pkg.HibernateTest},
			fields: fields{
				Kubectl:     kubectl,
				historyUtil: &HistoryImpl{},
				resourceAction: &ResourceActionImpl{
					Kubectl: kubectl,
				},
				resourceSelector: &ResourceSelectorImpl{
					Kubectl: kubectl,
					Mapper:  pkg.NewMockMapperFactory(),
					factory: pkg.NewMockFactory,
				},
			},
			want1: true,
			want: Response{
				impactedObjects: []pincherv1alpha1.ImpactedObject{
					{ResourceKey: "/pras/extensions/v1beta1/Deployment/nginx-deployment", OriginalCount: 3, Status: "success"},
				},
				excludedObjects: []pincherv1alpha1.ExcludedObject{
					{ResourceKey: "/pras/apps/v1/Deployment/rss-site"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorActionImpl{
				Kubectl:          tt.fields.Kubectl,
				historyUtil:      tt.fields.historyUtil,
				resourceAction:   tt.fields.resourceAction,
				resourceSelector: tt.fields.resourceSelector,
			}
			got, got1 := r.delete(tt.args.hibernator)
			if len(got.Status.History) != 1 {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			if len(got.Status.History[0].ImpactedObjects) != len(tt.want.impactedObjects) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			if len(got.Status.History[0].ExcludedObjects) != len(tt.want.excludedObjects) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
			impactedObjects := make(map[string]bool, 0)
			for _, object := range tt.want.impactedObjects {
				impactedObjects[object.ResourceKey] = true
			}
			excludedObjects := make(map[string]bool, 0)
			for _, object := range tt.want.excludedObjects {
				excludedObjects[object.ResourceKey] = true
			}
			for _, object := range got.Status.History[0].ImpactedObjects {
				if ok := impactedObjects[object.ResourceKey]; !ok {
					t.Errorf("hibernate() got = %v, want %v", got, tt.want)
				}
			}
			for _, object := range got.Status.History[0].ExcludedObjects {
				if ok := excludedObjects[object.ResourceKey]; !ok {
					t.Errorf("hibernate() got = %v, want %v", got, tt.want)
				}
			}
			for _, object := range got.Status.History[0].ImpactedObjects {
				parts := strings.Split(object.ResourceKey, "/")
				name := parts[len(parts)-1]
				namespace := parts[1]
				kind := parts[len(parts)-2]
				r := &pkg.GetRequest{
					Name:             name,
					Namespace:        namespace,
					GroupVersionKind: schema.GroupVersionKind{Kind: kind},
				}
				o, err := tt.fields.Kubectl.GetResource(context.Background(), r)
				if err != nil {
					t.Errorf("hibernate() error fetching to check %s", object.ResourceKey)
				}
				if o.Manifest.Object != nil {
					t.Errorf("hibernate() error deleting %s", object.ResourceKey)
				}
			}
			if got1 != tt.want1 {
				t.Errorf("delete() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
