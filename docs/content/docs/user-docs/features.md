---
title : "Alpha Features"
description: ""
lead: ""
draft: false
menu: 
  docs:
    parent: "getting-started"
weight: 30
toc: true
---

Currently we support 4 feature gates for our controllers.

### ResourceAdoption
This feature allows users to adopt AWS resources by specifying the adoption policy as an annotation `services.k8s.aws/adoption-policy` (currently only supporting `adopt` as a value), and providing the fields required for a read operation in an annotation called `services.k8s.aws/adoption-fields` in json format, and an empty spec.
Here's an example for how to adopt an EKS cluster:

```yaml
apiVersion: eks.services.k8s.aws/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
  annotations:
    services.k8s.aws/adoption-policy: "adopt"
    services.k8s.aws/adoption-fields: | 
        {
          "name": "my-cluster"
        }
```
Applying the above manifest allows users to adopt an existing EKS cluster named `my-cluster`.
After reconciliation, all the fields in the spec and status will be filled by the controleler.
This feature is currently available for the s3 controller, and we'll see more releases in the future for other controllers

### ReadOnlyResources
This feature allows users to mark a resource as read only, as in, it would ensure that the resource will not call any update operation, and will not be patching anything in the spec, but instead, it will be reconciling the status and ensuring it matches the resource it points to.
To ensure this feature works, the resource should first exist, hence the annotation either needs to be created and later marked as read only, or it can be adopted and reconciled as a read-only resource.
Here's an example of a manifest that adopts, and manages the resource as read-only:

```yaml
apiVersion: eks.services.k8s.aws/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
  annotations:
    services.k8s.aws/read-only: "true"
    services.k8s.aws/adoption-policy: "adopt"
    services.k8s.aws/adoption-fields: | 
        {
          "name": "my-cluster"
        }
```
Applying the above manifest allows users to adopt an existing EKS cluster named `my-cluster` and manage it as a read-only resource.


### TeamLevelCARM and ServiceLevelCARM
The TeamLevelCARM feature builds on [`Manage Resources In Multiple AWS Accounts`](cross-account-resource-management). It allows teams using the same account to annotate a different namespaces with a team ID, and each team ID is associated with a specific AWS role ARN, specified in a config map named `ack-role-team-map`, allowing the controller to have different roles for different namespaces. 

Here's an example:
`ack-role-team-map` config-map
```yaml
data:
  team-a: "arn:aws:iam::111111111111:role/team-a-global-role"
  team-b: "arn:aws:iam::111111111111:role/team-b-global-role"
```

`team-a` namespace
```yaml
metadata:
  name: testing-a
  annotations:
    services.k8s.aws/team-id: "team-a"
```
`team-b` namespace
```yaml
---
metadata:
  name: testing-b
  annotations:
    services.k8s.aws/team-id: "team-b"
```

The ServiceLevelCARM feature allows users to speciy different IAM roles for different service controllers within the same team and AWS Account.

For example, you might want the s3 controller to assume a different IAM Role than the dynamodb controller, even when managing resources in the same team/aws account.

Here's an example:

`team-a` namespace
```yaml
metadata:
  annotations:
    services.k8s.aws/team-id: "team-a-global"
    services.k8s.aws/team-id: "team-a"
```

`ack-role-account-map` config-map
```yaml
data:
  team-a: "arn:aws:iam::111111111111:role/team-a-global-role"
  s3.team-a: "arn:aws:iam::111111111111:role/team-a-s3-role"
  dynamodb.team-a: "arn:aws:iam::111111111111:role/team-a-dynamodb-role"
```
