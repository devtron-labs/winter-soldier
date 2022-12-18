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
	"github.com/devtron-labs/winter-soldier/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"

	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
)

func TestHibernatorReconciler_hibernate(t *testing.T) {
	hibernator := pincherv1alpha1.Hibernator{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: pincherv1alpha1.HibernatorSpec{
			When: pincherv1alpha1.TimeRangesWithZone{},
			Selectors: []pincherv1alpha1.Rule{{
				Inclusions: []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        []string{"app=delete"},
						Name:          "",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name: "pras",
					},
				}},
				Exclusions: []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        nil,
						Name:          "patch-demo",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name: "pras",
					},
				}},
			}},
			UnHibernate: false,
			Pause:       false,
			PauseUntil:  pincherv1alpha1.DateTimeWithZone{},
		},
		Status: pincherv1alpha1.HibernatorStatus{},
	}
	hibernator2 := pincherv1alpha1.Hibernator{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: pincherv1alpha1.HibernatorSpec{
			When: pincherv1alpha1.TimeRangesWithZone{},
			Selectors: []pincherv1alpha1.Rule{{
				Inclusions: []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        []string{"app=delete"},
						Name:          "",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name: "pras",
					},
				}},
				Exclusions: []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        nil,
						Name:          "patch-demo",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name: "pras",
					},
				}},
			}},
			UnHibernate: false,
			Pause:       false,
			PauseUntil:  pincherv1alpha1.DateTimeWithZone{},
		},
		Status: pincherv1alpha1.HibernatorStatus{},
	}

	type fields struct {
		kubectl          pkg.KubectlCmd
		historyUtil      History
		resourceAction   ResourceAction
		resourceSelector ResourceSelector
	}
	type args struct {
		hibernator v1alpha1.Hibernator
		timeGap    pincherv1alpha1.NearestTimeGap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *pincherv1alpha1.Hibernator
	}{
		{
			name: "hibernate sleep test",
			fields: fields{
				kubectl:     pkg.NewKubectl(),
				historyUtil: &HistoryImpl{},
			},
			args: args{
				hibernator: hibernator,
				timeGap:    pincherv1alpha1.NearestTimeGap{WithinRange: true},
			},
		},
		{
			name: "hibernate delete test",
			fields: fields{
				kubectl:     pkg.NewKubectl(),
				historyUtil: &HistoryImpl{},
			},
			args: args{hibernator: hibernator2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewHibernatorActionImpl(tt.fields.kubectl, tt.fields.historyUtil, tt.fields.resourceAction, tt.fields.resourceSelector)
			got, _ := r.hibernate(&tt.args.hibernator, tt.args.timeGap)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
