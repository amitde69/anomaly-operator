# Detector Deployment and Architecture  

K8S Anomaly Detector Operator supports creating a [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) which manages a few other Kubernetes resources:

* `Deployment`
* `ConfigMap`
* `ServiceAccount`

The name of the Detector CR is propagated to the `labels` `selectors` and `name` of every resource it creates.
```yaml
...
metadata:
  name: detector-name
...
```

```bash
> kubectl get cm,deploy,sa -n anomaly-operator

NAME                              DATA   AGE
configmap/detector-name           1      5m

NAME                              READY   AGE
deployment.apps/detector-name     1/1     5m

NAME                              SECRETS   AGE
serviceaccount/detector-name      0         5m
```

## Deployment

The `Deployment` resource is desgined to run as a single replica pod.
!!! tip
    If a Detector `queries` list is too long or the quries are too heavy
    it is reccomended to split it to smaller and distributed Detector resources

## ConfigMap

The `ConfigMap` resource is injected with the detector configs passed in the Detector resource.
!!! note
    If the spec of the Detector changes the `ConfigMap` will be updated with the new config 
    and the `Deployment` will be rolled out in order to load the new `ConfigMap`.

## ServiceAccount

The `ServiceAccount` resource is created by default and cannot be changed or modified.
It is used as the `Deployment` serviceaccount and can be changed to a custom serviceaccount using the `pod_spec` 
on the Detector resource.

