package pkg

var PodObjectMock = `
[{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "qos-demo-5",
    "namespace": "qos-example"
  },
  "spec": {
    "containers": [
      {
        "name": "qos-demo-ctr-5",
        "image": "nginx",
        "resources": {
          "limits": {
            "memory": "200Mi",
            "cpu": "700m"
          },
          "requests": {
            "memory": "200Mi",
            "cpu": "700m"
          }
        }
      }
    ]
  }
}]
`
