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
path. The controller binds its metrics server to `--metrics-addr`, which
defaults to `0.0.0.0:8080`, so metrics are available at
`http://<pod-ip>:8080/metrics`. The Helm chart does not set this flag; it only
declares the port on the pod (named `http`) using `deployment.containerPort`.
Leave `deployment.containerPort` at its default of `8080` so the two match.

You can confirm the endpoint is working by port-forwarding to the controller pod
and curling it. Look the pod up by label so this works regardless of the Helm
release name:

```bash
kubectl -n ack-system port-forward \
  $(kubectl -n ack-system get pod -l app.kubernetes.io/instance=ack-s3-controller -o name | head -1) 8080:8080
curl -s http://localhost:8080/metrics | grep ack_
```

To scrape it, point Prometheus at the controller pods on that port. If you use
the Prometheus Operator, a `PodMonitor` selecting the controller pods (port
`8080`, path `/metrics`) works out of the box. A `ServiceMonitor` additionally
requires the chart's metrics Service, which is off by default — install with
`--set metrics.service.create=true`. That Service exposes a port named
`metricsport` targeting the container port `http`.

## ACK-specific metrics

ACK records two counters that track the outbound AWS API calls a controller
makes while reconciling resources.

`ack_outbound_api_requests_total` counts the total number of outbound AWS API
requests made by the controller. It carries the labels:

- `service` – the AWS service the controller manages (for example `s3`, `rds`).
- `op_type` – the type of operation, such as `CREATE`, `READ_ONE`, `UPDATE`, or `DELETE`.
- `op_id` – the specific AWS API call, such as `CreateBucket`.

`ack_outbound_api_requests_error_total` counts the outbound requests that
returned an error. The controller increments it for **any** non-nil error from
the SDK call — this includes non-API errors and expected ones such as `NotFound`
during a `ReadOne` (for example right after a create, or during adoption), so a
nonzero count is normal. Its labels are:

- `service` – the AWS service the controller manages.
- `op_id` – the specific AWS API call that failed.
- `status_code` – despite the name, this is **not** an HTTP status code. It is
  the string form of smithy's `ErrorFault` classification returned by
  [`ackerr.HTTPStatusCode()`][ackerr-fault], with only these values:
    - `"-1"` – the error was not an AWS API error
    - `"0"` – the error isn't defined in the AWS API model for that operation, so its
      fault class is unknown. In practice this covers many errors, often including
      throttling and 5XX responses.
    - `"1"` – a server fault that the operation's API model defines
    - `"2"` – a client fault that the operation's API model defines (for example `NoSuchBucket`)

[ackerr-fault]: https://github.com/aws-controllers-k8s/runtime/blob/main/pkg/errors/error.go#L97-L103

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

Error ratio of outbound AWS API calls per service. Because expected errors (such
as `NotFound` on `ReadOne`) are counted too, a nonzero ratio is normal — watch it
as a trend rather than an absolute value:

```promql
sum by (service) (rate(ack_outbound_api_requests_error_total[5m]))
  /
sum by (service) (rate(ack_outbound_api_requests_total[5m]))
```

Errors per operation, broken down by fault class (see the `status_code` values above):

```promql
sum by (op_id, status_code) (rate(ack_outbound_api_requests_error_total[5m]))
```

Reconcile error rate per controller, from the controller-runtime metrics:

```promql
sum by (controller) (rate(controller_runtime_reconcile_errors_total[5m]))
```

Watching the outbound-request rate and error ratio before and after upgrading a
controller is a practical way to catch regressions early. For example, compare the
current error ratio against the same window a day earlier to spot a change
introduced by an upgrade:

```promql
sum by (service) (rate(ack_outbound_api_requests_error_total[5m]))
  / sum by (service) (rate(ack_outbound_api_requests_total[5m]))
-
sum by (service) (rate(ack_outbound_api_requests_error_total[5m] offset 1d))
  / sum by (service) (rate(ack_outbound_api_requests_total[5m] offset 1d))
```

Running a workload in a staging environment and comparing these metrics across
controller versions gives you a baseline before you upgrade in production.
