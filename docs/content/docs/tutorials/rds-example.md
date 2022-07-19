---
title: "Deploy PostgreSQL + MariaDB Instances Using the ACK RDS Controller"
description: "Create managed PostgreSQL or MariaDB instances in Amazon Relational Database Service (RDS) from a Amazon Elastic Kubernetes Service (EKS) deployment."
lead: "Create and use PostgreSQL or MariaDB instances in Amazon RDS using Amazon Elastic Kubernetes Service (EKS)."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 42
toc: true
---

The ACK service controller for Amazon Relational Database Service (RDS) lets you manage RDS database instances directly from Kubernetes. This includes the following database engines:

* [Amazon Aurora](https://aws.amazon.com/rds/aurora/) (MySQL & PostgreSQL)
* [Amazon RDS for PostgreSQL](https://aws.amazon.com/rds/postgresql/)
* [Amazon RDS for MySQL](https://aws.amazon.com/rds/mysql/)
* [Amazon RDS for MariaDB](https://aws.amazon.com/rds/mariadb/)
* [Amazon RDS for Oracle](https://aws.amazon.com/rds/oracle/)
* [Amazon RDS for SQL Server](https://aws.amazon.com/rds/sqlserver/)

This guide will show you how to create and connect to several types of database engines available in [Amazon RDS](https://aws.amazon.com/rds/) through Kubernetes.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see [Amazon EKS Setup](https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html). For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with Amazon EC2 Linux managed nodes.

### Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.16 or higher.
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.7+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.

### Install the ACK service controller for RDS

You can deploy the ACK service controller for Amazon RDS using the [rds-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/rds-chart).

Log into the Helm registry that stores the ACK charts:
```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

You can now use the Helm chart to deploy the ACK service controller for Amazon RDS to your EKS cluster. At a minimum, you need to specify the AWS Region to execute the RDS API calls.

For example, to specify that the RDS API calls go to the `us-east-1` region, you can deploy the service controller with the following command:

```bash
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/rds-chart --version=v0.0.27 --generate-name --set=aws.region=us-east-1
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/rds-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed, you will need to [configure the IAM permissions][irsa-permissions] for the controller to query the RDS API. For full details, please review the AWS Controllers for Kubernetes documentation for [how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the value of `rds` for `SERVICE`.

## Deploy Database Instances

You can deploy most RDS database instances using the `DBInstance` custom resource. The examples below show how to deploy using different database engines in RDS from your Kubernetes environment. For a full list of options available in the `DBInstance` custom resource definition, you can use `kubectl explain dbinstance`.

The examples below use the `db.t4g.micro` instance type. Please review the [RDS instance types](https://aws.amazon.com/rds/instance-types/) to select the most appropriate one for your workload.

### PostgreSQL

To create a [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/postgresql/) instance, you must first set up a master password. You can do this by [creating a Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret), e.g.:

```bash
RDS_INSTANCE_NAME="<your instance name>"

kubectl create secret generic "${RDS_INSTANCE_NAME}-password" \
  --from-literal=password="<your password>"
```

Next, create a `DBInstance` custom resource. The example below shows how to provision a RDS for PostgreSQL 14 instance with the credentials created in the previous step:

```bash
cat <<EOF > rds-postgresql.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBInstance
metadata:
  name: "${RDS_INSTANCE_NAME}"
spec:
  allocatedStorage: 20
  dbInstanceClass: db.t4g.micro
  dbInstanceIdentifier: "${RDS_INSTANCE_NAME}"
  engine: postgres
  engineVersion: "14"
  masterUsername: "postgres"
  masterUserPassword:
    namespace: default
    name: "${RDS_INSTANCE_NAME}-password"
    key: password
EOF

kubectl apply -f rds-postgresql.yaml
```

You can track the status of the provisioned database using `kubectl describe` on the `DBInstance` custom resource:

```bash
kubectl describe dbinstance "${RDS_INSTANCE_NAME}"
```

When the `DB Instance Status` says `Available`, you can connect to the database instance.

### MariaDB

To create a [AWS RDS for MariaDB](https://aws.amazon.com/rds/mariadb/) instance, you must first set up a master password. You can do this by [creating a Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret), e.g.:

```bash
RDS_INSTANCE_NAME="<your instance name>"

kubectl create secret generic "${RDS_INSTANCE_NAME}-password" \
  --from-literal=password="<your password>"
```

Next, create a `DBInstance` custom resource. The example below shows how to provision a RDS for MariaDB 10.6 instance with the credentials created in the previous step:

```bash
RDS_INSTANCE_NAME="<your instance name>"

cat <<EOF > rds-mariadb.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBInstance
metadata:
  name: "${RDS_INSTANCE_NAME}"
spec:
  allocatedStorage: 20
  dbInstanceClass: db.t4g.micro
  dbInstanceIdentifier: "${RDS_INSTANCE_NAME}"
  engine: mariadb
  engineVersion: "10.6"
  masterUsername: "admin"
  masterUserPassword:
    namespace: default
    name: "${RDS_INSTANCE_NAME}-password"
    key: password
EOF

kubectl apply -f rds-mariadb.yaml
```

You can track the status of the provisioned database by describing the DBInstance custom resource:

```bash
kubectl describe dbinstance "${RDS_INSTANCE_NAME}"
```

When the `DB Instance Status` says `Available`, you can connect to the database instance.

## Connect to Database Instances

The `DBInstance` status contains the information for connecting to a RDS database instance. The host information can be found in `status.endpoint.address` and the port information can be found in `status.endpoint.port`. The master user name can be found in `spec.masterUsername`.

The database password is in the Secret that is referenced in the `DBInstance` spec (`spec.masterPassword.name`).

You can extract this information and make it available to your Pods using a [`FieldExport`][field-export] resource. For example, to get the connection information from either RDS database instance created the above example, you can use the following example:

```bash
RDS_INSTANCE_CONN_CM="${RDS_INSTANCE_NAME}-conn-cm"

cat <<EOF > rds-field-exports.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${RDS_INSTANCE_CONN_CM}
data: {}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${RDS_INSTANCE_NAME}-host
spec:
  to:
    name: ${RDS_INSTANCE_CONN_CM}
    kind: configmap
  from:
    path: ".status.endpoint.address"
    resource:
      group: rds.services.k8s.aws
      kind: DBInstance
      name: ${RDS_INSTANCE_NAME}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${RDS_INSTANCE_NAME}-port
spec:
  to:
    name: ${RDS_INSTANCE_CONN_CM}
    kind: configmap
  from:
    path: ".status.endpoint.port"
    resource:
      group: rds.services.k8s.aws
      kind: DBInstance
      name: ${RDS_INSTANCE_NAME}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: ${RDS_INSTANCE_NAME}-user
spec:
  to:
    name: ${RDS_INSTANCE_CONN_CM}
    kind: configmap
  from:
    path: ".spec.masterUsername"
    resource:
      group: rds.services.k8s.aws
      kind: DBInstance
      name: ${RDS_INSTANCE_NAME}
EOF

kubectl apply -f rds-field-exports.yaml
```

You can inject these values into a container either as environmental variables or files. For example, here is a snippet of a Pod definition that will add the RDS instance connection info into the Pod:

```bash
cat <<EOF > rds-pods.yaml
apiVersion: v1
kind: Pod
metadata:
  name: app
spec:
  containers:
  - env:
    - name: PGHOST
      valueFrom:
        configMapKeyRef:
          name: ${RDS_INSTANCE_CONN_CM}
          key: "default.${RDS_INSTANCE_NAME}-host"
    - name: PGPORT
      valueFrom:
        configMapKeyRef:
          name: ${RDS_INSTANCE_CONN_CM}
          key: "default.${RDS_INSTANCE_NAME}-port"
    - name: PGUSER
      valueFrom:
        configMapKeyRef:
          name: ${RDS_INSTANCE_CONN_CM}
          key: "default.${RDS_INSTANCE_NAME}-user"
    - name: PGPASSWORD
      valueFrom:
        secretRef:
          name: "${RDS_INSTANCE_NAME}-password"
          key: password
EOF
```

## Restore a Database Snapshot

You can also restore a database snapshot to a specific `DBInstance` or `DBCluster` using the ACK for RDS controller.

To restore a database snapshot to a `DBInstance`, you must set the `Spec.DBSnapshotIdentifier` parameter. `Spec.DBSnapshotIdentifier` should match the identifier of an existing DBSnapshot.

To restore a database snapshot to a `DBCluster`, you must set the `Spec.SnapshotIdentifier`. The value of `Spec.SnapshotIdentifier` should match either an existing `DBCluster` snapshot identifier or an ARN of a `DBInstance`snapshot.

Once it's set and the resource is created, updating `Spec.SnapshotIdentifer` or `Spec.BSnapshotIdentifier` fields will have no effect.

The following examples show how you can restore database snapshots both to `DBCluster` and `DBInstance` resources:

```bash
RDS_CLUSTER_NAME="<your cluster name>"
RDS_REGION="<your aws region>"
RDS_CUSTOMER_ACCOUNT="<your aws account id>"
RDS_DB_SNAPSHOT_IDENTIFIER="<your db snapshot identifier>"

cat <<EOF > rds-restore-dbcluster-snapshot.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBCluster
metadata:
  name: "${RDS_CLUSTER_NAME}"
spec:
  dbClusterIdentifier: "${RDS_CLUSTER_NAME}"
  engine: aurora-postgresql
  engineVersion: "13.7"
  snapshotIdentifier: arn:aws:rds:${RDS_REGION}:${RDS_CUSTOMER_ACCOUNT}:snapshot:${RDS_DB_SNAPSHOT_IDENTIFIER}
EOF

kubectl apply -f rds-restore-dbcluster-snapshot.yaml
```

```bash
RDS_INSTANCE_NAME="<your instance name>"
RDS_DB_SNAPSHOT_IDENTIFIER="<your db snapshot identifier>"

cat <<EOF > rds-restore-dbinstance-snapshot.yaml
apiVersion: rds.services.k8s.aws/v1alpha1
kind: DBInstance
metadata:
  name: "${RDS_INSTANCE_NAME}"
spec:
  allocatedStorage: 20
  dbInstanceClass: db.m5.large
  dbInstanceIdentifier: "${RDS_INSTANCE_NAME}"
  engine: postgres
  engineVersion: "14"
  masterUsername: "postgres"
  multiAZ: true
  dbSnapshotIdentifier: "${RDS_DB_SNAPSHOT_IDENTIFIER}"
EOF

kubectl apply -f rds-restore-dbinstance-snapshot.yaml
```

## Next steps

You can learn more about each of the ACK service controller for RDS custom resources by using `kubectl explain` on the API resources. These include:

* `dbinstance`
* `dbparametergroup`
* `dbcluster`
* `dbclusterparametergroup`
* `dbsecuritygroup`
* `dbsubnetgroup`
* `globalclusters`

The ACK service controller for Amazon RDS is based on the [Amazon RDS API](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/). To get a full understanding of how all of the APIs work, please review the [Amazon RDS API documentation](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/).

You can learn more about [how to use Amazon RDS](https://docs.aws.amazon.com/rds/index.html) through the [documentation](https://docs.aws.amazon.com/rds/index.html).

### Cleanup

You can deprovision your RDS instances using `kubectl delete` command.

```bash
kubectl delete dbinstance "${RDS_INSTANCE_NAME}"
```

To remove the RDS ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html
[field-export]: ../../user-docs/field-export
