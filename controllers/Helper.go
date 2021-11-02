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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func getResourceKey(obj unstructured.Unstructured) string {
	return fmt.Sprintf("/%s/%s/%s/%s/%s", obj.GetNamespace(), obj.GroupVersionKind().Group, obj.GroupVersionKind().Version, obj.GetKind(), obj.GetName())
}

func componentsOfResourceKey(resourceKey string) (namespace, group, version, kind, name string) {
	components := strings.Split(resourceKey, "/")
	if len(components) < 6 {
		return "", "", "", "", ""
	}
	return components[1], components[2], components[3], components[4], components[5]
}

func getKey(hibernator *v1alpha1.Hibernator) string {
	return fmt.Sprintf("/%s/%s/%s/%s", hibernator.Namespace, hibernator.APIVersion, hibernator.Kind, hibernator.Name)
}

func getNamespacedName(hibernator *v1alpha1.Hibernator) string {
	return fmt.Sprintf("/%s/%s", hibernator.Namespace, hibernator.Name)
}
