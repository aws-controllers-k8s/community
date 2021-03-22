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

"""Cleans up the resources created by the bootstrapping process.
"""

import boto3
import logging
from common.aws import get_aws_region
from rds.bootstrap_resources import TestBootstrapResources


def delete_subnet(subnet_id: str):
    region = get_aws_region()
    ec2 = boto3.client("ec2", region_name=region)

    ec2.delete_subnet(SubnetId=subnet_id)

    logging.info(f"Deleted VPC Subnet {subnet_id}")


def delete_vpc(vpc_id: str):
    region = get_aws_region()
    ec2 = boto3.client("ec2", region_name=region)

    ec2.delete_vpc(VpcId=vpc_id)

    logging.info(f"Deleted VPC {vpc_id}")


def service_cleanup(config: dict):
    logging.getLogger().setLevel(logging.INFO)

    resources = TestBootstrapResources(
        **config
    )

    try:
        delete_subnet(resources.SubnetAZ1)
    except:
        logging.exception(f"Unable to delete VPC subnet {resources.SubnetAZ1}")

    try:
        delete_subnet(resources.SubnetAZ2)
    except:
        logging.exception(f"Unable to delete VPC subnet {resources.SubnetAZ2}")

    try:
        delete_vpc(resources.VPCID)
    except:
        logging.exception(f"Unable to delete VPC {resources.VPCID}")
