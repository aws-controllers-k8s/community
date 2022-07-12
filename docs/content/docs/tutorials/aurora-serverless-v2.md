---
title: "Manage an Aurora Serverless v2 cluster with the ACK RDS Controller"
description: "Create an Aurora Serverless v2 cluster from an Amazon Elastic Kubernetes Service (EKS) deployment."
lead: "Create and manage an Aurora Serverless v2 cluster directly from Kubernetes"
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 45
toc: true
---

Aurora Serverless v2 introduces the ability to automatically and instantly scale
database capacity for Aurora MySQL-compatiable and Aurora PostgreSQL-compatible
clusters. Scaling uses fine-grained increments called Aurora capacity units
(ACUs) that incrementally scale up and down over smaller units (e.g. 0.5, 1,
1.5, 2) instead of doubling on each scaling operation (e.g. 16 => 32).
Aurora Serverless v2 helps applications with variable workloads or multitenancy
to only use the resources they need and manage costs, instead of having to
provision for a peak workload.

In this tutorial you will learn how to create and manage
[Aurora Serverless v2](https://aws.amazon.com/rds/aurora/serverless/) instances
from an Amazon Elastic Kubernetes (EKS) deployment.

## Prerequisites

This tutorial uses [Amazon EKS Workshop](https://www.eksworkshop.com/010_introduction/) to deploy EKS cluster.

### Set up Amazon EKS Workshop

1. Create a [workspace](https://www.eksworkshop.com/020_prerequisites/workspace/)
2. Install the [Kubernetes tools](https://www.eksworkshop.com/020_prerequisites/k8stools/)
3. Create an [IAM role for workspace](https://www.eksworkshop.com/020_prerequisites/iamrole/)
4. Attach [IAM role to workspace](https://www.eksworkshop.com/020_prerequisites/ec2instance/)
5. Update [IAM settings](https://www.eksworkshop.com/020_prerequisites/workspaceiam/)
6. [Create KMS customer managed keys](https://www.eksworkshop.com/020_prerequisites/kmskey/)

### Deploy an Amazon EKS cluster

2. Install [eksctl tools](https://www.eksworkshop.com/030_eksctl/prerequisites/)
3. Install [Helm](https://www.eksworkshop.com/beginner/060_helm/helm_intro/install/)
4. Launch an [EKS cluster](https://www.eksworkshop.com/030_eksctl/launcheks/)
5. Test [EKS cluster](https://www.eksworkshop.com/030_eksctl/test/)
6. (Optional) [Grant console access to EKS cluster](https://www.eksworkshop.com/030_eksctl/console/#:~:text=The%20EKS%20console%20allows%20you,granted%20permission%20within%20the%20cluster.)

## Install ACK service controller for RDS

To manage an Aurora Serverless v2 cluster from Kubernetes / Amazon EKS, you will need to install the ACK for RDS service controller. You can deploy the ACK service controller for Amazon RDS using the [rds-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/rds-chart).

Define environment variables

```
export SERVICE=rds
RELEASE_VERSION=$(curl -sL "https://api.github.com/repos/aws-controllers-k8s/${SERVICE}-controller/releases/latest" | grep '"tag_name":' | cut -d'"' -f4)
export ACK_SYSTEM_NAMESPACE=ack-system
export AWS_REGION=$(curl -s 169.254.169.254/latest/dynamic/instance-identity/document | jq -r '.region')
```
Log into the Helm registry that stores the ACK charts:

```bash
aws ecr-public get-login-password --region us-east-1 | \
  helm registry login --username AWS --password-stdin public.ecr.aws
```

You can now use the Helm chart to deploy the ACK service controller for Amazon RDS to your EKS cluster. At a minimum, you need to specify the AWS Region to execute the RDS API calls.

For example, to specify that the RDS API calls go to the `us-east-1` region, you can deploy the service controller with the following command:

```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
helm install --create-namespace -n "${ACK_SYSTEM_NAMESPACE}" "oci://public.ecr.aws/aws-controllers-k8s/${SERVICE}-chart" --version="${RELEASE_VERSION}" --generate-name --set=aws.region="${AWS_REGION}"
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/rds-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed, you will need to [configure the IAM permissions][irsa-permissions] for the controller to query the RDS API. For full details, please review the AWS Controllers for Kubernetes documentation for [how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the value of `rds` for `SERVICE`.

## Create an Aurora Serverless v2 PostgreSQL database

To create an Aurora Serverless v2 database using the PostgreSQL engine, you must
first create a DBSubnetGroup and a SecurityGroup for the VPC:

```bash
export APP_NAMESPACE=mydb
kubectl create ns "${APP_NAMESPACE}"

EKS_VPC_ID=$(aws eks describe-cluster --name "${EKS_CLUSTER_NAME}" --query "cluster.resourcesVpcConfig.vpcId" --output text)

RDS_SUBNET_GROUP_NAME="mydbSubnetGroup"
RDS_SUBNET_GROUP_DESCRIPTION="mydb-subnetgroup"
EKS_SUBNET_IDS=$(aws ec2 describe-subnets --filter "Name=vpc-id,Values=${EKS_VPC_ID}" --query 'Subnets[?MapPublicIpOnLaunch==`false`].SubnetId' --output text)

cat <<-EOF > db-subnet-groups.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBSubnetGroup
metadata:
 name: ${RDS_SUBNET_GROUP_NAME}
 namespace: ${APP_NAMESPACE}
spec:
 name: ${RDS_SUBNET_GROUP_NAME}
 description: ${RDS_SUBNET_GROUP_DESCRIPTION}
 subnetIDs:
$(printf " - %s\n" ${EKS_SUBNET_IDS})
 tags: []
EOF

kubectl apply -f db-subnet-groups.yaml

RDS_SECURITY_GROUP_NAME="ackSecurityGroup"
RDS_SECURITY_GROUP_DESCRIPTION="ACK security group"

EKS_CIDR_RANGE=$(aws ec2 describe-vpcs \
 --vpc-ids "${EKS_VPC_ID}" \
 --query "Vpcs[].CidrBlock" \
 --output text
)

RDS_SECURITY_GROUP_ID=$(aws ec2 create-security-group \
 --group-name "${RDS_SECURITY_GROUP_NAME}" \
 --description "${RDS_SECURITY_GROUP_DESCRIPTION}" \
 --vpc-id "${EKS_VPC_ID}" \
 --output text
)
aws ec2 authorize-security-group-ingress \
 --group-id "${RDS_SECURITY_GROUP_ID}" \
 --protocol tcp \
 --port 5432 \
 --cidr "${EKS_CIDR_RANGE}"
```

Set up a master password using a Kubernetes Secret. Set `RDS_DB_USERNAME` and `RDS_DB_PASSWORD` to your preferred values for your RDS credentials:

```bash

RDS_DB_USERNAME="adminer"
RDS_DB_PASSWORD="password"

kubectl create secret generic -n "${APP_NAMESPACE}" ack-creds \
  --from-literal=username="${RDS_DB_USERNAME}" \
  --from-literal=password="${RDS_DB_PASSWORD}"
```


You can now create an Aurora Serverless v2 cluster for both the PostgreSQL and
MySQL database engines. The example below uses the PostgreSQL engine. To use
MySQL, set `ENGINE_TYPE` to `aurora-mysql` and `ENGINE_VERSION` to `8.0`.


```bash
export AURORA_DB_CLUSTER_NAME="ack-db"
export AURORA_DB_INSTANCE_NAME="ack-db-instance01"
export AURORA_DB_INSTANCE_CLASS="db.serverless"
export MAX_ACU=64
export MIN_ACU=4

export ENGINE_TYPE=aurora-postgresql
export ENGINE_VERSION=13


cat <<-EOF > asv2-db-cluster.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBCluster
metadata:
  name: ${AURORA_DB_CLUSTER_NAME}
  namespace: ${APP_NAMESPACE}
spec:
  backupRetentionPeriod: 7
  serverlessV2ScalingConfiguration:
    maxCapacity: ${MAX_ACU}
    minCapacity: ${MIN_ACU}
  dbClusterIdentifier: ${AURORA_DB_CLUSTER_NAME}
  dbSubnetGroupName: ${RDS_SUBNET_GROUP_NAME}
  engine: ${ENGINE_TYPE}
  engineVersion: ${ENGINE_VERSION}
  masterUsername: adminer
  masterUserPassword:
    namespace: ${APP_NAMESPACE}
    name: ack-creds
    key: password
  vpcSecurityGroupIDs:
     - ${RDS_SECURITY_GROUP_ID}
EOF

kubectl apply -f asv2-db-cluster.yaml


cat <<-EOF > asv2-db-instance.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBInstance
metadata:
  name: ${AURORA_DB_INSTANCE_NAME}
  namespace: ${APP_NAMESPACE}
spec:
  dbInstanceClass: ${AURORA_DB_INSTANCE_CLASS}
  dbInstanceIdentifier: ${AURORA_DB_INSTANCE_NAME}
  dbClusterIdentifier: ${AURORA_DB_CLUSTER_NAME}
  dbSubnetGroupName: ${RDS_SUBNET_GROUP_NAME}
  engine: aurora-postgresql
  engineVersion: "13"
  publiclyAccessible: false
EOF

kubectl apply -f asv2-db-instance.yaml
```

{{% hint type="info" title="Required `serverlessV2ScalingConfiguration` attributes" %}}
In the `DBCluster` custom resource, you **must** set both the `minCapacity` and
`maxCapacity` attributes in the `serverlessV2ScalingConfiguration` section,
otherwise the database cluster will not be created.
{{% /hint %}}

To see your newly created Aurora Serverless v2 cluster, you can run the
following command:

```bash
kubectl describe -n "${APP_NAMESPACE}" "dbclusters/${AURORA_DB_CLUSTER_NAME}"
```

## Connect to Database Instances

The `DBInstance` status contains the information for connecting to a RDS database instance. The host information can be found in `status.endpoint.address` and the port information can be found in `status.endpoint.port`. The master user name can be found in `spec.masterUsername`.

The database password is in the Secret that is referenced in the `DBInstance` spec (`spec.masterPassword.name`).

You can extract this information and make it available to your Pods using a [`FieldExport`][field-export] resource. For example, to get the connection information from either RDS database instance created the above example, you can use the following example:

```bash
AURORA_INSTANCE_CONN_CM="asv2-db-instance-conn-cm"

cat <<EOF > asv2-db-field-exports.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${AURORA_INSTANCE_CONN_CM}
  namespace: ${APP_NAMESPACE}
data: {}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${AURORA_DB_INSTANCE_NAME}-host
  namespace: ${APP_NAMESPACE}
spec:
  to:
    name: ${AURORA_INSTANCE_CONN_CM}
    kind: configmap
  from:
    path: ".status.endpoint.address"
    resource:
      group: rds.services.k8s.aws
      kind: DBInstance
      name: ${AURORA_DB_INSTANCE_NAME}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${AURORA_DB_INSTANCE_NAME}-port
  namespace: ${APP_NAMESPACE}
spec:
  to:
    name: ${AURORA_INSTANCE_CONN_CM}
    kind: configmap
  from:
    path: ".status.endpoint.port"
    resource:
      group: rds.services.k8s.aws
      kind: DBInstance
      name: ${AURORA_DB_INSTANCE_NAME}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${AURORA_DB_INSTANCE_NAME}-user
  namespace: ${APP_NAMESPACE}
spec:
  to:
    name: ${AURORA_INSTANCE_CONN_CM}
    kind: configmap
  from:
    path: ".spec.masterUsername"
    resource:
      group: rds.services.k8s.aws
      kind: DBInstance
      name: ${AURORA_DB_INSTANCE_NAME}
EOF

kubectl apply -f asv2-db-field-exports.yaml
```

You can inject these values into a container either as environmental variables or files. For example, here is a snippet of a Pod definition that will add the RDS instance connection info into the Pod:

```bash
cat <<EOF > rds-pods.yaml
apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: ${APP_NAMESPACE}
spec:
  containers:
  -image: busybox
   name: myapp
   env:
    - name: PGHOST
      valueFrom:
        configMapKeyRef:
          name: ${AURORA_INSTANCE_CONN_CM}
          key: "${APP_NAMESPACE}.${AURORA_DB_INSTANCE_NAME}-host"
    - name: PGPORT
      valueFrom:
        configMapKeyRef:
          name: ${AURORA_INSTANCE_CONN_CM}
          key: "${APP_NAMESPACE}.${AURORA_DB_INSTANCE_NAME}-port"
    - name: PGUSER
      valueFrom:
        configMapKeyRef:
          name: ${AURORA_INSTANCE_CONN_CM}
          key: "${APP_NAMESPACE}.${AURORA_DB_INSTANCE_NAME}-user"
    - name: PGPASSWORD
      valueFrom:
        secretRef:
          name: "ack-creds"
          key: password
EOF

kubectl apply -f rds-pods.yaml
```

## Cleanup

You can delete your Aurora Serverless v2 cluster using the following command:

```bash
kubectl delete -f asv2-db-instance.yaml
kubectl delete -f asv2-db-cluster.yaml
```

To remove the RDS ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html
[field-export]: ../../user-docs/field-export
