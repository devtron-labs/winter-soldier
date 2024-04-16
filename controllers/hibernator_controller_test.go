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
	"encoding/json"
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		log              logr.Logger
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
			r := NewHibernatorActionImpl(tt.fields.kubectl, tt.fields.historyUtil, tt.fields.resourceAction, tt.fields.resourceSelector, tt.fields.log)
			got, _ := r.hibernate(&tt.args.hibernator, tt.args.timeGap)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHibernatorReconciler_process(t *testing.T) {
	var hibernator pincherv1alpha1.Hibernator
	err := json.Unmarshal([]byte(hibernator_mock_1), &hibernator)
	if err != nil {
		panic(err)
	}
	history := NewHistoryImpl()
	timeUtil := NewTimeUtilImpl(history)
	type fields struct {
		Client           client.Client
		Log              logr.Logger
		Scheme           *runtime.Scheme
		Kubectl          pkg.KubectlCmd
		Mapper           *pkg.Mapper
		HibernatorAction HibernatorAction
		TimeUtil         TimeUtil
	}
	type args struct {
		hibernator pincherv1alpha1.Hibernator
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    controllerruntime.Result
		wantErr bool
	}{
		{
			name: "previous day - current time outside",
			fields: fields{
				Client:           nil,
				Log:              controllerruntime.Log.WithName("controllers").WithName("Hibernator"),
				Scheme:           nil,
				Kubectl:          pkg.NewKubectlMock(pkg.DeploymentObjectsMock),
				Mapper:           nil,
				HibernatorAction: nil,
				TimeUtil:         timeUtil,
			},
			args: args{hibernator: hibernator},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorReconciler{
				Client:           tt.fields.Client,
				Log:              tt.fields.Log,
				Scheme:           tt.fields.Scheme,
				Kubectl:          tt.fields.Kubectl,
				Mapper:           tt.fields.Mapper,
				HibernatorAction: tt.fields.HibernatorAction,
				TimeUtil:         tt.fields.TimeUtil,
			}
			got, err := r.process(tt.args.hibernator)
			if (err != nil) != tt.wantErr {
				t.Errorf("process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("process() got = %v, want %v", got, tt.want)
			}
		})
	}
}
