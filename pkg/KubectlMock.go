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

package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

type kubectlMock struct {
	db   map[string]unstructured.Unstructured
}

func NewKubectlMock(bulk string) KubectlCmd {
	k := &kubectlMock{db: make(map[string]unstructured.Unstructured, 0)}
	k.bulkAdd(bulk)
	return k
}

func (k *kubectlMock) bulkAdd(bulk string) error {
	ll := make([]map[string]interface{}, 0)
	err := json.Unmarshal([]byte(bulk), &ll)
	if err != nil {
		return err
	}
	for _, l := range ll {
		o := unstructured.Unstructured{}
		o.SetUnstructuredContent(l)
		k.db[k.key(o)] = o
	}
	return nil
}

func (k *kubectlMock) add(obj string) error {
	k8sObj := unstructured.Unstructured{}
	err := k8sObj.UnmarshalJSON([]byte(obj))
	if err != nil {
		return err
	}
	k.db[k.key(k8sObj)] = k8sObj
	return nil
}

func (k *kubectlMock) key(obj unstructured.Unstructured) string {
	return fmt.Sprintf("/%s/%s/%s", obj.GetNamespace(), obj.GetKind(), obj.GetName())
}

func (k *kubectlMock) ListResources(ctx context.Context, r *ListRequest) (*ListResponse, error) {
	var response []unstructured.Unstructured
	labels := strings.Split(r.LabelSelector, ",")
	labelSelectors := make(map[string]string, 0)
	for _, label := range labels {
		if label == "" {
			continue
		}
		keyVal := strings.Split(label, "=")
		labelSelectors[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
	}
	for _, item := range k.db {
		found := true
		for key, val := range labelSelectors {
			labels := item.GetLabels()
			if labels[key] != val {
				found = false
			}
		}
		if item.GetNamespace() != r.Namespace {
			found = false
		}
		if !found {
			continue
		}
		response = append(response, item)
	}
	return &ListResponse{Manifests: response}, nil
}

func (k *kubectlMock) GetResource(ctx context.Context, r *GetRequest) (*ManifestResponse, error) {
	key := fmt.Sprintf("/%s/%s/%s", r.Namespace, r.GroupVersionKind.Kind, r.Name)
	obj := k.db[key]
	return &ManifestResponse{obj}, nil
}

func (k *kubectlMock) DeleteResource(ctx context.Context, r *DeleteRequest) (*ManifestResponse, error) {
	key := fmt.Sprintf("/%s/%s/%s", r.Namespace, r.GroupVersionKind.Kind, r.Name)
	obj := k.db[key]
	delete(k.db, key)
	return &ManifestResponse{obj}, nil
}

func (k *kubectlMock) PatchResource(ctx context.Context, r *PatchRequest) (*ManifestResponse, error) {
	key := fmt.Sprintf("/%s/%s/%s", r.Namespace, r.GroupVersionKind.Kind, r.Name)


	var k8sObj unstructured.Unstructured
	ok := false
	if k8sObj, ok = k.db[key]; !ok {
		return nil, fmt.Errorf("object not found")
	}
	patchJSON := []byte(r.Patch)

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		panic(err)
	}
	obj, err := k8sObj.MarshalJSON()
	modified, err := patch.Apply(obj)
	if err != nil {
		return nil, err
	}
	response := unstructured.Unstructured{}
	err = response.UnmarshalJSON(modified)
	k.db[k.key(response)] = response
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{response}, nil
}
