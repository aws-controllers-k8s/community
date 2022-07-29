---
title: "Managing Tags on your AWS Resources"
description: "Managing Tags on your AWS Resources"
lead: ""
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 67
toc: true
---

Most AWS resources can have one or more Tags, defined as simple key/value
pairs, associated with them. These Tags allow you to organize and categorize
your AWS resources for accounting and informational purposes. ACK custom
resources (CRs) that support Tags will have a *Spec.Tags* field that
stores user-defined key/value pairs. In addition to user-defined Tags,
ACK also supports a set of default tags, which are Tags that the ACK
controller will automatically ensure are on all resources that it manages.

The two default tags added by ACK controller are `services.k8s.aws/controller-version`
and `services.k8s.aws/namespace`. The *controller-version* tag value is the name of
corresponding AWS service and version for that controller(Ex: `s3-v0.1.3`).
And the *namespace* tag value is the Kubernetes namespace for the ACK
resource.(Ex: `default`)

When tags are already present inside the Kubernetes custom resource's `Spec.Tags`,
ACK default tags are added to the AWS resource's tags collection along with those
tags from `Spec.Tags`. Priority is given to `Spec.Tags` when there is a
conflict between ACK default tag keys and tag keys in `Spec.Tags`.

## Example

For a resource manifest like

```yaml
apiVersion: ecr.services.k8s.aws/v1alpha1
kind: Repository
metadata:
  name: my-ack-tagging-repo
  namespace: default
spec:
  name: my-ack-tagging-repo
  tags:
  - key: "first"
    value: "1"
  - key: "second"
    value: "2"
```

The sample response for `list-tags-for-resource` will look like

```bash
aws ecr list-tags-for-resource --resource-arn arn:aws:ecr:us-west-2:************:repository/my-ack-tagging-repo
{
    "tags": [
        {
            "Key": "services.k8s.aws/controller-version",
            "Value": "ecr-v0.1.4"
        },
        {
            "Key": "first",
            "Value": "1"
        },
        {
            "Key": "services.k8s.aws/namespace",
            "Value": "default"
        },
        {
            "Key": "second",
            "Value": "2"
        }
    ]
}

```

## Configuring Default Tags

The default tags added by ACK controllers are configurable during controller
installation.

* To remove the ACK default tags, set the `resourceTags` Helm value to be `{}` inside
*values.yaml* file or use `--set 'resourceTags={}'` during helm chart installation.

* To override the default ACK tags, include each tag "key=value" pair as a list under
`resourceTags` in *values.yaml* file
  ```yaml
  resourceTags:
  - tk1=tv1
  - tk2=tv2
  ```
  You can also override default ACK tags using `--set 'resourceTags=[tk1=tv1, tk2=tv2]'`
  during helm chart installation.

* ACK supports variable expansion inside tag values for following variables:
  - %CONTROLLER_SERVICE%
  - %CONTROLLER_VERSION%
  - %K8S_NAMESPACE%
  - %K8S_RESOURCE_NAME%

  A custom resource tag `k8s-name=%K8S_RESOURCE_NAME` in above ecr repository example
  would be expanded to "k8s-name=my-ack-tagging-repo"
