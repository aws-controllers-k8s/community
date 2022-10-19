---
title: "Manage a VPC Workflow with the ACK EC2-Controller"
description: "Create and manage a network topology using ACK EC2-Controller deployed on Amazon Elastic Kubernetes Service (EKS) The ACK service controller for Elastic Compute Cloud (EC2-Controller) lets users manage EC2 resources directly from Kubernetes. This guide demonstrates how to deploy a basic network topology (consisting of VPC resources) using a single Kubernetes resource manifest."
lead: "Create and manage a basic network topology using ACK EC2-Controller."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 43
toc: true
---
 

## Setup
Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) or Amazon Elastic Container Registry (Amazon ECR) with ACK, this guide assumes that you have 
access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster and Amazon ECR repository, see [Amazon EKS Setup][eks-setup] and [Amazon ECR Setup](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html). 
 

### Prerequisites
 
This guide assumes that you have:
 
- Created an EKS cluster with Kubernetes version 1.16 or higher.
- Have access to Amazon ECR
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
  - [Docker](https://docs.docker.com/engine/install/) - A tool to build, share, and run containers.
 
 
### Install ACK EC2-Controller
 
Deploy the EC2-Controller using the Helm chart, [ec2-chart](https://gallery.ecr.aws/aws-controllers-k8s/ec2-chart). Note, this example creates resources in the `us-west-2` region, but you can use any other region supported in AWS.
 
* Log into the Helm registry that stores the ACK charts:
 
```bash
aws ecr-public get-login-password --region us-west-2 | helm registry login --username AWS --password-stdin public.ecr.aws
```

* Install Helm chart:
 
```bash
SERVICE=ec2
RELEASE_VERSION=$(curl -sL "https://api.github.com/repos/aws-controllers-k8s/${SERVICE}-controller/releases/latest" | grep '"tag_name":' | cut -d'"' -f4)
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/ec2-chart "--version=${RELEASE_VERSION}" --generate-name --set=aws.region=us-west-2
```
 
For a full list of available values in the Helm chart, refer to [values.yaml](https://github.com/aws-controllers-k8s/ec2-controller/blob/main/helm/values.yaml).
 
 
### Configure IAM permissions
 
The controller requires permissions to invoke EC2 APIs. Once the service controller is deployed [configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) using the value `SERVICE=ec2` throughout. The recommended IAM Policy for EC2-Controller can be found in [recommended-policy-arn](https://github.com/aws-controllers-k8s/ec2-controller/blob/5414fc242d29af59309b4191b5a085439bf8730a/config/iam/recommended-policy-arn#L1).
 
 
### [Optional] Create a VPC and Subnet
 
This section is optional and will NOT be using a single manifest file to deploy the VPC and Subnet. The purpose of this section is to demonstrate a simple use case to shed light on some of the functionality before jumping into a more complex deployment.
 
* Create the **VPC** using the provided YAML and `kubectl apply`:
```
cat <<EOF > vpc.yaml
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: VPC
metadata:
  name: vpc-ack-test
spec:
  cidrBlocks: 
  - 10.0.0.0/16
  enableDNSSupport: true
  enableDNSHostnames: true
EOF
 
kubectl apply -f vpc.yaml
```
 
* Check the **VPC** `Status` using `kubectl describe`:
```
> kubectl describe vpcs
...
Status:
  Ack Resource Metadata:
    Owner Account ID:  <ID>
    Region:            us-west-2
  Cidr Block Association Set:
    Association ID:  vpc-cidr-assoc-<ID>
    Cidr Block:      10.0.0.0/16
    Cidr Block State:
      State:  associated
  Conditions:
    Last Transition Time:  2022-10-12T17:26:08Z
    Message:               Resource synced successfully
    Reason:
    Status:                True
    Type:                  ACK.ResourceSynced
  Dhcp Options ID:         dopt-<ID>
  Is Default:              false
  Owner ID:                <ID>
  State:                   available
  Vpc ID:                  vpc-<ID>
Events:                    <none>
```
 
* The **VPC** resource synced successfully and is available. Note the `vpc-<ID>`.
 
* Create the **Subnet** using `vpc-<ID>`, the provided YAML, and `kubectl apply`:
```
cat <<EOF > subnet.yaml
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: Subnet
metadata:
  name: subnet-ack-test
spec:
  cidrBlock: 10.0.0.0/20
  vpcID: vpc-<ID>
EOF
 
kubectl apply -f subnet.yaml
```
 
* Check the **Subnet** availability and ID using `kubectl describe`:
```
> kubectl describe subnets
...
Status:
  Ack Resource Metadata:
    Arn:                       arn:aws:ec2:us-west-2:<ID>:subnet/subnet-<ID>
    Owner Account ID:          <ID>
    Region:                    us-west-2
  Available IP Address Count:  4091
  Conditions:
    Last Transition Time:           2022-10-12T17:36:53Z
    Message:                        Resource synced successfully
    Reason:
    Status:                         True
    Type:                           ACK.ResourceSynced
  Default For AZ:                   false
  Map Customer Owned IP On Launch:  false
  Owner ID:                         <ID>
  Private DNS Name Options On Launch:
  State:      available
  Subnet ID:  subnet-0f0eb8efb37eb43b0
Events:       <none>
```
 
* Delete the resources:
  * `kubectl delete -f subnet.yaml`
  * `kubectl delete -f vpc.yaml`
 
 
Both resources were successfully deployed, managed, then deleted by their respective controllers. Although contrived, this example highlights how easy it can be to deploy AWS resources via YAML files and how it feels like managing any other K8s resource. 
 
In this example, waiting for the `vpc-ID` in order to create the **Subnet** was sub-optimal; the workflow didn't feel declarative. The next example alleviates these concerns using a single manifest (YAML) to deploy the entire network topology.
 
 
### Create a VPC Workflow
 
This section details the steps to create a network topology consisting of multiple, connected resources from a single manifest file. The following resources will be present in said manifest:
* 1 VPC
* 1 Instance
* 1 Internet Gateway
* 1 NAT Gateways
* 1 Elastic IPs
* 2 Route Tables
* 2 Subnets (1 Public; 1 Private)
* 1 Security Group
 
 
The VPC is connected to the internet through an Internet Gateway. The NAT Gateway is created in the public Subnet with an associated Elastic IP. The Instance is deployed into the private Subnet which can connect to the internet using the NAT Gateway in the public Subnet. Lastly, one Route Table (public) will contain a route to the Internet Gateway while the other Route Table (private) contains a route to the NAT Gateway.

* Deploy the resources using the provided YAML and `kubectl apply -f vpc-workflow.yaml`:
 
```
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: VPC
metadata:
  name: ack-vpc
spec:
  cidrBlocks: 
  - 10.0.0.0/16
  enableDNSSupport: true
  enableDNSHostnames: true
  tags:
    - key: purpose
      value: ack-tutorial
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: InternetGateway
metadata:
  name: ack-igw
spec:
  vpcRef:
    from:
      name: ack-vpc
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: NATGateway
metadata:
  name: ack-natgateway1
spec:
  subnetRef:
    from:
      name: ack-public-subnet1
  allocationRef:
    from:
      name: ack-eip1
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: ElasticIPAddress
metadata:
  name: ack-eip1
spec:
  tags:
    - key: purpose
      value: ack-tutorial
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: RouteTable
metadata:
  name: ack-public-route-table
spec:
  vpcRef:
    from:
      name: ack-vpc
  routes:
  - destinationCIDRBlock: 0.0.0.0/0
    gatewayRef:
      from:
        name: ack-igw
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: RouteTable
metadata:
  name: ack-private-route-table-az1
spec:
  vpcRef:
    from:
      name: ack-vpc
  routes:
  - destinationCIDRBlock: 0.0.0.0/0
    natGatewayRef:
      from:
        name: ack-natgateway1
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: Subnet
metadata:
  name: ack-public-subnet1
spec:
  availabilityZone: us-west-2a
  cidrBlock: 10.0.0.0/20
  mapPublicIPOnLaunch: true
  vpcRef:
    from:
      name: ack-vpc
  routeTableRefs:
  - from:
      name: ack-public-route-table
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: Subnet
metadata:
  name: ack-private-subnet1
spec:
  availabilityZone: us-west-2a
  cidrBlock: 10.0.128.0/20
  vpcRef:
    from:
      name: ack-vpc
  routeTableRefs:
  - from:
      name: ack-private-route-table-az1
---
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: SecurityGroup
metadata:
  name: ack-security-group
spec:
  description: "ack security group"
  name: ack-sg
  vpcRef:
     from:
       name: ack-vpc
  ingressRules:
    - ipProtocol: tcp
      fromPort: 22
      toPort: 22
      ipRanges:
        - cidrIP: "0.0.0.0/0"
          description: "ingress"
```
 
 
{{% hint type="info" title="Referencing Resources" %}}
Notice that the ACK custom resources reference each other using "*Ref" fields inside the manifest and the user does not have to worry about finding `vpc-ID` when creating the Subnet resource manifests.
 
Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *EC2*
to find the supported reference fields.
{{% /hint %}}
 
 
The output should look similar to:
```
vpc.ec2.services.k8s.aws/ack-vpc created
internetgateway.ec2.services.k8s.aws/ack-igw created
routetable.ec2.services.k8s.aws/ack-public-route-table created
routetable.ec2.services.k8s.aws/ack-private-route-table created
natgateway.ec2.services.k8s.aws/ack-natgateway created
elasticipaddress.ec2.services.k8s.aws/ack-eip created
subnet.ec2.services.k8s.aws/ack-public-subnet created
subnet.ec2.services.k8s.aws/ack-private-subnet created
securitygroup.ec2.services.k8s.aws/ack-security-group created
```
 
* Check the **Custom Resource's** using `kubectl describe`:
```
kubectl describe vpcs
kubectl describe internetgateways
kubectl describe routetables
kubectl describe natgateways
kubectl describe elasticipaddresses
kubectl describe subnets
kubectl describe securitygroups
```
 
* Subnet gets into an 'available' state with a `ACK.ReferencesResolved = True` condition attached notifying users that the references (VPC, RouteTable) have been found and resolved:
 
```
Status:
  Ack Resource Metadata:
    Arn:                       arn:aws:ec2:us-west-2:<ID>:subnet/subnet-0ba22f5820bb41584
    Owner Account ID:          <ID>
    Region:                    us-west-2
  Available IP Address Count:  4091
  Conditions:
    Last Transition Time:           2022-10-13T14:54:39Z
    Status:                         True
    Type:                           ACK.ReferencesResolved
    Last Transition Time:           2022-10-13T14:54:41Z
    Message:                        Resource synced successfully
    Reason:
    Status:                         True
    Type:                           ACK.ResourceSynced
  Default For AZ:                   false
  Map Customer Owned IP On Launch:  false
  Owner ID:                         515336597380
  Private DNS Name Options On Launch:
  State:      available
  Subnet ID:  subnet-<ID>
 
```
 
### Validate
 
This network setup should allow Instances deployed in the Private Subnet to connect to the internet. To validate this behavior deploy an Instance into the Private Subnet and the Public Subnet (bastion host). After deployments, `ssh` into the bastion host, then `ssh` into the Private Subnet Instance, and test internet connection. Security group is required by both instances launched in public and private subnets.
 
* Deploy an Instance into the Private Subnet:
 
```
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: Instance
metadata:
  name: ack-instance-private
spec:
  imageID: ami-02b92c281a4d3dc79 # AL2; us-west-2
  instanceType: c3.large
  subnetID: subnet-<private-ID>
  securityGroupIDs:
  	- sg-<ID>
  keyName: us-west-2-key # created via console
  tags:
    - key: producer
      value: ack
```
 
* Deploy an Instance into the Public Subnet:
 
```
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: Instance
metadata:
  name: ack-instance-public
spec:
  imageID: ami-02b92c281a4d3dc79 # AL2 in us-west-2
  instanceType: c3.large
  subnetID: subnet-<public-ID>
  securityGroupIDs:
  	- sg-<ID>
  keyName: us-west-2-key # created via console
  tags:
    - key: producer
      value: ack
```
 
* Deployed 2 instances; one to each Subnet
  * The instance in the public subnet will be the bastion host so we can ssh to the Instance in the private Subnet
    ```bash
    scp "/path/created_key_in_console_for_region.pem" ec2-user@<Public IPV4 DNS>:
    ssh -i "/path/created_key_in_console_for_region.pem" ec2-user@<Public IPV4 DNS>
    ssh -i "created_key_in_console_for_region.pem" ec2-user@<Private IP>
    ```
* Validate instance in private subnet can connect to internet
  * Try to ping websites from your private subnet, sample output looks like
    ```bash
    ping google.com
  
    PING google.com (142.250.217.78) 56(84) bytes of data.
    64 bytes from sea09s29-in-f14.1e100.net (142.250.217.78): icmp_seq=1 ttl=102 time=8.30 ms
    64 bytes from sea09s29-in-f14.1e100.net (142.250.217.78): icmp_seq=2 ttl=102 time=7.82 ms
    64 bytes from sea09s29-in-f14.1e100.net (142.250.217.78): icmp_seq=3 ttl=102 time=7.77 ms
    ^C
    --- google.com ping statistics ---
    3 packets transmitted, 3 received, 0% packet loss, time 2003ms
    ```
 
### Cleanup
 
Remove all the resources using `kubectl delete` command.

```bash
kubectl delete -f ack-instance-public.yaml
kubectl delete -f ack-instance-private.yaml
kubectl delete -f vpc-workflow.yaml
```
 
The output of delete commands should look like
 
```bash
instance.ec2.services.k8s.aws/ack-instance-public.yaml deleted
instance.ec2.services.k8s.aws/ack-instance-private.yaml deleted
vpc.ec2.services.k8s.aws/ack-vpc deleted
internetgateway.ec2.services.k8s.aws/ack-igw deleted
routetable.ec2.services.k8s.aws/ack-public-route-table deleted
routetable.ec2.services.k8s.aws/ack-private-route-table deleted
natgateway.ec2.services.k8s.aws/ack-natgateway deleted
elasticipaddress.ec2.services.k8s.aws/ack-eip deleted
subnet.ec2.services.k8s.aws/ack-public-subnet deleted
subnet.ec2.services.k8s.aws/ack-private-subnet deleted
```
 
To remove the EC2 ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].
 
To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].
 
[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html