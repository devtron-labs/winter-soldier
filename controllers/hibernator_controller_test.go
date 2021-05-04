package controllers

import (
	"fmt"
	"github.com/devtron-labs/winter-soldier/api/v1alpha1"
	"github.com/devtron-labs/winter-soldier/pkg"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	pincherv1alpha1 "github.com/devtron-labs/winter-soldier/api/v1alpha1"
)

func TestHibernatorReconciler_getHPATargetRefKey(t *testing.T) {
	hpaRaw := `{
    "apiVersion": "autoscaling/v1",
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
        "annotations": {
            "autoscaling.alpha.kubernetes.io/conditions": "[{\"type\":\"AbleToScale\",\"status\":\"False\",\"lastTransitionTime\":\"2021-02-05T09:05:49Z\",\"reason\":\"FailedGetScale\",\"message\":\"the HPA controller was unable to get the target's current scale: rollouts.argoproj.io \\\"test-demo-devtron\\\" not found\"}]",
            "autoscaling.alpha.kubernetes.io/metrics": "[{\"type\":\"Resource\",\"resource\":{\"name\":\"memory\",\"targetAverageUtilization\":80}}]",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"autoscaling/v2beta2\",\"kind\":\"HorizontalPodAutoscaler\",\"metadata\":{\"annotations\":{},\"labels\":{\"app.kubernetes.io/instance\":\"test-demo-devtron\"},\"name\":\"test-demo-devtron-hpa\",\"namespace\":\"demo-devtron\"},\"spec\":{\"maxReplicas\":2,\"metrics\":[{\"resource\":{\"name\":\"memory\",\"target\":{\"averageUtilization\":80,\"type\":\"Utilization\"}},\"type\":\"Resource\"},{\"resource\":{\"name\":\"cpu\",\"target\":{\"averageUtilization\":90,\"type\":\"Utilization\"}},\"type\":\"Resource\"}],\"minReplicas\":1,\"scaleTargetRef\":{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Rollout\",\"name\":\"test-demo-devtron\"}}}\n"
        },
        "creationTimestamp": "2021-02-05T09:05:34Z",
        "labels": {
            "app.kubernetes.io/instance": "test-demo-devtron"
        },
        "name": "test-demo-devtron-hpa",
        "namespace": "demo-devtron",
        "resourceVersion": "83497993",
        "selfLink": "/apis/autoscaling/v1/namespaces/demo-devtron/horizontalpodautoscalers/test-demo-devtron-hpa",
        "uid": "4dfeee64-6791-11eb-b23b-028e50ae9084"
    },
    "spec": {
        "maxReplicas": 2,
        "minReplicas": 1,
        "scaleTargetRef": {
            "apiVersion": "argoproj.io/v1alpha1",
            "kind": "Rollout",
            "name": "test-demo-devtron"
        },
        "targetCPUUtilizationPercentage": 90
    },
    "status": {
        "currentReplicas": 0,
        "desiredReplicas": 0
    }
}`
	hpa := &unstructured.Unstructured{}
	err := hpa.UnmarshalJSON([]byte(hpaRaw))
	if err != nil {
		t.Errorf("error: %v", err)
	}
	type fields struct {
		Client  client.Client
		Log     logr.Logger
		Scheme  *runtime.Scheme
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
	}
	type args struct {
		hpa unstructured.Unstructured
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "hpa test",
			fields: fields{},
			args:   args{hpa: *hpa},
			want:   fmt.Sprintf("/%s/%s/%s/%s", "demo-devtron", "argoproj.io/v1alpha1", "Rollout", "test-demo-devtron"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorReconciler{
				Client:  tt.fields.Client,
				Log:     tt.fields.Log,
				Scheme:  tt.fields.Scheme,
				kubectl: tt.fields.kubectl,
				mapper:  tt.fields.mapper,
			}
			if got := r.getHPATargetRefKey(tt.args.hpa); got != tt.want {
				t.Errorf("getHPATargetRefKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHibernatorReconciler_createHPATargetObjectMapping(t *testing.T) {
	hpaRaw := `{
    "apiVersion": "autoscaling/v1",
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
        "annotations": {
            "autoscaling.alpha.kubernetes.io/conditions": "[{\"type\":\"AbleToScale\",\"status\":\"False\",\"lastTransitionTime\":\"2021-02-05T09:05:49Z\",\"reason\":\"FailedGetScale\",\"message\":\"the HPA controller was unable to get the target's current scale: rollouts.argoproj.io \\\"test-demo-devtron\\\" not found\"}]",
            "autoscaling.alpha.kubernetes.io/metrics": "[{\"type\":\"Resource\",\"resource\":{\"name\":\"memory\",\"targetAverageUtilization\":80}}]",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"autoscaling/v2beta2\",\"kind\":\"HorizontalPodAutoscaler\",\"metadata\":{\"annotations\":{},\"labels\":{\"app.kubernetes.io/instance\":\"test-demo-devtron\"},\"name\":\"test-demo-devtron-hpa\",\"namespace\":\"demo-devtron\"},\"spec\":{\"maxReplicas\":2,\"metrics\":[{\"resource\":{\"name\":\"memory\",\"target\":{\"averageUtilization\":80,\"type\":\"Utilization\"}},\"type\":\"Resource\"},{\"resource\":{\"name\":\"cpu\",\"target\":{\"averageUtilization\":90,\"type\":\"Utilization\"}},\"type\":\"Resource\"}],\"minReplicas\":1,\"scaleTargetRef\":{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Rollout\",\"name\":\"test-demo-devtron\"}}}\n"
        },
        "creationTimestamp": "2021-02-05T09:05:34Z",
        "labels": {
            "app.kubernetes.io/instance": "test-demo-devtron"
        },
        "name": "test-demo-devtron-hpa",
        "namespace": "demo-devtron",
        "resourceVersion": "83497993",
        "selfLink": "/apis/autoscaling/v1/namespaces/demo-devtron/horizontalpodautoscalers/test-demo-devtron-hpa",
        "uid": "4dfeee64-6791-11eb-b23b-028e50ae9084"
    },
    "spec": {
        "maxReplicas": 2,
        "minReplicas": 1,
        "scaleTargetRef": {
            "apiVersion": "argoproj.io/v1alpha1",
            "kind": "Rollout",
            "name": "test-demo-devtron"
        },
        "targetCPUUtilizationPercentage": 90
    },
    "status": {
        "currentReplicas": 0,
        "desiredReplicas": 0
    }
}`
	ro1Raw := `{
  "apiVersion": "argoproj.io/v1alpha1",
  "kind": "Rollout",
  "metadata": {
    "creationTimestamp": "2021-01-27T06:52:41Z",
    "generation": 1,
    "labels": {
      "app": "dashboard",
      "app.kubernetes.io/instance": "dashboard-demo-devtroncd",
      "chart": "dashboard-3.5.1",
      "pipelineName": "demo-cd-pipeline",
      "release": "dashboard-demo-devtroncd",
      "releaseVersion": "223"
    },
    "name": "test-demo-devtron",
    "namespace": "demo-devtron"
  },
  "spec": {
    "minReadySeconds": 60,
    "replicas": 1,
    "revisionHistoryLimit": 3,
    "selector": {
      "matchLabels": {
        "app": "dashboard",
        "release": "dashboard-demo-devtroncd"
      }
    },
    "strategy": {
      "canary": {
        "maxSurge": "25%",
        "maxUnavailable": 1,
        "stableService": "dashboard-demo-devtroncd-service"
      }
    },
    "template": {
      "metadata": {
        "creationTimestamp": null,
        "labels": {
          "app": "dashboard",
          "appId": "1",
          "envId": "7",
          "release": "dashboard-demo-devtroncd"
        }
      },
      "spec": {
        "containers": [
          {
            "env": [
              {
                "name": "CONFIG_HASH",
                "value": "d1bc91092301de1ffeb881f38b8ccf0b726d1878a124dfa6cfafa686c6125f88"
              },
              {
                "name": "SECRET_HASH",
                "value": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
              },
              {
                "name": "POD_NAME",
                "valueFrom": {
                  "fieldRef": {
                    "fieldPath": "metadata.name"
                  }
                }
              }
            ],
            "envFrom": [
              {
                "configMapRef": {
                  "name": "dashboard-cm-1"
                }
              }
            ],
            "image": "686244538589.dkr.ecr.us-east-2.amazonaws.com/release/dashboard:4222ce1e-51-4118",
            "imagePullPolicy": "IfNotPresent",
            "name": "dashboard",
            "ports": [
              {
                "containerPort": 80,
                "name": "app",
                "protocol": "TCP"
              }
            ],
            "resources": {
              "limits": {
                "cpu": "100m",
                "memory": "200Mi"
              },
              "requests": {
                "cpu": "10m",
                "memory": "100Mi"
              }
            }
          }
        ],
        "restartPolicy": "Always",
        "terminationGracePeriodSeconds": 30
      }
    }
  }
}`
	ro2Raw := `{
  "apiVersion": "argoproj.io/v1alpha1",
  "kind": "Rollout",
  "metadata": {
    "creationTimestamp": "2021-01-27T06:52:41Z",
    "generation": 1,
    "labels": {
      "app": "dashboard",
      "app.kubernetes.io/instance": "dashboard-demo-devtroncd",
      "chart": "dashboard-3.5.1",
      "pipelineName": "demo-cd-pipeline",
      "release": "dashboard-demo-devtroncd",
      "releaseVersion": "223"
    },
    "name": "test-demo-devtron2",
    "namespace": "demo-devtron"
  },
  "spec": {
    "minReadySeconds": 60,
    "replicas": 1,
    "revisionHistoryLimit": 3,
    "selector": {
      "matchLabels": {
        "app": "dashboard",
        "release": "dashboard-demo-devtroncd"
      }
    },
    "strategy": {
      "canary": {
        "maxSurge": "25%",
        "maxUnavailable": 1,
        "stableService": "dashboard-demo-devtroncd-service"
      }
    },
    "template": {
      "metadata": {
        "creationTimestamp": null,
        "labels": {
          "app": "dashboard",
          "appId": "1",
          "envId": "7",
          "release": "dashboard-demo-devtroncd"
        }
      },
      "spec": {
        "containers": [
          {
            "env": [
              {
                "name": "CONFIG_HASH",
                "value": "d1bc91092301de1ffeb881f38b8ccf0b726d1878a124dfa6cfafa686c6125f88"
              },
              {
                "name": "SECRET_HASH",
                "value": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
              },
              {
                "name": "POD_NAME",
                "valueFrom": {
                  "fieldRef": {
                    "fieldPath": "metadata.name"
                  }
                }
              }
            ],
            "envFrom": [
              {
                "configMapRef": {
                  "name": "dashboard-cm-1"
                }
              }
            ],
            "image": "686244538589.dkr.ecr.us-east-2.amazonaws.com/release/dashboard:4222ce1e-51-4118",
            "imagePullPolicy": "IfNotPresent",
            "name": "dashboard",
            "ports": [
              {
                "containerPort": 80,
                "name": "app",
                "protocol": "TCP"
              }
            ],
            "resources": {
              "limits": {
                "cpu": "100m",
                "memory": "200Mi"
              },
              "requests": {
                "cpu": "10m",
                "memory": "100Mi"
              }
            }
          }
        ],
        "restartPolicy": "Always",
        "terminationGracePeriodSeconds": 30
      }
    }
  }
}`
	hpa := &unstructured.Unstructured{}
	err := hpa.UnmarshalJSON([]byte(hpaRaw))
	if err != nil {
		t.Errorf("error: %v", err)
	}
	ro1 := &unstructured.Unstructured{}
	err = ro1.UnmarshalJSON([]byte(ro1Raw))
	if err != nil {
		t.Errorf("error: %v", err)
	}
	ro2 := &unstructured.Unstructured{}
	err = ro2.UnmarshalJSON([]byte(ro2Raw))
	if err != nil {
		t.Errorf("error: %v", err)
	}
	type fields struct {
		Client  client.Client
		Log     logr.Logger
		Scheme  *runtime.Scheme
		kubectl pkg.KubectlCmd
		mapper  *pkg.Mapper
	}
	type args struct {
		targetObjects []unstructured.Unstructured
		hpas          []unstructured.Unstructured
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []TargetObjectHPAPair
	}{
		{
			name: "matching ",
			args: args{
				targetObjects: []unstructured.Unstructured{*ro1, *ro2},
				hpas:          []unstructured.Unstructured{*hpa},
			},
			fields: fields{},
			want: []TargetObjectHPAPair{{
				TargetObject: ro1,
				HPA:          hpa,
			},
				{
					TargetObject: ro2,
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorReconciler{
				Client:  tt.fields.Client,
				Log:     tt.fields.Log,
				Scheme:  tt.fields.Scheme,
				kubectl: tt.fields.kubectl,
				mapper:  tt.fields.mapper,
			}
			if got := r.createHPAAndTargetObjectMapping(tt.args.hpas, tt.args.targetObjects); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createHPAAndTargetObjectMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHibernatorReconciler_hibernate(t *testing.T) {
	hibernator := pincherv1alpha1.Hibernator{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       pincherv1alpha1.HibernatorSpec{
			TimeRangesWithZone:           pincherv1alpha1.TimeRangesWithZone{},
			Rules:                        []pincherv1alpha1.Rule{{
				Inclusions:  []pincherv1alpha1.Selector{{
					Labels:        []string{"app=delete"},
					Name:          "",
					Namespace:     "pras",
					Type:          "deployment",
					FieldSelector: nil,
				}},
				Exclusions:  []pincherv1alpha1.Selector{{
					Labels:        nil,
					Name:          "patch-demo",
					Namespace:     "pras",
					Type:          "deployment",
					FieldSelector: nil,
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
			name: "hibernate test",
			fields: fields{
				Client:  nil,
				Log:     nil,
				Scheme:  nil,
				kubectl: pkg.NewKubectl(),
				mapper:  pkg.NewMapperFactory(),
			},
			args: args{hibernator: hibernator},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HibernatorReconciler{
				Client:  tt.fields.Client,
				Log:     tt.fields.Log,
				Scheme:  tt.fields.Scheme,
				kubectl: tt.fields.kubectl,
				mapper:  tt.fields.mapper,
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
