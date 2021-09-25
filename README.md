# Helm3 Monitor Plugin

Monitor a Helm3 release, rollback to a previous version depending on the result of a PromQL (Prometheus) query result

## Install

```shell
> helm plugin install https://github.com/yq314/helm3-monitor
```

## Usage

A rollback happen only if the number of result from the query is greater than 0.

You can find a step-by-step example in the ./examples directory.

```shell
> helm monitor prometheus --prometheus=http://prometheus:9090 \
    <release_name> \
    'rate(http_requests_total{statusClass!="2XX"}[5m]) / rate(http_requests_total[5m]) > 0.01'
```
