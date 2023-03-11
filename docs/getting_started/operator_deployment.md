# K8S Anomaly Detection Operator Deployment
## Deploy Helm Chart

The operator helm chart can be simply cloned and deployed from this directory https://github.com/amitde69/anomaly-operator/tree/main/helm.

The IAM permissions can either be setup via [IAM roles for ServiceAccount (IRSA)](https://docs.aws.amazon.com/emr/latest/EMR-on-EKS-DevelopmentGuide/setting-up-enable-IAM.html) or can be attached directly to the worker node IAM roles. If you are using kops or vanilla k8s, polices must be manually attached to node instances.
