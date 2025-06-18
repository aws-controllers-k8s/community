---
title: "Creating a Certificate Authority (CA) Hierarchy with the AWS Private CA ACK Controller"
lead: "Use the AWS Private CA ACK Controller to create a CA Hierarchy with a Root and Subordinate CA"
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 40
toc: true
---

The CA hierarchy will consist of a root CA that signs the certificate of a subordinate CA with both CAs hosted in AWS Private CA.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you have access to an Amazon EKS cluster.

For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html).

## Prerequisites

This guide assumes that you have:
- Created an EKS cluster with Kubernetes version 1.16 or higher
- Installed the following tools on the client machine used to access your Kubernetes cluster
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters. You can configure kubectl to point to the EKS cluster created during setup via [aws eks update-kubeconfig](https://docs.aws.amazon.com/cli/latest/reference/eks/update-kubeconfig.html)
  - [Helm 3.7+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
- AWS IAM permissions to create roles and permissions to attach policies to roles (`iam:CreateRole` and `iam:AttachRolePolicy`). These permissions should be available to the AWS CLI that you have installed on your machine as you work through this tutorial.

## Setting up Hierarchy

### 1. Set a region parameter for where you want your resources to be deployed

Run the following command updated with the AWS region that you would like to deploy to:
```
export REGION=us-east-1
```

### 2. Install the latest version of the AWS Private Certificate Authority ACK controller into the EKS cluster

```
export RELEASE_VERSION=$(curl -sL https://api.github.com/repos/aws-controllers-k8s/acmpca-controller/releases/latest | jq -r '.tag_name | ltrimstr("v")')

aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws

helm install --create-namespace -n ack-system ack-acmpca-controller \ oci://public.ecr.aws/aws-controllers-k8s/acmpca-chart --version=$RELEASE_VERSION --set=aws.region=$REGION
```

You can verify the installation succeeded by doing the following:

```
kubectl --namespace ack-system get pods -l "app.kubernetes.io/instance=ack-acmpca-controller"
```

The output from the above command should look like this. The STATUS of Running shows us that the pod has come up successfully:

```
NAME                                                                           READY STATUS  RESTARTS AGE
ack-acmpca-controller-acmpca-chart-8664d4979b-fcm5h                            1/1   Running 0        20h
```

### 3. Give the ACK controller the required IAM permissions to call AWS Private CA

The controller requires permissions to invoke AWS Private CA APIs. Once the service controller is deployed, configure the [IAM permissions via Instance Role for Service Accounts (IRSA)](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) using the value `SERVICE=acmpca` throughout. The recommended IAM Managed Policy can be found in [recommended-policy-arn](https://github.com/aws-controllers-k8s/acmpca-controller/blob/main/config/iam/recommended-policy-arn).

Alternatively, you can also use EKS Pod Identities to give the service account permissions to call AWS Private CA. Following the setup above, the ACK controller will be running in the EKS cluster with a service account named `aws-acmpca-controller`. This is the service account you will give permissions to. You can learn more about EKS Pod Identities [here](https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html). You can use the following [IAM managed policy](https://github.com/aws-controllers-k8s/acmpca-controller/blob/main/config/iam/recommended-policy-arn) when setting up permissions for the service account.

### 4. Create a Certificate Authority Hierarchy

You will be setting up a CA hierarchy in AWS Private Certificate Authority. The CA hierarchy will consist of a root CA and a subordinate CA that is signed by the root.

Save the following contents in a file named `certificate_hierarchy.yaml`:

```
apiVersion: acmpca.services.k8s.aws/v1alpha1
kind: CertificateAuthority
metadata:
  name: root-ca
spec:
  type: ROOT
  certificateAuthorityConfiguration:
    keyAlgorithm: RSA_2048
    signingAlgorithm: SHA256WITHRSA
    subject:
      commonName: root
      organization: string
      organizationalUnit: string
      country: US
      state: VA
      locality: Arlington
---
apiVersion: acmpca.services.k8s.aws/v1alpha1
kind: Certificate
metadata:
  name: root-ca-certificate
spec:
  certificateOutput:
    namespace: default
    name: root-ca-certificate-secret
    key: certificate
  certificateAuthorityRef:
    from:
      name: root-ca
  certificateSigningRequestRef:
    from:
      name: root-ca
  signingAlgorithm: SHA256WITHRSA
  templateARN: arn:aws:acm-pca:::template/RootCACertificate/V1
  validity:
    type: DAYS
    value: 100
---
apiVersion: acmpca.services.k8s.aws/v1alpha1
kind: CertificateAuthorityActivation
metadata:
  name: root-ca-activation
spec:
  completeCertificateChainOutput:
    namespace: default
    name: root-ca-certificate-secret
    key: certificateChain
  certificateAuthorityRef:
    from:
      name: root-ca
  certificate:
    namespace: default
    name: root-ca-certificate-secret
    key: certificate
  status: ACTIVE

---
apiVersion: acmpca.services.k8s.aws/v1alpha1
kind: CertificateAuthority
metadata:
  name: sub-ca
spec:
  type: SUBORDINATE
  certificateAuthorityConfiguration:
    keyAlgorithm: RSA_2048
    signingAlgorithm: SHA256WITHRSA
    subject:
      commonName: sub
      organization: string
      organizationalUnit: string
      country: US
      state: VA
      locality: Arlington
---
apiVersion: acmpca.services.k8s.aws/v1alpha1
kind: Certificate
metadata:
  name: sub-ca-certificate
spec:
  certificateOutput:
    namespace: default
    name: sub-ca-certificate-secret
    key: certificate
  certificateAuthorityRef:
    from:
      name: root-ca
  certificateSigningRequestRef:
    from:
      name: sub-ca
  signingAlgorithm: SHA256WITHRSA
  templateARN: arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen3/V1
  validity:
    type: DAYS
    value: 90
---
apiVersion: acmpca.services.k8s.aws/v1alpha1
kind: CertificateAuthorityActivation
metadata:
  name: sub-ca-activation
spec:
  completeCertificateChainOutput:
    namespace: default
    name: sub-ca-certificate-chain-secret
    key: certificateChain
  certificateAuthorityRef:
    from:
      name: sub-ca
  certificate:
    namespace: default
    name: sub-ca-certificate-secret
    key: certificate
  certificateChain:
    namespace: default
    name: root-ca-certificate-secret
    key: certificateChain
  status: ACTIVE
---
apiVersion: v1
kind: Secret
metadata:
  name: root-ca-certificate-secret
  namespace: default
data:
  certificate: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: root-ca-certificate-chain-secret
  namespace: default
data:
  certificateChain: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: sub-ca-certificate-secret
  namespace: default
data:
  certificate: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: sub-ca-certificate-chain-secret
  namespace: default
data:
  certificateChain: ""
```

Explanation of the resources listed in the file above:
  - `CertificateAuthority`: The entity that will sign your certificates with a private key stored in the AWS Private CA service. Initially this will be created in a PENDING_CERTIFICATE state waiting to have its Certificate Signing Request (CSR) signed by a trusted Certificate Authority.
    - For each CertificateAuthority resource, there will be 2 secrets created. One secret will contain the CA’s certificate chain and the other will contain the CA’s certificate. The secrets will be suffixed with `-certificate-chain-secret-` and `- certificate-secret` respectively.
  - `CertificateAuthorityActivation`: This resource is responsible for taking an AWS Private CA that is in the PENDING_CERTIFICATE state into an ACTIVE state by having you pass in a Certificate Authority as well as the signed CA CSR and the chain of trust for that CA.
  - `Certificate`: A certificate issued by the referenced CA with the given CSR.

If you want to have a closer look at the fields that can be passed into the resources, you can find that [here](https://aws-controllers-k8s.github.io/community/reference/) under the `PCA` header.

Run the following command to create the hierarchy:

```
kubectl apply -f certificate_hierarchy.yaml
```

### 5. Verify that you have successfully activated your root and subordinate CA

Running `kubectl describe certificateAuthority/root-ca` should have an output like (the output is abbreviated):

```
Name:         root-ca
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  acmpca.services.k8s.aws/v1alpha1
Kind:         CertificateAuthority
Metadata:
  Creation Timestamp:  2024-10-28T19:13:20Z
  Finalizers:
    finalizers.acmpca.services.k8s.aws/CertificateAuthority
  Generation:        1
  Resource Version:  2958
  UID:               928f0194-ba74-4db1-8e6e-dafc85ce9bf3
……
Conditions:
    Status:              True
    Type:                ACK.ResourceSynced
  Created At:            2024-10-28T19:16:05Z
  Last State Change At:  2024-10-28T19:30:47Z
  Not After:             2025-02-05T19:24:53Z
  Not Before:            2024-10-28T18:24:53Z
  Owner Account:         159621700825
  Serial:                69366911318143413430302484583889344272
  Status:                ACTIVE
Events:                  <none>
```

Here we see a `Status: ACTIVE` which indicates the successful creation of the Root CA.

Running `kubectl describe certificateAuthority/sub-ca` should have an output like (the output is abbreviated):

```
Name:         sub-ca
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  acmpca.services.k8s.aws/v1alpha1
Kind:         CertificateAuthority
Metadata:
  Creation Timestamp:  2024-10-28T19:13:20Z
  Finalizers:
    finalizers.acmpca.services.k8s.aws/CertificateAuthority
  Generation:        1
  Resource Version:  3319
  UID:               d1d76401-4869-4ccb-9c96-e890ff24a08a
……
  Conditions:
    Status:              True
    Type:                ACK.ResourceSynced
  Created At:            2024-10-28T19:14:43Z
  Last State Change At:  2024-10-28T19:35:18Z
  Not After:             2025-01-26T19:30:53Z
  Not Before:            2024-10-28T18:30:53Z
  Owner Account:         159621700825
  Serial:                336277119079656621834457892948306189246
  Status:                ACTIVE
Events:                  <none>
```

Here we see a `Status: ACTIVE` which indicates the successful creation of the subordinate CA

> _Note:_ It can take a few minutes for the CAs to get into an `ACTIVE` state. You may see intermittent error messages while the ACK controller reconciles the resource state.

If the CA Status ends up in a `FAILED` state, there should be messages in the status that should explain why. Remedy this issue and retry creating the CA hierarchy. First you should run `kubectl delete -f certificate_hierarchy.yaml` to clean up what you have done thus far. Afterwards return to step 4 and reattempt creating the CA hierarchy.

## Importing an Existing CA into ACK

If you already have an existing activated CA that you want to now manage via the ACK controller, you can do the following.

### 1. Get the CA ARN of the CA that you want to now manage via ACK

### 2. Create a template called `import-ca.yaml` and paste in the following but replacing the arn field with the ARN of your existing Private CA.

```
apiVersion: services.k8s.aws/v1alpha1
kind: AdoptedResource
metadata:
  name: adopted-ca
spec:
  aws:
    arn: arn:aws:acm-pca:us-east-1:111111111111:certificate-authority/9f0d73ed-be9c-49bb-a8bf-009bee09fdc8
  kubernetes:
    group: acmpca.services.k8s.aws
    kind: CertificateAuthority
    metadata:
      name: adopted-ca
      namespace: default
```

And then run:

```
kubectl apply -f import-ca.yaml
```

This should produce an output like:

```
adoptedresource.services.k8s.aws/adopted-ca created
```

### 3. Verify that the CA was imported successfully

Run the following command:

```
kubectl describe adoptedresource.services.k8s.aws/adopted-ca
```

The output should look like:

```
Name:         adopted-ca
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  services.k8s.aws/v1alpha1
Kind:         AdoptedResource
Metadata:
  Creation Timestamp:  2024-10-30T18:58:56Z
  Finalizers:
    finalizers.services.k8s.aws/AdoptedResource
  Generation:        1
  Resource Version:  330474
  UID:               ffbc6c5c-f4c4-4289-98c8-5bb45b4522c9
Spec:
  Aws:
    Arn:  arn:aws:acm-pca:us-east-1:111111111111:certificate-authority/9f0d73ed-be9c-49bb-a8bf-009bee09fdc8
  Kubernetes:
    Group:  acmpca.services.k8s.aws
    Kind:   CertificateAuthority
    Metadata:
      Name:       adopted-ca
      Namespace:  default
Status:
  Conditions:
    Status:  True
    Type:    ACK.Adopted
Events:      <none>
```

The `Status: True` help confirm that the import happened succesfully.

More details on the `AdoptedResource` type in ACK can be found [here](https://aws-controllers-k8s.github.io/community/docs/user-docs/adopted-resource/).

## Clean-up

You can now clean up the Private CAs that were created in the hierarchy by running the following command:

```
kubectl delete -f certificate_hierarchy.yaml
```

You can visit the AWS Private CA console to verify that CAs created here were successfully set to deleted. The controller will set the CAs to be deleted within 30 days. To prevent unwanted charges, you can set the restoration period of the CA as low as 7 days. See [here](https://docs.aws.amazon.com/privateca/latest/userguide/PCADeleteCA.html) for more information.

To remove the AWS Private CA ACK service controller, related CRDs, and namespaces see [ACK Cleanup](https://aws-controllers-k8s.github.io/community/docs/user-docs/cleanup/).

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster](https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html).
