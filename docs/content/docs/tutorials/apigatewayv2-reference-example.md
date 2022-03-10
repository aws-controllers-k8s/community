---
title: "Manage HTTP APIs with the ACK APIGatewayv2 Controller"
description: "Create and Invoke an Amazon APIGateway HTTP API using ACK APIGatewayv2 controller deployed on Amazon Elastic Kubernetes Service (EKS)."
lead: "Create and invoke an Amazon APIGateway HTTP API using Amazon Elastic Kubernetes Service (EKS)."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 43
toc: true
---

The ACK service controller for Amazon APIGatewayv2 lets you manage HTTP APIs and VPC Links directly from Kubernetes.
This guide will show you how to create and invoke an HTTP API using a single Kubernetes resource manifest.

In this tutorial we will invoke a single public endpoint by fronting it with an [HTTP API][apis-aws-guide]. We create a
[Route][routes-aws-guide] with `GET` HTTP method and an `HTTP_PROXY` [Integration][integrations-aws-guide] forwarding
traffic to the public endpoint. We also create an auto-deployable [Stage][stages-aws-guide] which will deploy the HTTP
API and make it invokable.

To invoke many endpoints using the single HTTP API, add multiple [Routes][routes-aws-guide] and
[Integrations][integrations-aws-guide] to the same API.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you
have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see
[Amazon EKS Setup][eks-setup]. For automated cluster creation using `eksctl`, see
[Getting started with Amazon EKS - `eksctl`][eksctl-guide].

### Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.16 or higher.
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.7+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.

### Install the ACK service controller for APIGatewayv2

Deploy the ACK service controller for Amazon APIGatewayv2 using the [apigatewayv2-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/apigatewayv2-chart).
Download it to your workspace using the following command:

