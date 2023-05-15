---
title: "Create a managed Amazon MemoryDB for Redis Cluster using the ACK MemoryDB Controller"
description: "Create a managed Amazon MemoryDB for Redis Cluster using the memorydb-controller"
lead: "Create and use a managed Amazon MemoryDB for Redis Cluster using ACK MemoryDB Controller directly from Kubernetes"
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 44
toc: true
---

The ACK service controller for Amazon MemoryDB for Redis lets you manage Amazon MemoryDB Cluster directly from Kubernetes.
This guide will show you how to create a [Amazon MemoryDB for Redis](https://aws.amazon.com/memorydb/) Cluster using Kubernetes resource manifest.

In this tutorial we will install ACK service controller for Amazon MemoryDB for Redis on an Amazon EKS Cluster. We configure IAM permissions for the controller to invoke Amazon MemoryDB API. We create Amazon MemoryDB Cluster instances. We also deploy a sample POD on the Amazon EKS Cluster to connect to the Amazon MemoryDB Cluster instance from the POD.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see [Amazon EKS Setup](https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html). For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with Amazon EC2 Linux managed nodes. If you follow this document, install AWS CLI first. Use `aws configure` to access IAM permissions before creating EKS cluster.

### Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.18 or higher.
- Setup the [Amazon VPC Container Network Interface (CNI) plugin for Kubernetes](https://docs.aws.amazon.com/eks/latest/userguide/managing-vpc-cni.html) for the EKS Cluster.
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.

### Install the ACK service controller for Amazon MemoryDB

You can deploy the ACK service controller for Amazon MemoryDB using the [memorydb-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/memorydb-chart).

Log into the Helm registry that stores the ACK charts:
```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

You can install the Helm chart to deploy the ACK service controller for Amazon MemoryDB to your EKS cluster. At a minimum, you need to specify the AWS Region to execute the Amazon MemoryDB API calls.

For example, to specify that the Amazon MemoryDB API calls go to the `us-east-1` region, you can deploy the service controller with the following command:

```bash
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/memorydb-chart --version=1.0.0 --generate-name --set=aws.region=us-east-1
```
You can find the latest version of ACK MemoryDB controller on GitHub [release page](https://github.com/aws-controllers-k8s/memorydb-controller/releases).
Replace value for `--version` to the desired version.

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/memorydb-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed, you will need to [configure the IAM permissions][irsa-permissions] for the controller to query the Amazon MemoryDB API. For full details, please review the AWS Controllers for Kubernetes documentation for [how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the value of `memorydb` for `SERVICE`. Install wget, [oc](https://docs.openshift.com/container-platform/4.8/cli_reference/openshift_cli/getting-started-cli.html#installing-openshift-cli), [jq](https://github.com/stedolan/jq/wiki/Installation), sed first. Skip Next Steps in the documentation.

## Create Amazon MemoryDB Cluster Instances

You can create Amazon MemoryDB Clusters using the `Cluster` custom resource. The examples below show how to deploy it from your Kubernetes environment. For a full list of options available in the `Cluster` custom resource definition, you can use `kubectl explain cluster` command.

### Amazon MemoryDB Cluster

To create a Amazon MemoryDB Cluster, create a `Cluster` custom resource. The examples below shows how to provision a Amazon MemoryDB Cluster :
* The first example creates a Amazon MemoryDB Cluster in default VPC
* The second example creates a Amazon MemoryDB Cluster in specific VPC subnets and security groups.
You may choose any option from these examples. You can check more [yaml examples](https://github.com/aws-controllers-k8s/examples/tree/main/resources/memorydb/v1alpha1) of all MemoryDB resources.

#### Create Amazon MemoryDB Cluster in default VPC
The following YAML creates a MemoryDB Cluster using the default VPC subnets and security group.
```bash
MEMORYDB_CLUSTER_NAME="example-memorydb-cluster"

cat <<EOF > memorydb-cluster.yaml
apiVersion: memorydb.services.k8s.aws/v1alpha1
kind: Cluster
metadata:
  name: "${MEMORYDB_CLUSTER_NAME}"
spec:
  name: "${MEMORYDB_CLUSTER_NAME}"
  nodeType: db.t4g.small
  aclName: open-access
EOF

kubectl apply -f memorydb-cluster.yaml
```

#### Create Amazon MemoryDB Cluster in specific VPC subnets and security groups
To create a Amazon MemoryDB Cluster using specific subnets from a VPC, create a MemoryDB `SubnectGroup` custom resource first and then specify it, and a security group name, in the `Cluster` specification.

##### Create Amazon MemoryDB subnet group
The following example uses the VPC ID of the EKS Cluster. You may specify any other VPC ID by updating the `VPC_ID` variable in the following example.
Replace `EKS_CLUSTER_NAME` to the eks cluster name you created under 'Prerequisites' section.
```bash
EKS_CLUSTER_NAME="example-eks-cluster"
AWS_REGION="us-east-1"
VPC_ID=$(aws --region $AWS_REGION eks describe-cluster --name $EKS_CLUSTER_NAME --query cluster.resourcesVpcConfig.vpcId)
SUBNET_IDS=$(aws --region $AWS_REGION ec2 describe-subnets \
  --filters "Name=vpc-id,Values=${VPC_ID}" \
  --query 'Subnets[*].SubnetId' \
  --output text
)

MEMORYDB_SUBNETGROUP_NAME="example-subnet-group"

cat <<EOF > memorydb-subnetgroup.yaml
apiVersion: memorydb.services.k8s.aws/v1alpha1
kind: SubnetGroup
metadata:
  name: "${MEMORYDB_SUBNETGROUP_NAME}"
spec:
  name: "${MEMORYDB_SUBNETGROUP_NAME}"
  description: "MemoryDB cluster subnet group"
  subnetIDs:
$(printf "    - %s\n" ${SUBNET_IDS})

EOF

kubectl apply -f memorydb-subnetgroup.yaml
kubectl describe subnetgroup "${MEMORYDB_SUBNETGROUP_NAME}"
```

If you observe that the `ACK.Terminal` condition is set for the SubnetGroup and the error is similar to the following:
```bash
Status:
  Conditions:
    Message:               SubnetNotAllowedFault: Subnets: [subnet-1d111111, subnet-27d22222] are not in a supported availability zone. Supported availability zones are [us-east-1c, us-east-1d, us-east-1b].
    Status:                True
    Type:                  ACK.Terminal
```
Then update the `subnetIDs` in the input YAML and provide the subnet Ids that are in a supported availability zone.

##### Create Amazon MemoryDB Cluster
The following example uses the MemoryDB subnet group created above. It uses the provisioning EKS Cluster's security group. You may specify any other VPC security group by modifying the list of `securityGroupIDs` in the specification.
It uses the `db.t4g.small` node type for the MemoryDB cluster. Please review the [MemoryDB node types](https://docs.aws.amazon.com/memorydb/latest/devguide/nodes.supportedtypes.html) to select the most appropriate one for your workload.

```shell
EKS_CLUSTER_NAME="example-eks-cluster"
MEMORYDB_CLUSTER_NAME="example-memorydb-cluster"
AWS_REGION="us-east-1"
SECURITY_GROUP_ID=$(aws --region $AWS_REGION eks describe-cluster --name $EKS_CLUSTER_NAME --query cluster.resourcesVpcConfig.clusterSecurityGroupId)

cat <<EOF > memorydb-cluster.yaml
apiVersion: memorydb.services.k8s.aws/v1alpha1
kind: Cluster
metadata:
  name: "${MEMORYDB_CLUSTER_NAME}"
spec:
  name: "${MEMORYDB_CLUSTER_NAME}"
  nodeType: db.t4g.small
  aclName: open-access
  securityGroupIDs:
    - ${SECURITY_GROUP_ID}
  subnetGroupName: ${MEMORYDB_SUBNETGROUP_NAME}
EOF

kubectl apply -f memorydb-cluster.yaml
kubectl describe cluster "${MEMORYDB_CLUSTER_NAME}"
```

You can track the status of the provisioned database using `kubectl describe` on the `Cluster` custom resource:

```bash
kubectl describe cluster "${MEMORYDB_CLUSTER_NAME}"
```

The output of the *Cluster* resource looks like:
```bash
Name:         clusters
Namespace:    default
Labels:       <none>
Annotations:  memorydb.services.k8s.aws/last-requested-node-type: db.t4g.medium
              memorydb.services.k8s.aws/last-requested-num-shards: 2
API Version:  memorydb.services.k8s.aws/v1alpha1
Kind:         Cluster
Metadata:
  Creation Timestamp:  2022-03-30T08:47:07Z
  Finalizers:
    finalizers.memorydb.services.k8s.aws/Cluster
  Generation:        5
  Resource Version:  158132376
  Self Link:         /apis/memorydb.services.k8s.aws/v1alpha1/namespaces/default/clusters/clusters
  UID:               2f6fc7ed-fe04-42cb-85bf-b7982dedec1c
Spec:
  Acl Name:                    open-access
  Auto Minor Version Upgrade:  true
  Engine Version:              6.2
  Maintenance Window:          sat:03:00-sat:04:00
  Name:                        clusters
  Node Type:                   db.t4g.medium
  Num Replicas Per Shard:      1
  Num Shards:                  2
  Parameter Group Name:        default.memorydb-redis6
  Snapshot Retention Limit:    0
  Snapshot Window:             05:30-06:30
  Subnet Group Name:           default
  Tls Enabled:                 true
Status:
  Ack Resource Metadata:
    Arn:               arn:aws:memorydb:us-east-1:************:cluster/clusters
    Owner Account ID:  ************
  Allowed Scale Down Node Types:
    db.t4g.small
  Allowed Scale Up Node Types:
    db.r6g.12xlarge
    db.r6g.16xlarge
    db.r6g.2xlarge
    db.r6g.4xlarge
    db.r6g.8xlarge
    db.r6g.large
    db.r6g.xlarge
  Cluster Endpoint:
    Address:  clustercfg.clusters.******.memorydb.us-east-1.amazonaws.com
    Port:     6379
  Conditions:
    Last Transition Time:  2022-04-06T22:05:25Z
    Message:               Resource synced successfully
    Reason:
    Status:                True
    Type:                  ACK.ResourceSynced
  Engine Patch Version:    6.2.4
  Number Of Shards:        2
  Parameter Group Status:  in-sync
  Shards:
    Name:  0001
    Nodes:
      Availability Zone:  us-east-1d
      Create Time:        2022-03-30T09:04:04Z
      Endpoint:
        Address:          clusters-0001-001.clusters.******.memorydb.us-east-1.amazonaws.com
        Port:             6379
      Name:               clusters-0001-001
      Status:             available
      Availability Zone:  us-east-1b
      Create Time:        2022-03-30T09:04:04Z
      Endpoint:
        Address:      clusters-0001-002.clusters.******.memorydb.us-east-1.amazonaws.com
        Port:         6379
      Name:           clusters-0001-002
      Status:         available
    Number Of Nodes:  2
    Slots:            0-8191
    Status:           available
    Name:             0002
    Nodes:
      Availability Zone:  us-east-1c
      Create Time:        2022-03-30T09:04:04Z
      Endpoint:
        Address:          clusters-0002-001.clusters.******.memorydb.us-east-1.amazonaws.com
        Port:             6379
      Name:               clusters-0002-001
      Status:             available
      Availability Zone:  us-east-1d
      Create Time:        2022-03-30T09:04:04Z
      Endpoint:
        Address:      clusters-0002-002.clusters.******.memorydb.us-east-1.amazonaws.com
        Port:         6379
      Name:           clusters-0002-002
      Status:         available
    Number Of Nodes:  2
    Slots:            8192-16383
    Status:           available
  Status:             available
Events:               <none>
```

When the `Cluster Status` says `available`, you can connect to the database instance.

## Connect to Amazon MemoryDB Cluster

>To connect to the MemoryDB Cluster from a Pod running inside Kubernetes cluster:
> * Ensure that the [Amazon VPC Container Network Interface (CNI) plugin for Kubernetes](https://docs.aws.amazon.com/eks/latest/userguide/managing-vpc-cni.html) has been setup for the EKS Cluster.
> * Review [access patterns for accessing a MemoryDB Cluster in an Amazon VPC](https://docs.aws.amazon.com/memorydb/latest/devguide/memorydb-vpc-accessing.html) to confirm that MemoryDB cluster is configured to allow connection from the Pod.

The `Cluster` status contains the information for connecting to a Amazon MemoryDB for Redis Cluster. The host information can be found in `status.clusterEndpoint.address` and the port information can be found in `status.clusterEndpoint.port`. For example, you can get the connection information for a `Cluster` created in one of the previous examples using the following commands:

```bash
kubectl get cluster "${MEMORYDB_CLUSTER_NAME}" -o jsonpath='{.status.clusterEndpoint.address}'
kubectl get cluster "${MEMORYDB_CLUSTER_NAME}" -o jsonpath='{.status.clusterEndpoint.port}'
```

You can extract this information and make it available to your Pods using a [`FieldExport`][field-export] resource. The following example makes the MemoryDB cluster endpoint and port available as ConfigMap data:

```shell
MEMORYDB_CLUSTER_NAME="example-memorydb-cluster"
MEMORYDB_CLUSTER_CONN_CM="${MEMORYDB_CLUSTER_NAME}-conn-cm"

cat <<EOF > memorydb-field-exports.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${MEMORYDB_CLUSTER_CONN_CM}
data: {}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${MEMORYDB_CLUSTER_NAME}-host
spec:
  to:
    name: ${MEMORYDB_CLUSTER_CONN_CM}
    kind: configmap
  from:
    path: ".status.clusterEndpoint.address"
    resource:
      group: memorydb.services.k8s.aws
      kind: Cluster
      name: ${MEMORYDB_CLUSTER_NAME}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${MEMORYDB_CLUSTER_NAME}-port
spec:
  to:
    name: ${MEMORYDB_CLUSTER_CONN_CM}
    kind: configmap
  from:
    path: ".status.clusterEndpoint.port"
    resource:
      group: memorydb.services.k8s.aws
      kind: Cluster
      name: ${MEMORYDB_CLUSTER_NAME}
EOF

kubectl apply -f memorydb-field-exports.yaml
```
Confirm that the Amazon MemoryDB endpoint details are available in the config map by running the following command.
```shell
kubectl get configmap/${MEMORYDB_CLUSTER_CONN_CM} -o jsonpath='{.data}'
```

These values can be injected into a container either as environmental variables or files. For example, here is a snippet of a deployment definition that will add the Amazon MemoryDB Cluster connection info into a Pod:
```shell
cat <<EOF > game_leaderboard.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: leaderboard-deployment
  labels:
    app: leaderboard
spec:
  replicas: 1
  selector:
    matchLabels:
      app: leaderboard
  template:
    metadata:
      labels:
        app: leaderboard
    spec:
      containers:
      - name: leaderboard
        image: public.ecr.aws/sam/build-python3.8:latest
        tty: true
        stdin: true
        env:
          - name: MEMORYDB_CLUSTER_HOST
            valueFrom:
              configMapKeyRef:
                name: ${MEMORYDB_CLUSTER_CONN_CM}
                key: "${MEMORYDB_CLUSTER_NAME}-host"
          - name: MEMORYDB_CLUSTER_PORT
            valueFrom:
              configMapKeyRef:
                name: ${MEMORYDB_CLUSTER_CONN_CM}
                key: "${MEMORYDB_CLUSTER_NAME}-port"
EOF

kubectl apply -f game_leaderboard.yaml
```

Confirm that the leaderboard application container has been deployed successfully by running the following:
```shell
kubectl get pods –selector=app=leaderboard
```

Verify that the Pod Status is `Running`.

Get a shell to the running leaderboard container.
```shell
LEADERBOARD_POD_NAME=$(kubectl get pods –selector=app=leaderboard -o jsonpath='{.items[*].metadata.name}')
kubectl exec –stdin –tty ${LEADERBOARD_POD_NAME}  -- /bin/bash
```

In the running leaderboard container shell, run the following commands and confirm that the MemoryDB cluster host and port are available as environment variables.
```shell
# Confirm that the memorydb cluster host, port are available as environment variables
echo $MEMORYDB_CLUSTER_HOST
echo $MEMORYDB_CLUSTER_PORT
```

## Next steps

You can learn more about each of the ACK service controller for Amazon MemoryDB custom resources by using `kubectl explain` on the API resources. These include:

* `cluster`
* `user`
* `acl`
* `parametergroup`
* `subnetgroup`
* `snapshot`

The ACK service controller for Amazon MemoryDB is based on the [Amazon MemoryDB API](https://docs.aws.amazon.com/memorydb/latest/APIReference/Welcome.html). To get a full understanding of how all the APIs work, please review the [Amazon MemoryDB API documentation](https://docs.aws.amazon.com/memorydb/latest/APIReference/Welcome.html).

You can learn more about [how to use Amazon MemoryDB](https://docs.aws.amazon.com/memorydb/index.html) through the [documentation](https://docs.aws.amazon.com/memorydb/index.html).

### Cleanup

You can deprovision your Amazon MemoryDB for Redis Cluster using `kubectl delete` command.

Following commands delete the resources that were created in this tutorial.
```bash
kubectl delete -f game_leaderboard.yaml
kubectl delete -f memorydb-field-exports.yaml
kubectl delete -f memorydb-cluster.yaml
kubectl delete -f memorydb-subnetgroup.yaml
```

To remove the MemoryDB ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html
[field-export]: ../../user-docs/field-export
