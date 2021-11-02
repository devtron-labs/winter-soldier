package pkg

import (
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var HibernateTest = v1alpha1.Hibernator{
	TypeMeta:   metav1.TypeMeta{
		Kind:       "hibernator",
		APIVersion: "pincher.devtron.ai/",
	},
	ObjectMeta: metav1.ObjectMeta{},
	Spec: v1alpha1.HibernatorSpec{
		When: v1alpha1.TimeRangesWithZone{},
		Selectors: []v1alpha1.Rule{{
			Inclusions: []v1alpha1.Selector{{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{"action=delete"},
					Name:          "",
					Type:          "deployment",
					FieldSelector: nil,
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
			Exclusions: []v1alpha1.Selector{{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        nil,
					Name:          "rss-site",
					Type:          "deployment",
					FieldSelector: nil,
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
		}},
		UnHibernate: false,
		Pause:       false,
		PauseUntil:  v1alpha1.DateTimeWithZone{},
	},
	Status: v1alpha1.HibernatorStatus{},
}

var UnHibernateTest = v1alpha1.Hibernator{
	TypeMeta:   metav1.TypeMeta{
		Kind:       "hibernator",
		APIVersion: "pincher.devtron.ai/",
	},
	ObjectMeta: metav1.ObjectMeta{},
	Spec: v1alpha1.HibernatorSpec{
		When: v1alpha1.TimeRangesWithZone{},
		Selectors: []v1alpha1.Rule{{
			Inclusions: []v1alpha1.Selector{{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{"action=delete"},
					Name:          "",
					Type:          "deployment",
					FieldSelector: nil,
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
			Exclusions: []v1alpha1.Selector{{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        nil,
					Name:          "nginx",
					Type:          "deployment",
					FieldSelector: nil,
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
		}},
		UnHibernate: false,
		Pause:       false,
		PauseUntil:  v1alpha1.DateTimeWithZone{},
	},
	Status: v1alpha1.HibernatorStatus{},
}

var DeleteTest = v1alpha1.Hibernator{
	TypeMeta:   metav1.TypeMeta{
		Kind:       "hibernator",
		APIVersion: "pincher.devtron.ai/",
	},
	ObjectMeta: metav1.ObjectMeta{},
	Spec: v1alpha1.HibernatorSpec{
		When: v1alpha1.TimeRangesWithZone{},
		Action: v1alpha1.Delete,
		Selectors: []v1alpha1.Rule{{
			Inclusions: []v1alpha1.Selector{{
				ObjectSelector: v1alpha1.ObjectSelector{
					Labels:        []string{"action=delete"},
					Name:          "",
					Type:          "deployment",
					FieldSelector: nil,
				},
				NamespaceSelector: v1alpha1.NamespaceSelector{
					Name: "pras",
				},
			}},
			Exclusions: []v1alpha1.Selector{},
		}},
		UnHibernate: false,
		Pause:       false,
		PauseUntil:  v1alpha1.DateTimeWithZone{},
	},
	Status: v1alpha1.HibernatorStatus{},
}