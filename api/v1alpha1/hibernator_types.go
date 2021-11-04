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
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HibernatorSpec defines the desired state of Hibernator
type HibernatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	When                 TimeRangesWithZone `json:"timeRangesWithZone,omitempty"`
	Selectors            []Rule             `json:"selectors"`
	Hibernate            bool               `json:"hibernate,omitempty"`
	UnHibernate          bool               `json:"unHibernate,omitempty"`
	ReSyncInterval       int                `json:"reSyncInterval,omitempty"`
	Pause                bool               `json:"pause,omitempty"`
	PauseUntil           DateTimeWithZone   `json:"pauseUntil,omitempty"`
	RevisionHistoryLimit *int               `json:"revisionHistoryLimit,omitempty"`
	Action               Action             `json:"action"`
	DeleteStore          bool               `json:"deleteStore,omitempty"`
}

type Rule struct {
	Inclusions []Selector `json:"inclusions"`
	Exclusions []Selector `json:"exclusions,omitempty"`
}

type DateTimeWithZone struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type TimeRangesWithZone struct {
	TimeRanges []TimeRange `json:"timeRanges"`
	TimeZone   string      `json:"timeZone,omitempty"`
}

type TimeRange struct {
	TimeZone           string  `json:"timeZone,omitempty"`
	TimeFrom           string  `json:"timeFrom"`
	TimeTo             string  `json:"timeTo"`
	CronExpressionFrom string  `json:"cronExpressionFrom,omitempty"`
	CronExpressionTo   string  `json:"cronExpressionTo,omitempty"`
	WeekdayFrom        Weekday `json:"weekdayFrom"`
	WeekdayTo          Weekday `json:"weekdayTo"`
}

type Selector struct {
	ObjectSelector    ObjectSelector    `json:"objectSelector"`
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
}

type ObjectSelector struct {
	Labels        []string `json:"labels,omitempty"`
	Name          string   `json:"name,omitempty"`
	Type          string   `json:"type"`
	FieldSelector []string `json:"fieldSelector,omitempty"`
}

type NamespaceSelector struct {
	Labels        []string `json:"labels,omitempty"`
	Name          string   `json:"name,omitempty"`
	FieldSelector []string `json:"fieldSelector,omitempty"`
}

// HibernatorStatus defines the observed state of Hibernator
type HibernatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	History       []RevisionHistory `json:"history"`
	Status        string            `json:"status"`
	Message       string            `json:"message"`
	IsHibernating bool              `json:"isHibernating"`
	Action        Action            `json:"action"`
}

type ImpactedObject struct {
	ResourceKey string `json:"resourceKey"`
	//Group                string `json:"group"`
	//Version              string `json:"version"`
	//Kind                 string `json:"kind"`
	//Name                 string `json:"name"`
	//Namespace            string `json:"namespace"`
	OriginalCount        int    `json:"originalCount"`
	RelatedDeletedObject string `json:"relatedDeletedObject"`
	Message              string `json:"message"`
	Status               string `json:"status"`
}

type ExcludedObject struct {
	ResourceKey string `json:"resourceKey"`
	//Group       string `json:"group"`
	//Version     string `json:"version"`
	//Kind        string `json:"kind"`
	//Name        string `json:"name"`
	//Namespace   string `json:"namespace"`
	Reason string `json:"reason"`
}

type RevisionHistory struct {
	Time            metaV1.Time      `json:"time"`
	ID              int64            `json:"id"`
	Action          Action           `json:"action"`
	ImpactedObjects []ImpactedObject `json:"impactedObjects"`
	ExcludedObjects []ExcludedObject `json:"excludedObjects"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Hibernator is the Schema for the hibernators API
type Hibernator struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HibernatorSpec   `json:"spec,omitempty"`
	Status HibernatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HibernatorList contains a list of Hibernator
type HibernatorList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []Hibernator `json:"items"`
}

type Action string

const (
	Delete      Action = "delete"
	Hibernate   Action = "hibernate"
	UnHibernate Action = "unhibernate"
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
