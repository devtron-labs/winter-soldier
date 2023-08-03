package controllers

const hibernator_mock_1 = `
{
        "apiVersion": "pincher.devtron.ai/v1alpha1",
        "kind": "Hibernator",
        "metadata": {
          "name": "coreos-consumer-hibernator"
        },
        "spec": {
          "action": "sleep",
          "selectors": [
            {
              "inclusions": [
                {
                  "namespaceSelector": {
                    "name": "coreos"
                  },
                  "objectSelector": {
                    "name": "",
                    "type": "Rollout"
                  }
                }
              ]
            }
          ],
          "timeRangesWithZone": {
            "timeRanges": [
              {
                "timeFrom": "00:22",
                "timeTo": "14:00",
                "weekdayFrom": "Fri",
                "weekdayTo": "Fri"
              }
            ],
            "timeZone": "Asia/Kolkata"
          }
        }
      }
`
