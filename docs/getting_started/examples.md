# Detector configuration options and examples
This document covers configuration of a `Detector` resource.

When configuring a `Detector` resource there are a few required fields and optional fields

### Minimal Spec
An example minimal spec with only required fields would be as follows:

```yaml
apiVersion: monitoring.amitdebachar/v1alpha1
kind: Detector
metadata:
  name: detector-name
spec:
  
  ## (required) The detector image
  image: "amitde7896/anomaly-operator:latest-detector"
  
  ## (required) Promethus HTTP API endpoint 
  prom_url: "http://prometheus.monitoring.svc.cluster.local"
  
  ## (required) Evaluation interval in minutes in which the Detector
  ## will query Prometheus and analyze for anomalies
  ## Note: use float for under 1 min interval e.g. "0.01"  
  interval_mins: "15"

  ## (required) List of PromQL expressions which the detector 
  ## will query and evaluate using the configured Prometheus endpoint and interval   
  queries:

    ## (required) Query name which will appear in the logs and metrics 
    ## in order to correlate the detected anomaly to the configured query  
  - name: "sum_pods_running_anomaly"

    ## (required) Query PromQL expression
    query: 'sum(kube_pod_status_phase{phase=~"Running", pod=~"application-pod-.*"}) > 1'
    
    ## (required) parse_datetime formated text such as - "1m" "4h" "2d" "3w"
    ## Past period for which the detector will train and learn the trend
    ## Note: bigger value means longer query time
    train_window: "14d"
```
### Custom Anomaly Spec
You can tune some parameters in order to avoid misclassifying an anomaly.

An example custom anomaly spec with optional fields would be as follows:

```yaml
apiVersion: monitoring.amitdebachar/v1alpha1
kind: Detector
metadata:
  name: detector-name
spec:
  ...
  queries:
  - name: "sum_pods_running_anomaly"
    query: 'sum(kube_pod_status_phase{phase=~"Running", pod=~"application-pod-.*"}) > 1'
    train_window: "14d"

    ## (optional) parse_datetime formated text such as - "1m" "4h" "2d" "3w"
    ## Past period for which the detector will evaluate anomalies based on "train_window" period
    ## Default: "1h" 
    detection_window_hours: "2h"

    ## (optional) float value which tunes the Prophet parameter "changepoint_prior_scale"
    ## https://facebook.github.io/prophet/docs/trend_changepoints.html#adjusting-trend-flexibility
    ## This will affect how flexible or stiff the model is between change points (date points)
    ## Note: Lower is stiffer Higher is more flexible  
    ## Default: "0.05" 
    flexibility: "10"

    ## (optional) integer value which is used for PromQL request in the "step" paramter
    ## This will affect how many data points the detector is getting from Prometheus for a given timeframe
    ## Note: Higher means longer query time and more sensitive to anomalies
    ## Default: number of hours in "train_window"
    ## Disclaimer: setting too small value for a longer "train_window" can cause PromQL to fail
    resolution: 1400

    ## (optional) integer value represent precetange 
    ## which is used to increase/decrease the buffer the detector will append to the detection threshold 
    ## Note: Higher means less likely to find anomalies
    ## Default: 100
    buffer_pct: 150
```

### Custom Pod Spec
You can override the pod template and spec used for the detector pod.

!!!warning
    The override should be mindful of the default pod spec that is created
    by the operator.

    It is reccommended to at least use the same default spec and add more options
    and not change the existing ones such as `env` section and `volumeMounts`/`volumes`.

    Change at your own risk. 

!!!warning
    Both `image` and the `pod_spec.spec.containers.*.image` are required
    in this case but only the `containers` section will determine the image used in the created deployment

An example custom pod spec would be as follows:

```yaml
apiVersion: monitoring.amitdebachar/v1alpha1
kind: Detector
metadata:
  name: detector-name
spec:
  image: "..."
  ...
  pod_spec:
    spec:
      containers:
        - env:
            - name: LOG_LEVEL
              value: INFO
          image: amitde7896/anomaly-operator:latest-detector
          imagePullPolicy: Always
          name: custom-detector-spec
          ports:
            - containerPort: 9090
              name: http
              protocol: TCP
          resources:
            limits:
              cpu: 1500m
              memory: 1G
            requests:
              cpu: 1500m
              memory: 1G
          volumeMounts:
            - mountPath: /app/config.yaml
              name: detector-name
              subPath: detector-name-conf.yaml
      serviceAccount: detector-name
      serviceAccountName: detector-name
      volumes:
        - configMap:
            defaultMode: 420
            name: detector-name
          name: detector-name
  ...
  queries:
  ...
```