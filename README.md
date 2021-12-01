# Winter Soldier
Winter Soldier can be used to

- cleans up (delete) Kubernetes resources
- reduce workload pods to 0

at user defined time of the day and conditions.
Winter Soldier is an operator which expects conditions to be defined using CRD hibernator.

## Motivation
Overtime Kubernetes clusters end up with workloads which have outlived their utility and add to the TCO of infrastructure.  Some prominent use cases are

1. Microservices in QA environment are not required during off-work hours or during weekends.
2. UAT environment is required only before releasing to production but is kept running because of time required to bring it up.
3. Workload created for POC purpose, eg - Kafka, Mongodb, SQL workload, are left running long after POC is done.

## Configurations

### Actions
Winter Soldier supports two type of actions on the workloads.
#### Delete
This action can be used to delete any Kubernetes object.
 ```yaml
 spec:
  action: delete
```
#### Sleep
This condition can be used to change replicas of workload to 0.
```yaml
spec:
  action: sleep
```
At the end of hibernation cycle it sets replica count of workload to same number as it was before hibernation.

### Conditions
Hibernator uses [gjson](https://github.com/tidwall/gjson) to select fields in Kubernetes objects and [expr](github.com/antonmedv/expr) for conditions. Please check them out for advanced cases.

Objects can be included and excluded based on
1. Label Selector
2. Object Kind
3. Name
4. Namespace
5. Any field in the kubernetes object

```yaml
selectors:
- inclusions:
  - objectSelector:
      name: ""
      type: "deployment"
      fieldSelector:
      - AfterTime(Now(), AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '10h'))
    namespaceSelector:
      name: "all"
  exclusions: 
  - objectSelector:
      name: ""
      type: "deployment"
    namespaceSelector:
      name: "kube-system"
```
The above example will select `Deployment` kind objects which have been created 10 hours ago across all namespaces excluding `kube-system` namespace. Winter soldier exposes following functions to handle time, cpu and memory.

1. ParseTime - This function can be used to parse time. For eg to parse creationTimestamp use `ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z')`
2. AddTime - This can be used to add time. For eg `AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '-10h')` ll add 10h to the time. Use `d` for day, `h` for hour, `m` for minutes and `s` for seconds. Use negative number to get earlier time.
3. Now - This can be used to get current time.
4. CpuToNumber - This can be used to compare CPU. For eg `any({{spec.containers.#.resources.requests}}, { MemoryToNumber(.memory) < MemoryToNumber('60Mi')})` will check if any `resource.requests` is less than `60Mi`

### Time Range
This defines the execution time

```yaml
spec:
  action: sleep
  timeRangesWithZone:
    reSyncInterval: 300 # in minutes
    timeZone: "Asia/Kolkata"
    timeRanges:
      - timeFrom: 00:00
        timeTo: 23:59:59
        weekdayFrom: Sat
        weekdayTo: Sun
      - timeFrom: 00:00
        timeTo: 08:00
        weekdayFrom: Mon
        weekdayTo: Fri
      - timeFrom: 20:00
        timeTo: 23:59:59
        weekdayFrom: Mon
        weekdayTo: Fri
```
Above settings will take action on Sat and Sun from 00:00 to 23:59:59, and on Mon-Fri from 00:00 to 08:00 and 20:00 to 23:59:59. If `action:sleep` then runs hibernate at `timeFrom` and unhibernate at `timeTo`.  If `action: delete` then it will delete workloads at `timeFrom` and `timeTo`.

## How to install 

### Using Kubectl 

To install winter soldier, follow the steps given below - 

**STEP 1**

Execute the following command to apply the crd.

```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/winter-soldier/main/config/crd/bases/pincher.devtron.ai_hibernators.yaml 
```

**STEP 2**

Execute the following command to install winter-soldier. 

```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/winter-soldier/main/config/hibernator_install.yaml
```

**STEP 3**

After installing winter-soldier, create your own hibernator-policies, you can refer some example policies we have included in.
Refer to the folder - [Hibernators](/config/hibernators)



**STEP 4**

Now, apply the yaml for hibernator_policies. 

```bash
kubectl apply -f hibernator.yaml
```

### Using helm


### Other Configurations
1. Pause - To pause execution
```yaml
spec:
  pause: true
```
2. PauseUntil - To pause execution
```yaml
spec:
  pauseUntil: "Jan 2, 2026 3:04pm"
```
3. Hibernate - Hibernates immediately till this flag is unset
```yaml
spec:
  hibernate: true
```
4. UnHiberbate - UnHibernates immediately till this flag is unset
```yaml
spec:
  unhibernate: true
```
** Please Note: If both hibernate and unHibernate flag are set then hibernate flag is ignored

