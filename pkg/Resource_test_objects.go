package pkg

const DeploymentObjectsMock = `
[
{
  "kind": "Deployment",
  "apiVersion": "extensions/v1beta1",
  "metadata": {
    "name": "nginx-deployment",
    "labels": {
      "action": "delete",
      "type": "nginx"
    },
    "namespace": "pras"
  },
  "spec": {
    "replicas": 3,
    "strategy": "Recreate",
    "selector": {
      "matchLabels": {
        "deploy": "example"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "deploy": "example"
        }
      },
      "spec": {
        "containers": [
          {
            "name": "nginx",
            "image": "nginx:1.7.9"
          }
        ]
      }
    }
  }
},
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "rss-site",
    "labels": {
      "app": "web",
	  "action": "delete",
      "type": "nginx"
    },
    "annotations": {
      "hibernator.devtron.ai/replicas":"2"
	},
    "namespace": "pras"
  },
  "spec": {
    "replicas": 0,
    "selector": {
      "matchLabels": {
        "app": "web"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "app": "web"
        }
      },
      "spec": {
        "containers": [
          {
            "name": "front-end",
            "image": "nginx",
            "ports": [
              {
                "containerPort": 80
              }
            ]
          },
          {
            "name": "rss-reader",
            "image": "nickchase/rss-php-nginx:v1",
            "ports": [
              {
                "containerPort": 88
              }
            ]
          }
        ]
      }
    }
  }
},
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "nginx",
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
            "image": "nginx:1.7.9",
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
},
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "nginx",
    "labels": {
      "app": "nginx"
    },
    "namespace": "pras"
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
            "image": "nginx",
            "name": "nginx",
            "ports": [
              {
                "containerPort": 8080
              }
            ]
          }
        ]
      }
    }
  }
},
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "postgres",
    "labels": {
	  "action": "delete"
	}
  },
  "spec": {
    "selector": {
      "matchLabels": {
        "app": "postgres"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "app": "postgres"
        }
      },
      "spec": {
        "containers": [
          {
            "name": "postgres",
            "image": "postgres",
            "ports": [
              {
                "containerPort": 5432
              }
            ],
            "env": [
              {
                "name": "POSTGRES_DB",
                "value": "mydatabase"
              },
              {
                "name": "POSTGRES_USER",
                "valueFrom": {
                  "configMapKeyRef": {
                    "name": "postgres-config",
                    "key": "my.username"
                  }
                }
              },
              {
                "name": "POSTGRES_PASSWORD",
                "valueFrom": {
                  "secretKeyRef": {
                    "name": "postgres-secret",
                    "key": "secret.password"
                  }
                }
              }
            ]
          }
        ]
      }
    }
  }
}
]`
