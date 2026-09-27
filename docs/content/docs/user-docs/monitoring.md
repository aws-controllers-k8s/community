---
title: "Monitoring ACK Controllers with Prometheus Metrics"
description: "Scrape and interpret the Prometheus metrics exposed by ACK controllers"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 80
toc: true
---

Every ACK service controller exposes Prometheus metrics that let you monitor
the health of the controller and its interactions with the AWS APIs. These come
from two sources: metrics that ACK records itself (prefixed with `ack_`), and
the standard metrics that [controller-runtime][controller-runtime-metrics]
registers for every controller.

[controller-runtime-metrics]: https://book.kubebuilder.io/reference/metrics-reference

## Metrics endpoint

Each controller serves its metrics in Prometheus text format at the `/metrics`
path. The address is controlled by the `--metrics-addr` flag and defaults to
`0.0.0.0:8080`, so metrics are available at `http://<pod-ip>:8080/metrics`.

You can confirm the endpoint is working by port-forwarding to the controller pod
and curling it:

```bash
kubectl -n ack-system port-forward deployment/ack-s3-controller 8080:8080
curl -s http://localhost:8080/metrics | grep ack_
```

To scrape it, point Prometheus at the controller pod on that port. If you use
the Prometheus Operator, a `PodMonitor` or `ServiceMonitor` selecting the
controller pods (port `8080`, path `/metrics`) is the usual approach.

## ACK-specific metrics

ACK records two counters that track the outbound AWS API calls a controller
makes while reconciling resources.

`ack_outbound_api_requests_total` counts the total number of outbound AWS API
requests made by the controller. It carries the labels:

- `service` – the AWS service the controller manages (for example `s3`, `rds`).
- `op_type` – the type of operation, such as `CREATE`, `READ_ONE`, `UPDATE`, or `DELETE`.
- `op_id` – the specific AWS API call, such as `CreateBucket`.

`ack_outbound_api_requests_error_total` counts the outbound requests that came
back with a 4XX or 5XX HTTP status code. Its labels are:

- `service` – the AWS service the controller manages.
- `op_id` – the specific AWS API call that failed.
- `status_code` – the HTTP status code returned by the API.

Because both are counters, you typically look at their rate rather than the raw
value.

## controller-runtime metrics

ACK registers its metrics into the controller-runtime metrics registry, so each
controller also exposes the standard controller-runtime and client-go metrics on
the same endpoint. The ones most useful for monitoring controller health include
`controller_runtime_reconcile_total` (reconciles by controller and result),
`controller_runtime_reconcile_errors_total`, `controller_runtime_reconcile_time_seconds`
(reconcile duration histogram), the `workqueue_*` metrics (queue depth, latency,
and retries), and `rest_client_requests_total` (calls to the Kubernetes API
server). See the [controller-runtime metrics reference][controller-runtime-metrics]
for the full list.

## Example PromQL queries

Rate of outbound AWS API calls per service and operation over the last five
minutes:

```promql
sum by (service, op_id) (rate(ack_outbound_api_requests_total[5m]))
```

Error ratio of outbound AWS API calls per service, which is a good signal that a
controller is being throttled or is hitting permission or validation errors:

```promql
sum by (service) (rate(ack_outbound_api_requests_error_total[5m]))
  /
sum by (service) (rate(ack_outbound_api_requests_total[5m]))
```

Throttled requests (HTTP 429) per operation:

```promql
sum by (op_id) (rate(ack_outbound_api_requests_error_total{status_code="429"}[5m]))
```

Reconcile error rate per controller, from the controller-runtime metrics:

```promql
sum by (controller) (rate(controller_runtime_reconcile_errors_total[5m]))
```

Watching the outbound-request rate and error ratio before and after upgrading a
controller is a practical way to catch regressions early. Running a workload in a
staging environment and comparing these metrics across controller versions gives
you a baseline before you upgrade in production.
