The detector pods expose Prometheus metrics at `/metrics` on port 9090 by default.
!!!note
    You will have to create your own `Service` and `ScrapeConfig` or `ServiceMonitor` to expose and scrape the metrics.
    See [example](https://github.com/amitde69/anomaly-operator/blob/main/examples/minimal_spec_servicemontor.yaml)

`anomaly_counter` - Total number of anomalies found for the last interval. 

You can create your own rules/alerts to get notified whenever an anomaly was detected based on the metrics.
