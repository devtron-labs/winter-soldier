package controllers

import (
	"fmt"
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func (r *HibernatorReconciler) getResourceKey(obj unstructured.Unstructured) string {
	return fmt.Sprintf("/%s/%s/%s/%s/%s", obj.GetNamespace(), obj.GroupVersionKind().Group, obj.GroupVersionKind().Version, obj.GetKind(), obj.GetName())
}

func (r *HibernatorReconciler) componentsOfResourceKey(resourceKey string) (namespace, group, version, kind, name string) {
	components := strings.Split(resourceKey, "/")
	if len(components) < 6 {
		return "", "", "", "", ""
	}
	return  components[1], components[2], components[3], components[4], components[5]
}

func (r *HibernatorReconciler) getKey(hibernator *v1alpha1.Hibernator) string {
	return fmt.Sprintf("/%s/%s/%s/%s", hibernator.Namespace, hibernator.APIVersion, hibernator.Kind, hibernator.Name)
}

func (r *HibernatorReconciler) getNamespacedName(hibernator *v1alpha1.Hibernator) string {
	return fmt.Sprintf("/%s/%s", hibernator.Namespace, hibernator.Name)
}

