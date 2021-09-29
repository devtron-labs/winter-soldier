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
	"testing"
)

const pod = `
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "frontend"
  },
  "spec": {
    "containers": [
      {
        "name": "app",
        "image": "images.my-company.example/app:v4",
        "resources": {
          "requests": {
            "memory": "64Mi",
            "cpu": "250m"
          },
          "limits": {
            "memory": "128Mi",
            "cpu": "500m"
          }
        }
      },
      {
        "name": "log-aggregator",
        "image": "images.my-company.example/log-aggregator:v6",
        "resources": {
          "requests": {
            "memory": "68Mi",
            "cpu": "250m"
          },
          "limits": {
            "memory": "128Mi",
            "cpu": "500m"
          }
        }
      }
    ]
  }
}`

const deployment = `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "nginx-deployment",
    "labels": {
      "app": "nginx"
    }
  },
  "spec": {
    "replicas": 3,
    "selector": {
      "matchLabels": {
        "app": "nginx"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "app": "nginx"
        }
      },
      "spec": {
        "containers": [
          {
            "name": "nginx",
            "image": "nginx:1.14.2",
            "ports": [
              {
                "containerPort": 80
              }
            ]
          }
        ]
      }
    }
  }
}`

const ro = `
{
  "apiVersion": "argoproj.io/v1alpha1",
  "kind": "Rollout",
  "metadata": {
    "creationTimestamp": "2021-09-09T07:56:32Z",
    "managedFields": [
      {
        "time": "2021-08-31T08:32:48Z"
      }
    ],
    "name": "nitish-new-demo3",
    "namespace": "demo3"
  },
  "spec": {
    "minReadySeconds": 10,
    "replicas": 1,
    "revisionHistoryLimit": 3,
    "selector": {
      "matchLabels": {
        "app": "nitish-new",
        "release": "nitish-new-demo3"
      }
    },
    "template": {
      "spec": {
        "containers": [
          {
            "command": [
              "/usr/local/bin/envoy"
            ],
            "image": "envoyproxy/envoy:v1.14.1",
            "name": "envoy",
            "ports": [
              {
                "containerPort": 9901,
                "name": "envoy-admin",
                "protocol": "TCP"
              },
              {
                "containerPort": 8799,
                "name": "app",
                "protocol": "TCP"
              }
            ],
            "resources": {
              "limits": {
                "cpu": "50m",
                "memory": "100Mi"
              },
              "requests": {
                "cpu": "50m",
                "memory": "10Mi"
              }
            },
            "volumeMounts": [
              {
                "mountPath": "/etc/envoy-config/",
                "name": "envoy-config-volume"
              }
            ]
          },
          {
            "image": "686244538589.dkr.ecr.us-east-2.amazonaws.com/devtron:3f1262d4-20-247",
            "imagePullPolicy": "IfNotPresent",
            "name": "nitish-new",
            "ports": [
              {
                "containerPort": 8000,
                "name": "app",
                "protocol": "TCP"
              }
            ],
            "resources": {
              "limits": {
                "cpu": "50m",
                "memory": "100Mi"
              },
              "requests": {
                "cpu": "50m",
                "memory": "10Mi"
              }
            },
            "volumeMounts": []
          }
        ],
        "restartPolicy": "Always",
        "terminationGracePeriodSeconds": 30,
        "volumes": [
          {
            "configMap": {
              "name": "sidecar-config-nitish-new"
            },
            "name": "envoy-config-volume"
          }
        ]
      }
    }
  }
}`

func TestExpressionEvaluator(t *testing.T) {

	type args struct {
		expression string
		json       string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "base test",
			args: args{
				expression: "{{spec.replicas}} + {{spec.template.spec.containers.0.ports.0.containerPort}} == 80",
				json:       deployment,
			},
			want: false,
		},
		{
			name: "string match",
			args: args{
				expression: "{{spec.template.spec.containers.0.name}} == 'nginx'",
				json:       deployment,
			},
			want: true,
		},
		{
			name: "array match",
			args: args{
				expression: "any({{spec.containers.#.resources.requests}}, { MemoryToNumber(.memory) > MemoryToNumber('68M')})",
				json:       pod,
			},
			want: true,
		},
		{
			name: "array match",
			args: args{
				expression: "any({{spec.containers.#.resources.requests}}, { MemoryToNumber(.memory) == MemoryToNumber('68Mi')})",
				json:       pod,
			},
			want: true,
		},
		{
			name: "array match",
			args: args{
				expression: "any({{spec.containers.#.resources.requests}}, { MemoryToNumber(.memory) < MemoryToNumber('60Mi')})",
				json:       pod,
			},
			want: false,
		},
		{
			name: "and or match",
			args: args{
				expression: "( {{spec.containers.0.name}} == 'app' && {{spec.containers.1.name}} == 'log-aggregator' ) || {{spec.containers.1.image}} == 'log-aggregator' ",
				json:       pod,
			},
			want: true,
		},
		{
			name: "time check",
			args: args{
				expression: "AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '20d'), AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '19d'))",
				json:       ro,
			},
			want: true,
		},
		{
			name: "time check",
			args: args{
				expression: "AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '-1d'), ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'))",
				json:       ro,
			},
			want: false,
		},

		{
			name: "time check",
			args: args{
				expression: "AfterTime(AddTime(Now(), '-1d'), Now())",
				json:       ro,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpressionEvaluator(tt.args.expression, tt.args.json); got != tt.want {
				t.Errorf("ExpressionEvaluator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertMemory(t *testing.T) {
	type args struct {
		memory string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name:    "base test",
			args:    args{memory: "1Gi"},
			want:    float64(1073741824),
			wantErr: false,
		},
		{
			name:    "base test - scientifc notation",
			args:    args{memory: "1e2G"},
			want:    float64(100 * 1000000000),
			wantErr: false,
		},
		{
			name:    "base test - scientifc notation",
			args:    args{memory: "1e2"},
			want:    float64(100),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MemoryToNumber(tt.args.memory)
			if (err != nil) != tt.wantErr {
				t.Errorf("memory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("memory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
