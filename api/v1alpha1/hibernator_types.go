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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HibernatorSpec defines the desired state of Hibernator
type HibernatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TimeRangesWithZone           TimeRangesWithZone `json:"time_ranges_with_zone,omitempty"`
	Rules                        []Rule             `json:"rules"`
	UnHibernate                  bool               `json:"reset,omitempty"`
	CanUnHibernateObjectManually bool               `json:"can_un_hibernate_object_manually,omitempty"`
	Pause                        bool               `json:"pause,omitempty"`
	PauseUntil                   TimeRange          `json:"pause_until,omitempty"`
}

type Rule struct {
	Inclusions  []Selector `json:"inclusions"`
	Exclusions  []Selector `json:"exclusions,omitempty"`
	Action      `json:"action"`
	DeleteStore bool `json:"delete_store,omitempty"`
}

type TimeRangesWithZone struct {
	TimeRanges []TimeRange `json:"time_ranges"`
	TimeZone   string      `json:"timezone,omitempty"`
}

type TimeRange struct {
	TimeZone       string  `json:"timezone,omitempty"`
	TimeFrom       string  `json:"time_from"`
	TimeTo         string  `json:"time_to"`
	CronExpression string  `json:"cron_expression,omitempty"`
	WeekdayFrom    Weekday `json:"weekday_from"`
	WeekdayTo      Weekday `json:"weekday_to"`
}

type Selector struct {
	ObjectSelector    ObjectSelector    `json:"object_selector"`
	NamespaceSelector NamespaceSelector `json:"namespace_selector,omitempty"`
	//Labels            []string          `json:"labels,omitempty"`
	//Name              string            `json:"name,omitempty"`
	//Namespace         string            `json:"namespace"`
	//Type              string            `json:"type"`
	//FieldSelector     []string          `json:"field_selector,omitempty"`
}

type ObjectSelector struct {
	Labels        []string `json:"labels,omitempty"`
	Name          string   `json:"name,omitempty"`
	Type          string   `json:"type"`
	FieldSelector []string `json:"field_selector,omitempty"`
}

type NamespaceSelector struct {
	Labels        []string `json:"labels,omitempty"`
	Name          string   `json:"name,omitempty"`
	FieldSelector []string `json:"field_selector,omitempty"`
}

// HibernatorStatus defines the observed state of Hibernator
type HibernatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	History       []RevisionHistory `json:"history"`
	Status        string            `json:"status"`
	Message       string            `json:"message"`
	IsHibernating bool              `json:"is_hibernating"`
}

type ImpactedObject struct {
	Group                string `json:"group"`
	Version              string `json:"version"`
	Kind                 string `json:"kind"`
	Name                 string `json:"name"`
	Namespace            string `json:"namespace"`
	OriginalCount        int64  `json:"original_count"`
	RelatedDeletedObject string `json:"related_deleted_object"`
	Message              string `json:"message"`
	Status               string `json:"status"`
}

type ExcludedObject struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Reason    string `json:"reason"`
}

type RevisionHistory struct {
	Time            metav1.Time      `json:"time"`
	ID              int64            `json:"id"`
	Hibernate       bool             `json:"hibernate"`
	ImpactedObjects []ImpactedObject `json:"impacted_objects"`
	ExcludedObjects []ExcludedObject `json:"excluded_objects"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Hibernator is the Schema for the hibernators API
type Hibernator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HibernatorSpec   `json:"spec,omitempty"`
	Status HibernatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HibernatorList contains a list of Hibernator
type HibernatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hibernator `json:"items"`
}

type Action string

const (
	Delete Action = "delete"
	Sleep  Action = "sleep"
)

type Weekday string

const (
	Sun = "Sun"
	Mon = "Mon"
	Tue = "Tue"
	Wed = "Wed"
	Thu = "Thu"
	Fri = "Fri"
	Sat = "Sat"
)

func init() {
	SchemeBuilder.Register(&Hibernator{}, &HibernatorList{})
}
