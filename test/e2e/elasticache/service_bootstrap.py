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
"""Bootstraps the resources required to run Elasticache integration tests.
"""

import boto3
import logging

from common.aws import get_aws_account_id, get_aws_region
from common.resources import random_suffix_name
from dataclasses import dataclass
from elasticache.util import wait_usergroup_active, wait_snapshot_available

def create_sns_topic() -> str:
    topic_name = random_suffix_name("ack-sns-topic", 32)

    sns = boto3.client("sns")
    response = sns.create_topic(Name=topic_name)
    logging.info(f"Created SNS topic {response['TopicArn']}")

    return response['TopicArn']

# create an EC2 VPC security group from the default VPC (not an ElastiCache security group)
def create_security_group() -> str:
    region = get_aws_region()
    account_id = get_aws_account_id()

    ec2 = boto3.client("ec2")
    vpc_response = ec2.describe_vpcs(Filters=[{"Name": "isDefault", "Values": ["true"]}])
    if len(vpc_response['Vpcs']) == 0:
        raise ValueError(f"Default VPC not found for account {account_id} in region {region}")
    default_vpc_id = vpc_response['Vpcs'][0]['VpcId']

    sg_name = random_suffix_name("ack-security-group", 32)
    sg_description = "Security group for ACK ElastiCache tests"
    sg_response = ec2.create_security_group(GroupName=sg_name, VpcId=default_vpc_id, Description=sg_description)
    logging.info(f"Created VPC Security Group {sg_response['GroupId']}")

    return sg_response['GroupId']

def create_user_group() -> str:
    ec = boto3.client("elasticache")

    usergroup_id = random_suffix_name("ack-ec-usergroup", 32)
    _ = ec.create_user_group(UserGroupId=usergroup_id,
                                    Engine="Redis",
                                    UserIds=["default"])
    logging.info(f"Creating ElastiCache User Group {usergroup_id}")
    assert wait_usergroup_active(usergroup_id)

    return usergroup_id

def create_kms_key() -> str:
    kms = boto3.client("kms")

    response = kms.create_key(Description="Key for ACK ElastiCache tests")
    key_id = response['KeyMetadata']['KeyId']
    logging.info(f"Created KMS key {key_id}")

    return key_id

# create a cache cluster, snapshot it, and return the snapshot name
def create_cc_snapshot():
    ec = boto3.client("elasticache")

    cc_id = random_suffix_name("ack-cache-cluster", 32)
    _ = ec.create_cache_cluster(
        CacheClusterId=cc_id,
        NumCacheNodes=1,
        CacheNodeType="cache.m6g.large",
        Engine="redis"
    )
    waiter = ec.get_waiter('cache_cluster_available')
    waiter.wait(CacheClusterId=cc_id)
    logging.info(f"Created cache cluster {cc_id} for snapshotting")

    snapshot_name = random_suffix_name("ack-cc-snapshot", 32)
    _ = ec.create_snapshot(
        CacheClusterId=cc_id,
        SnapshotName=snapshot_name
    )
    assert wait_snapshot_available(snapshot_name)

    return snapshot_name

@dataclass
class BootstrapResources:
    SnsTopicARN: str
    SecurityGroupID: str
    UserGroupID: str
    KmsKeyID: str
    SnapshotName: str

def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)

    return BootstrapResources(
        create_sns_topic(),
        create_security_group(),
        create_user_group(),
        create_kms_key(),
        create_cc_snapshot()
    ).__dict__
