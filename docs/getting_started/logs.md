In order to investigate an anomaly that was detected, the detector pod will output logs to STDOUT with the relevant information.

### Example
For the following `Detector`:
```yaml
apiVersion: monitoring.amitdebachar/v1alpha1
kind: Detector
metadata:
  name: minimal-detector
spec:
  image: "amitde7896/anomaly-operator:latest-detector"
  prom_url: "http://prometheus.monitoring.svc.cluster.local"
  interval_mins: "15"
  queries:
  - name: "sum_pods_running_anomaly"
    query: 'sum(kube_pod_labels{label_app=~"example-application-.*"}) by (label_app) > 1'
    train_window: "14d"
```
The following logs will be printed when anomaly detected:
```json
{
    "timestamp": "2023-04-12T22:03:05.261Z",
    "level": "WARNING",
    "message": "Found 1 anomalies for sum_pods_running_anomaly in {'label_app': 'example-application-b'}"
}
{
    "timestamp": "2023-04-12T22:03:05.378Z",
    "level": "WARNING",
    "message": "[sum_pods_running_anomaly] {'label_app': 'example-application-b'} time: 2023-04-12 21:41:51 value: 12"
}
```

!!!note
    In cases where a single query returns many different results, the log will contain the returned values from the Prometheus response.<br/>
    For example in the above `Detector` a `by (label_app)` was added to the PromQL therefore the `label_app` value was added to the log for the detected anomaly.<br/>
