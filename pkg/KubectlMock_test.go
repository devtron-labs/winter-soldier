package pkg

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func Test_kubectlMock_DeleteResource(t *testing.T) {
	type fields struct {
		db map[string]unstructured.Unstructured
	}
	type args struct {
		ctx context.Context
		r   *DeleteRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ManifestResponse
		wantErr bool
	}{
		{
			name:   "base delete case",
			fields: fields{db: make(map[string]unstructured.Unstructured, 0)},
			args: args{
				ctx: context.Background(),
				r: &DeleteRequest{
					Name:             "nginx-deployment",
					Namespace:        "pras",
					GroupVersionKind: schema.GroupVersionKind{Kind: "Deployment"},
					Force:            nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kubectlMock{
				db: tt.fields.db,
			}
			err := k.bulkAdd(DeploymentObjectsMock)
			if err != nil {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := k.DeleteResource(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			exists, err := k.GetResource(tt.args.ctx, &GetRequest{
				Name:             tt.args.r.Name,
				Namespace:        tt.args.r.Namespace,
				GroupVersionKind: tt.args.r.GroupVersionKind,
			})
			if err != nil || exists.Manifest.Object != nil {
				t.Errorf("DeleteResource() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_kubectlMock_GetResource(t *testing.T) {
	type fields struct {
		db map[string]unstructured.Unstructured
	}
	type args struct {
		ctx context.Context
		r   *GetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ManifestResponse
		wantErr bool
	}{
		{
			name:   "simple fetch",
			fields: fields{db: make(map[string]unstructured.Unstructured, 0)},
			args: args{
				r:   &GetRequest{Name: "nginx", Namespace: "pras", GroupVersionKind: schema.GroupVersionKind{Kind: "Deployment"}},
				ctx: context.Background()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kubectlMock{
				db: tt.fields.db,
			}
			err := k.bulkAdd(DeploymentObjectsMock)
			if err != nil {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := k.GetResource(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.args.r.Namespace != got.Manifest.GetNamespace() || tt.args.r.Name != got.Manifest.GetName() || tt.args.r.GroupVersionKind.Kind != got.Manifest.GetKind() {
				t.Errorf("GetResource() got = %v, want %v", got, tt.want)
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetResource() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func Test_kubectlMock_ListResources(t *testing.T) {
	type fields struct {
		db map[string]unstructured.Unstructured
	}
	type args struct {
		ctx context.Context
		r   *ListRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{
			name:   "base listing case",
			fields: fields{db: make(map[string]unstructured.Unstructured, 0)},
			args: args{
				r: &ListRequest{
					Namespace:            "pras",
					GroupVersionResource: schema.GroupVersionResource{Resource: "deployment"},
					ListOptions:          metav1.ListOptions{LabelSelector: "action=delete"},
				},
				ctx: context.Background()},
			wantErr: false,
			want: map[string]bool{
				fmt.Sprintf("/%s/%s/%s", "pras", "Deployment", "nginx-deployment"): true,
				fmt.Sprintf("/%s/%s/%s", "pras", "Deployment", "rss-site"):         true,
			},
		},
		{
			name:   "base double selector listing case",
			fields: fields{db: make(map[string]unstructured.Unstructured, 0)},
			args: args{
				r: &ListRequest{
					Namespace:            "pras",
					GroupVersionResource: schema.GroupVersionResource{Resource: "deployment"},
					ListOptions:          metav1.ListOptions{LabelSelector: "action=delete,type=nginx"},
				},
				ctx: context.Background()},
			wantErr: false,
			want: map[string]bool{
				fmt.Sprintf("/%s/%s/%s", "pras", "Deployment", "nginx-deployment"): true,
				fmt.Sprintf("/%s/%s/%s", "pras", "Deployment", "rss-site"):         true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kubectlMock{
				db: tt.fields.db,
			}
			err := k.bulkAdd(DeploymentObjectsMock)
			if err != nil {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := k.ListResources(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, manifest := range got.Manifests {
				if !tt.want[k.key(manifest)] {
					t.Errorf("ListResources() got = %v, want %v", k.key(manifest), tt.want)
				}
			}
		})
	}
}

func Test_kubectlMock_PatchResource(t *testing.T) {
	patch := `[{"op": "replace", "path": "/spec/replicas", "value":0}, {"op": "add", "path": "/metadata/annotations", "value": {"hibernator.devtron.ai/replicas":"2"}}]`
	type fields struct {
		db map[string]unstructured.Unstructured
	}
	type args struct {
		ctx context.Context
		r   *PatchRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ManifestResponse
		wantErr bool
	}{
		{
			name:   "base case of patch",
			fields: fields{db: make(map[string]unstructured.Unstructured, 0)},
			args: args{
				ctx: context.Background(),
				r: &PatchRequest{
					Name:             "nginx-deployment",
					Namespace:        "pras",
					GroupVersionKind: schema.GroupVersionKind{Kind: "Deployment"},
					Patch:            patch,
					PatchType:        "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kubectlMock{
				db: tt.fields.db,
			}
			err := k.bulkAdd(DeploymentObjectsMock)
			if err != nil {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := k.PatchResource(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			exists, err := k.GetResource(tt.args.ctx, &GetRequest{
				Name:             tt.args.r.Name,
				Namespace:        tt.args.r.Namespace,
				GroupVersionKind: tt.args.r.GroupVersionKind,
			})
			//if err != nil || exists.Manifest.Object != nil {
			//	t.Errorf("DeleteResource() got = %v, want %v", got, tt.want)
			//}
			if err != nil || exists.Manifest.Object == nil {
				t.Errorf("PatchResource() got = %v, want %v", got, tt.want)
			}
		})
	}
}
