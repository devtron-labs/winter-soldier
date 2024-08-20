package controllers

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"testing"
)

var hibernator_pod = `
{
  "apiVersion": "pincher.devtron.ai/v1alpha1",
  "kind": "Hibernator",
  "metadata": {
    "generation": 3,
    "labels": {
      "app.kubernetes.io/instance": "lynx-prod-winter-soldier-lynx-utils"
    },
    "name": "dd-flare-allocation-hpa-hibernator",
    "namespace": "utils"
  },
  "spec": {
    "action": "scale",
    "pauseUntil": {
      "dateTime": "",
      "timeZone": ""
    },
    "selectors": [
      {
        "inclusions": [
          {
            "namespaceSelector": {
              "name": "qos-example"
            },
            "objectSelector": {
              "name": "qos-demo-5",
              "type": "po"
            }
          }
        ]
      }
    ],
    "targetReplicas": [
      3,
      3,
      45
    ],
	"targetResources": [{
	  "qos-demo-ctr-5": {
          "limits": {
            "memory": "201Mi",
            "cpu": "701m"
          },
          "requests": {
            "memory": "201Mi",
            "cpu": "701m"
          }
	  }
	},
	{
	  "qos-demo-ctr-5": {
          "limits": {
            "memory": "201Mi"
          },
          "requests": {
            "cpu": "701m"
          }
	  }
	}],
    "timeRangesWithZone": {
      "timeRanges": [
        {
          "timeFrom": "00:00",
          "timeTo": "03:59",
          "weekdayFrom": "Mon",
          "weekdayTo": "Sun"
        },
        {
          "timeFrom": "04:00",
          "timeTo": "17:59",
          "weekdayFrom": "Mon",
          "weekdayTo": "Sun"
        }
      ],
      "timeZone": "Asia/Kolkata"
    }
  }
}
`

func TestResourceActionImpl_ScaleResourceActionFactory(t *testing.T) {
	var hb v1alpha1.Hibernator
	err := json.Unmarshal([]byte(hibernator_pod), &hb)
	if err != nil {
		panic(err)
	}
	kubectlMock := pkg.NewKubectlMock(pkg.PodObjectMock)
	resources, err := kubectlMock.GetResource(context.Background(), &pkg.GetRequest{
		Name:             "qos-demo-5",
		Namespace:        "qos-example",
		GroupVersionKind: schema.GroupVersionKind{Kind: "Pod"},
	})
	if err != nil {
		panic(err)
	}
	uns := resources.Manifest
	type fields struct {
		Kubectl     pkg.KubectlCmd
		historyUtil History
	}
	type args struct {
		hibernator *v1alpha1.Hibernator
		timeGap    v1alpha1.NearestTimeGap
		uns        unstructured.Unstructured
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Execute
	}{
		{
			name: "pod test1",
			fields: fields{
				Kubectl:     pkg.NewKubectlMock(pkg.PodObjectMock),
				historyUtil: &HistoryImpl{},
			},
			args: args{
				hibernator: &hb,
				timeGap:    v1alpha1.NearestTimeGap{MatchedIndex: 0},
				uns:        uns,
			},
		},
		{
			name: "pod test2",
			fields: fields{
				Kubectl:     pkg.NewKubectlMock(pkg.PodObjectMock),
				historyUtil: &HistoryImpl{},
			},
			args: args{
				hibernator: &hb,
				timeGap:    v1alpha1.NearestTimeGap{MatchedIndex: 1},
				uns:        uns,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceActionImpl{
				Kubectl:     tt.fields.Kubectl,
				historyUtil: tt.fields.historyUtil,
			}
			if got1, got2 := r.ScaleResourceActionFactory(tt.args.hibernator, tt.args.timeGap)([]unstructured.Unstructured{tt.args.uns}); !reflect.DeepEqual(got1, tt.want) {
				t.Errorf("ScaleResourceActionFactory() = %v, %v, want %v", got1, got2, tt.want)
			}
		})
	}
}
