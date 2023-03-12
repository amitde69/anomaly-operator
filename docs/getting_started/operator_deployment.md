# K8S Anomaly Detection Operator Deployment
## Deploy Helm Chart

The operator helm chart can be simply cloned and deployed from this directory 
<a ref="https://github.com/amitde69/anomaly-operator/tree/main/helm">https://github.com/amitde69/anomaly-operator/tree/main/helm<a/>.


## Deploy Operator to Cluster

We recommend using the Helm chart.

=== "Via Helm"

    1. Clone the Operator repo
    ```
    git clone https://github.com/amitde69/anomaly-operator
    ```
    2. Install the chart via `helm install`.
    ```
    helm install anomaly-operator helm/
    ```

        !!!note
            The `helm install` command automatically applies the CRDs.


    Helm install command to override image repo and tag : 
    ```
    helm install anomaly-operator helm \
        --set repo="xxx/anomaly-operator" \
        --set tag="xxx-operator"
    ```

    Helm install command to override namespace : 
    ```
    helm install anomaly-operator helm -n custom-namespace
    ```

=== "Via YAML manifests"
    1. Clone the Operator repo
    ```
    git clone https://github.com/amitde69/anomaly-operator
    ```
    2. Template The Helm Chart
    ```
    helm template helm > deploy.yaml
    ```
    3. Apply the deploy.yaml
    ```
    kubectl apply -f deploy.yaml
    ```

    Helm template command to override image repo and tag : 
    ```
    helm template helm \
        --set repo="xxx/anomaly-operator" \
        --set tag="xxx-operator" > deploy.yaml
    ```

    Helm template command to override namespace : 
    ```
    helm install helm -n custom-namespace > deploy.yaml
    ```

## Upgrade The Operator

The operator doesn't receive security updates automatically. You need to manually upgrade to a newer version when it becomes available.

