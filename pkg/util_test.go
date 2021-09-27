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
			args: args {
				expression: "{{spec.replicas}} + {{spec.template.spec.containers.0.ports.0.containerPort}} == 80",
				json: deployment,
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
				expression: "any({{spec.containers.#.resources.requests}}, { .memory == '68Mi'})",
				json:       pod,
			},
			want: true,
		},
		{
			name: "and or match",
			args: args{
				expression: "( {{spec.containers.0.name}} == 'app' && {{spec.containers.1.name}} == 'log-aggregator' ) || {{spec.containers.1.image}} == 'log-aggregator' ",
				json:       pod,
			},
			want: true,
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