---
title: "Recovering from Drift"
description: "Recovering from Drift"
lead: "How ACK controllers detect and remediate resource drift"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 55
toc: true
---

Kubernetes controllers work on the principal of [constant
reconciliation][constant-reconciliation]. In essence, they continuously look at
the current desired state of the system and compare it to the actual state,
using the difference to determine the action required to get to the desired end
result.

Once a controller has reconciled a resource to it's desired state, the
controller shouldn't need to continue reconciling - the actual state of the
resource meets the specification. However, this is only true for closed systems,
where the controller is the only actor interacting with a resource.
Unfortunately, ACK controllers don't act in a closed system. ACK controllers are
not the only actor capable of modifying the actual state of any AWS resources -
other programs, or even people, may have their own privileges. When another
actor modifies a resource after the ACK controller has reconciled it to its
desired state, that's called "Drift".

ACK controllers detect drift by continuing to reconcile resources after they
have reached their desired state, but with much longer delays between
reconciliation attempts. By default, all ACK controllers attempt to detect drift
once every **10 hours**. That is, every 10 hours after a resource has been
marked with the `ResourceSynced = true` condition, its owner controller will
describe the resource in AWS to see if it no longer matches the desired state.
If the controller detects a difference, it then starts the reconciliation loop
again to get back to that state (just as when any other change has been made).

{{% hint type="info" title="Existing resource overrides" %}}
Some resources require more frequent drift remediation. For example, if a
resource runs a stateful workload whose status changes frequently (such as a
SageMaker `TrainingJob`). For these resources, the drift remediation period may
already have been decreased by the controller authors to improve the
responsiveness of the resource's `Status`.

All override periods are logged to stdout when the controller is started.
{{% /hint %}}

## Overriding the drift remediation period

### For all resources owned by a controller

If you would like to decrease the drift remediation period for *all* resources
owned by a controller, update the `reconcile.defaultResyncPeriod` value in the
`values.yaml` file with the number of seconds for the new period, like so:

```yaml
reconcile:
    defaultResyncPeriod: 1800 # 30 minutes (in seconds)
```

### For a single resource type

The most granular configuration for setting reconciliation periods is to apply
it to all resources of a given type. For example, all S3 `Bucket` managed by a
single controller. 

Add the resource name and the overriding period (in seconds) to the 
`reconcile.resourceResyncPeriods` value in the Helm chart `values.yaml` like
so: 

```yaml
reconcile:
    resourceResyncPeriods:
        Bucket: 1800 # 30 minutes (in seconds)
```

[constant-reconciliation]: https://book.kubebuilder.io/cronjob-tutorial/controller-overview.html#whats-in-a-controller