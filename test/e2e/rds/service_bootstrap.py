# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Bootstraps the resources required to run the RDS integration tests.
"""

import boto3
import logging
import time

from common.aws import get_aws_account_id, get_aws_region, duplicate_s3_contents
from rds.bootstrap_resources import (
    TestBootstrapResources,
    VPC_CIDR,
    SUBNET_AZ1_CIDR,
    SUBNET_AZ2_CIDR,
)


def create_vpc() -> str:
    region = get_aws_region()
    ec2 = boto3.client("ec2", region_name=region)

    logging.debug(f"Creating VPC with CIDR {VPC_CIDR}")

    resp = ec2.create_vpc(
        CidrBlock=VPC_CIDR,
    )
    vpc_id = resp['Vpc']['VpcId']

    # TODO(jaypipes): Put a proper waiter here...
    time.sleep(3)

    vpcs = ec2.describe_vpcs(VpcIds=[vpc_id])
    if len(vpcs['Vpcs']) != 1:
        raise RuntimeError(
            f"failed to describe VPC we just created '{vpc_id}'",
        )

    vpc = vpcs['Vpcs'][0]
    vpc_state = vpc['State']
    if vpc_state != "available":
        raise RuntimeError(
            f"VPC we just created '{vpc_id}' is not available. current state: {vpc_state}",
        )

    logging.info(f"Created VPC {vpc_id}")

    return vpc_id


def create_subnet(vpc_id: str, az_id: str, cidr: str) -> str:
    region = get_aws_region()
    ec2 = boto3.client("ec2", region_name=region)

    logging.debug(f"Creating subnet with CIDR {cidr} in AZ {az_id}")

    resp = ec2.create_subnet(
        VpcId=vpc_id,
        AvailabilityZone=az_id,
        CidrBlock=cidr,
    )
    subnet_id = resp['Subnet']['SubnetId']

    # TODO(jaypipes): Put a proper waiter here...
    time.sleep(3)

    subnets  = ec2.describe_subnets(SubnetIds=[subnet_id])
    if len(subnets['Subnets']) != 1:
        raise RuntimeError(
            f"failed to describe subnet we just created '{subnet_id}'",
        )

    subnet = subnets['Subnets'][0]
    subnet_state = subnet['State']
    if subnet_state != "available":
        raise RuntimeError(
            f"Subnet we just created '{subnet_id}' is not available. current state: {subnet_state}",
        )

    logging.info(f"Created VPC Subnet {subnet_id} in AZ {az_id}")

    return subnet_id


def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)
    region = get_aws_region()

    vpc_id = create_vpc()
    az1 = f"{region}a"
    subnet_az1_id = create_subnet(vpc_id, az1, SUBNET_AZ1_CIDR)
    az2 = f"{region}b"
    subnet_az2_id = create_subnet(vpc_id, az2, SUBNET_AZ2_CIDR)


    return TestBootstrapResources(
        vpc_id,
        subnet_az1_id,
        subnet_az2_id,
    ).__dict__