```bash
helm pull oci://public.ecr.aws/aws-controllers-k8s/apigatewayv2-chart --version=v0.0.17
````

Use the following command to decompress and extract the Helm chart:

```bash
tar xzvf apigatewayv2-chart-v0.0.17.tgz
```

Deploy the controller using Helm chart, specifying that APIGatewayv2 resources should be created by default in the
`us-east-1` region:

```bash
helm install apigatewayv2-chart --generate-name --set=aws.region=us-east-1
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/apigatewayv2-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed [configure the IAM permissions][irsa-permissions] for the
controller to invoke the APIGatewayv2 API. For full details, please review the AWS Controllers for Kubernetes documentation
for [how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the
value of `apigatewayv2` for `SERVICE`.

## Create HTTP API
Execute the following command to create a manifest containing all the APIGatewayv2 custom resources and submit this
manifest to EKS cluster using kubectl.

{{% hint type="info" title="Referencing Resources" %}}
Notice that the ACK custom resources reference each other using "*Ref" fields inside the manifest and the user does not
have to worry about finding APIID, IntegrationID when creating the K8s resource manifests.

Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *APIGatewayv2*
to find the supported reference fields.
{{% /hint %}}

```bash
API_NAME="ack-api"
INTEGRATION_NAME="ack-integration"
INTEGRATION_URI="https://httpbin.org/get"
ROUTE_NAME="ack-route"
ROUTE_KEY_NAME="ack-route-key"
STAGE_NAME="ack-stage"

cat <<EOF > apigwv2-httpapi.yaml
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: API
metadata:
  name: "${API_NAME}"
spec:
  name: "${API_NAME}"
  protocolType: HTTP

---

apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Integration
metadata:
  name: "${INTEGRATION_NAME}"
spec:
  apiRef:
    from:
      name: "${API_NAME}"
  integrationType: HTTP_PROXY
  integrationURI: "${INTEGRATION_URI}"
  integrationMethod: GET
  payloadFormatVersion: "1.0"

---

apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Route
metadata:
  name: "${ROUTE_NAME}"
spec:
  apiRef:
    from:
      name: "${API_NAME}"
  routeKey: "GET /${ROUTE_KEY_NAME}"
  targetRef:
    from:
      name: "${INTEGRATION_NAME}"

---

apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Stage
metadata:
  name: "${STAGE_NAME}"
spec:
  apiRef:
    from:
      name: "${API_NAME}"
  stageName: "${STAGE_NAME}"
  autoDeploy: true
  description: "auto deployed stage for ${API_NAME}"
EOF

kubectl apply -f apigwv2-httpapi.yaml
```

The manifest contains 4 APIGatewayv2 custom resources: API, Integration, Route and Stage.
When this manifest is submitted using *kubectl*, it creates corresponding 4 custom resources in the EKS cluster.

The output of above command looks like
```
api.apigatewayv2.services.k8s.aws/ack-api created
integration.apigatewayv2.services.k8s.aws/ack-integration created
route.apigatewayv2.services.k8s.aws/ack-route created
stage.apigatewayv2.services.k8s.aws/ack-stage created
```

## Describe Custom Resources
View these custom resources using following commands:
```bash
kubectl describe api/"${API_NAME}"
kubectl describe integration/"${INTEGRATION_NAME}"
kubectl describe route/"${ROUTE_NAME}"
kubectl describe stage/"${STAGE_NAME}"
```

Output of describing *Route* resource looks like
```bash
Name:         ack-route
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  apigatewayv2.services.k8s.aws/v1alpha1
Kind:         Route
Metadata:
  Creation Timestamp:  2022-03-08T18:13:16Z
  Finalizers:
    finalizers.apigatewayv2.services.k8s.aws/Route
  Generation:  2
  Resource Version:  116729769
  UID:               0286a10e-0389-4ea8-90ae-890946d5d280
Spec:
  API Key Required:  false
  API Ref:
    From:
      Name:            ack-api
  Authorization Type:  NONE
  Route Key:           GET /ack-route-key
  Target Ref:
    From:
      Name:  ack-integration
Status:
  Ack Resource Metadata:
    Owner Account ID:  ***********
  Conditions:
    Last Transition Time:  2022-03-08T18:13:23Z
    Status:                True
    Type:                  ACK.ReferencesResolved
    Last Transition Time:  2022-03-08T18:13:23Z
    Message:               Resource synced successfully
    Reason:
    Status:                True
    Type:                  ACK.ResourceSynced
  Route ID:                *****
Events:                    <none>
```

{{% hint type="info" title="Referencing Resources" %}}
ACK controller reads the referenced resources and determines the identifiers, like APIID, from the referenced
resources. Find the *ACK.ReferencesResolved* condition inside the *Status* of Route, Integration and Stage
resources to see the progress of reference resolution.
{{% /hint %}}

## Invoke HTTP API
Execute the following command to invoke the HTTP API
```bash
curl $(kubectl get api/"${API_NAME}" -o=jsonpath='{.status.apiEndpoint}')/"${STAGE_NAME}"/"${ROUTE_KEY_NAME}"
```

The above commands finds the invocation endpoint from the *Api* custom resource and appends the required *Stage* name,
*Route Key* to the url before invoking.

The output should look similar to
```bash
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Content-Length": "0",
    "Forwarded": "by=****;for=****;host=******.execute-api.us-west-2.amazonaws.com;proto=https",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.64.1",
    "X-Amzn-Trace-Id": "Self=****;Root=****"
  },
  "origin": "****",
  "url": "https://httpbin.org/get"
}
```

## Next steps

The ACK service controller for Amazon APIGatewayv2 is based on the [Amazon APIGatewayv2 API](https://docs.aws.amazon.com/apigatewayv2/latest/api-reference/api-reference.html).

Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *APiGatewayv2* to find
all the supported Kubernetes custom resources and fields.

{{% hint type="info" title="Note" %}}
* Currently ACK service controller for APIGatewayv2 only supports HTTP APIs.
* WebSocket API support will be added in future releases.
* Support for DomainName and APIMapping will also be added in future releases.
{{% /hint %}}

### Cleanup

Remove all the resource created in this tutorial using `kubectl delete` command.

```bash
kubectl delete -f apigwv2-httpapi.yaml
```

The output of delete command should look like

```bash
api.apigatewayv2.services.k8s.aws "ack-api" deleted
integration.apigatewayv2.services.k8s.aws "ack-integration" deleted
route.apigatewayv2.services.k8s.aws "ack-route" deleted
stage.apigatewayv2.services.k8s.aws "ack-stage" deleted
```

To remove the APIGatewayv2 ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[apis-aws-guide]: https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html
[integrations-aws-guide]: https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations.html
[routes-aws-guide]: https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-routes.html
[stages-aws-guide]: https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-stages.html
[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html
[eksctl-guide]: https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html
[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html

