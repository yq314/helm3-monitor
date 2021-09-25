# Helm3 Monitor Example

This example demonstrate how to use helm3-monitor to rollback a Helm release based on events.

## Pre-requisite

- Helm3
- Minikube

## Prepare

Install Prometheus
```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add kube-state-metrics https://kubernetes.github.io/kube-state-metrics
helm repo update
helm --namespace prometheus upgrade -i -f ./prometheus.yml --create-namespace prometheus prometheus-community/prometheus
```

Install Infrabin
```shell
helm upgrade -i infrabin --set image.tag=0.9.0 -f ./infrabin.yml ./charts/infrabin
```

Upgrade Infrabin (to create a new release)
```shell
helm upgrade -i infrabin --set image.tag=0.9.1 -f ./infrabin.yml ./charts/infrabin
```

Monitor Release
```shell
export PROM_URL=$(minikube --namespace=prometheus service prometheus-server --url)
helm --debug monitor prometheus --prometheus="$PROM_URL" infrabin 'rate(infrabin_http_requests_total{code!="200"}[5m]) / rate(infrabin_http_requests_total[5m]) > 0.01'
```

To create some non success metrics
```shell
export INFRABIN_URL=$(minikube service infrabin --url)
curl "${INFRABIN_URL}/status/500"
```