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
To use these feature, ensure you enable the [feature gates](https://github.com/aws-controllers-k8s/ec2-controller/blob/b6dff777c35d03335ebb0c3ffca5ee7577e70f18/helm/values.yaml#L164-L172) during helm install, as they are disabled by default.

### ResourceAdoption
This feature allows users to adopt AWS resources by specifying the adoption policy as an annotation `services.k8s.aws/adoption-policy`.
This annotation currently supports two values, `adopt` and `adopt-or-create`.
`adopt` adoption policy strictly adopts resources as they are in AWS, and it is highly recommended to provide an empty spec (as it will be overriden
if adoption is successful) and `services.k8s.aws/adoption-fields` annotation with all the fields necessary to retrieve the resource from AWS 
(this would be the `name` for EKS cluster, `queueURL` for SQS queues, `vpcID` for VPCs, or `arn` for SNS Topic, etc.)
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
After reconciliation, all the fields in the spec and status will be populated by the controller.

When you want the controller to create resources if they don't exist, you can set
`adopt-or-create` adoption policy. With this policy, as the name suggests, the controller
will adopt the resource if it exists, or create it if it doesn't.
For `adopt-or-create` the controller expects the spec to be populated by the user with all the 
fields necessary for a find/create. If the read operation required field is in the status
the `adoption-fields` annotation will be used to retrieve such fields.
If the adoption is successful for `adopt-or-create`, the controller will attempt updating
your AWS resource, to ensure your ACK manifest is the source of truth. 
Here are some sample manifests:
EKS Cluster
```yaml
apiVersion: eks.services.k8s.aws/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
  annotations:
    services.k8s.aws/adoption-policy: "adopt-or-create"
spec:
  name: my-cluster
  roleARN: arn:role:123456789/myrole
  version: "1.32"
  resourcesVPCConfig:
    endpointPrivateAccess: true
    endpointPublicAccess: true
    subnetIDs:
      - subnet-312ensdj2313dnsa2
      - subnet-1e323124ewqe43213

```
VPC
```yaml
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: VPC
metadata:
  name: hello
  annotations:
    services.k8s.aws/adoption-policy: adopt-or-create
    services.k8s.aws/adoption-fields: |
      {"vpcID": "vpc-123456789012"}
spec:
  cidrBlocks: 
  - "2.0.0.0/16"
  tags:
  - key: k1
    value: v1
```

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
