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
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
)

func TestHibernatorReconciler_hibernate(t *testing.T) {
	hibernator := pincherv1alpha1.Hibernator{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       pincherv1alpha1.HibernatorSpec{
			TimeRangesWithZone:           pincherv1alpha1.TimeRangesWithZone{},
			Rules:                        []pincherv1alpha1.Rule{{
				Inclusions:  []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        []string{"app=delete"},
						Name:          "",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name:     "pras",
					},
				}},
				Exclusions:  []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        nil,
						Name:          "patch-demo",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name:     "pras",
					},
				}},
				Action:      "sleep",
				DeleteStore: false,
			}},
			UnHibernate:                  false,
			CanUnHibernateObjectManually: false,
			Pause:                        false,
			PauseUntil:                   pincherv1alpha1.TimeRange{},
		},
		Status:     pincherv1alpha1.HibernatorStatus{},
	}
	hibernator2 := pincherv1alpha1.Hibernator{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       pincherv1alpha1.HibernatorSpec{
			TimeRangesWithZone:           pincherv1alpha1.TimeRangesWithZone{},
			Rules:                        []pincherv1alpha1.Rule{{
				Inclusions:  []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        []string{"app=delete"},
						Name:          "",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name:     "pras",
					},
				}},
				Exclusions:  []pincherv1alpha1.Selector{{
					ObjectSelector: pincherv1alpha1.ObjectSelector{
						Labels:        nil,
						Name:          "patch-demo",
						Type:          "deployment",
						FieldSelector: nil,
					},
					NamespaceSelector: pincherv1alpha1.NamespaceSelector{
						Name:     "pras",
					},
				}},
				Action:      "delete",
				DeleteStore: false,
			}},
			UnHibernate:                  false,
			CanUnHibernateObjectManually: false,
			Pause:                        false,
			PauseUntil:                   pincherv1alpha1.TimeRange{},
		},
		Status:     pincherv1alpha1.HibernatorStatus{},
	}
	type fields struct {
		Client  client.Client
		Log     logr.Logger
		Scheme  *runtime.Scheme
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
	}
	type args struct {
		hibernator v1alpha1.Hibernator
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pincherv1alpha1.Hibernator
		wantErr bool
	}{
		{
			name: "hibernate sleep test",
			fields: fields{
				Client:  nil,
				Log:     nil,
				Scheme:  nil,
				kubectl: pkg.NewKubectl(),
				mapper:  pkg.NewMapperFactory(),
			},
			args: args{hibernator: hibernator},
		},
		{
			name: "hibernate delete test",
			fields: fields{
				Client:  nil,
				Log:     nil,
				Scheme:  nil,
				kubectl: pkg.NewKubectl(),
				mapper:  pkg.NewMapperFactory(),
			},
			args: args{hibernator: hibernator2},
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
			got, err := r.hibernate(tt.args.hibernator)
			if (err != nil) != tt.wantErr {
				t.Errorf("hibernate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hibernate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
