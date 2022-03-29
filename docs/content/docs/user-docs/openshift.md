---
title: "Red Hat OpenShift"
description: "Configuration details specific to OpenShift clusters"
lead: ""
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 99
toc: true
---

Configuration for ACK controllers in an OpenShift cluster.

## Pre-installation instructions

When ACK service controllers are installed via OperatorHub, a cluster administrator will need to perform the following pre-installation steps to provide the controller any credentials and authentication context it needs to interact with the AWS API.

Configuration and authentication in OpenShift requires the use of IAM users and policies. Authentication credentials are set inside a `ConfigMap` and a `Secret` before installation of the controller.

### Step 1: Create the installation namespace

If the default `ack-system` namespace does not exist already, create it:
```bash
oc new-project ack-system
```

### Step 2: Bind an AWS IAM principal to a service user account

Create a user with the `aws` CLI (named `ack-elasticache-service-controller` in our example):
```bash
aws iam create-user --user-name ack-elasticache-service-controller
```

Enable programmatic access for the user you just created:
```bash
aws iam create-access-key --user-name ack-elasticache-service-controller
```

You should see output with important credentials:
```json
{
    "AccessKey": {
        "UserName": "ack-elasticache-service-controller",
        "AccessKeyId": "00000000000000000000",
        "Status": "Active",
        "SecretAccessKey": "abcdefghIJKLMNOPQRSTUVWXYZabcefghijklMNO",
        "CreateDate": "2021-09-30T19:54:38+00:00"
    }
}
```

Save or note `AccessKeyId` and `SecretAccessKey` for later use.

Each service controller repository provides a recommended policy ARN for use with the controller. For an example, see the recommended policy for [Elasticache here](https://github.com/aws-controllers-k8s/elasticache-controller/blob/main/config/iam/recommended-policy-arn).

Attach the recommended policy to the user we created in the previous step:
```bash
aws iam attach-user-policy \
    --user-name ack-elasticache-service-controller \
    --policy-arn 'arn:aws:iam::aws:policy/AmazonElastiCacheFullAccess'
```

### Step 3: Create `ack-$SERVICE-user-config` and `ack-$SERVICE-user-secrets` for authentication

Enter the `ack-system` namespace. Create a file, `config.txt`, with the following variables, leaving `ACK_WATCH_NAMESPACE` blank so the controller can properly watch all namespaces, and change any other values to suit your needs:

```bash
ACK_ENABLE_DEVELOPMENT_LOGGING=true
ACK_LOG_LEVEL=debug
ACK_WATCH_NAMESPACE=
AWS_REGION=us-west-2
AWS_ENDPOINT_URL=
ACK_RESOURCE_TAGS=hellofromocp
```

Now use `config.txt` to create a `ConfigMap` in your OpenShift cluster:
```bash
export SERVICE=elasticache

oc create configmap \
--namespace ack-system \
--from-env-file=config.txt ack-$SERVICE-user-config
```

Save another file, `secrets.txt`, with the following authentication values, which you should have saved from earlier when you created your user's access keys:
```bash
AWS_ACCESS_KEY_ID=00000000000000000000
AWS_SECRET_ACCESS_KEY=abcdefghIJKLMNOPQRSTUVWXYZabcefghijklMNO
```

Use `secrets.txt` to create a `Secret` in your OpenShift cluster:
```bash
oc create secret generic \
--namespace ack-system \
--from-env-file=secrets.txt ack-$SERVICE-user-secrets
```

Delete `config.txt` and `secrets.txt`.

{{% hint type="warning" title="Warning" %}}
If you change the name of either the `ConfigMap` or the `Secret` from the values given above, i.e. `ack-$SERVICE-user-config` and `ack-$SERVICE-user-secrets`, then installations from OperatorHub will not function properly. The Deployment for the controller is preconfigured for these key values.
{{% /hint %}}

### Step 4: Install the controller

Follow the instructions for [installing the controller using OperatorHub](../install/#install-an-ack-service-controller-with-operatorhub-in-red-hat-openshift).


## Additional uninstallation steps

Perform the following cleanup steps in addition to the steps in [Uninstall an ACK Controller](../cleanup).

### Uninstall the ACK Controller

Navigate in the OpenShift dashboard to the OperatorHub page and search for the controller name. Select __Uninstall__ to remove the controller.

### Delete ConfigMap

Delete the following `ConfigMap` you created in pre-installation:
```bash
oc delete configmap ack-$SERVICE-user-config
```

### Delete user Secret

Delete the folllowing `Secret` you created in pre-installation:
```bash
oc delete secret ack-$SERVICE-user-secrets
```

## Next Steps

After you install the controller, you can follow the [Cross Account Resource Management](../cross-account-resource-management) instructions to manage resources in multiple AWS accounts.
