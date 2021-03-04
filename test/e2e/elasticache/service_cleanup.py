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
"""Cleans up the resources created by the Elasticache bootstrapping process.
"""

import boto3
import logging

from elasticache.service_bootstrap import BootstrapResources

def delete_sns_topic(topic_ARN: str):
    sns = boto3.client("sns")
    sns.delete_topic(TopicArn=topic_ARN)
    logging.info(f"Deleted SNS topic {topic_ARN}")

def delete_security_group(sg_id: str):
    ec2 = boto3.client("ec2")
    ec2.delete_security_group(GroupId=sg_id)
    logging.info(f"Deleted VPC Security Group {sg_id}")

def delete_user_group(usergroup_id: str):
    ec = boto3.client("elasticache")
    ec.delete_user_group(UserGroupId=usergroup_id)
    logging.info(f"Deleted ElastiCache User Group {usergroup_id}")

# KMS does not allow immediate key deletion; 7 days is the shortest deletion window
def delete_kms_key(key_id: str):
    kms = boto3.client("kms")
    kms.schedule_key_deletion(KeyId=key_id, PendingWindowInDays=7)
    logging.info(f"Deletion scheduled for KMS key {key_id}")

# delete snapshot and also associated cluster/RG
def delete_snapshot(snapshot_name: str):
    ec = boto3.client("elasticache")

    # delete actual snapshot
    response = ec.describe_snapshots(SnapshotName=snapshot_name)
    snapshot = response['Snapshots'][0]
    ec.delete_snapshot(SnapshotName=snapshot_name)
    logging.info(f"Deleted snapshot {snapshot_name}")

    # delete resource that was used to create snapshot
    if snapshot['CacheClusterId']:
        ec.delete_cache_cluster(CacheClusterId=snapshot['CacheClusterId'])
        logging.info(f"Deleted cache cluster {snapshot['CacheClusterId']}")
    elif snapshot['ReplicationGroupId']: # should not happen
        ec.delete_replication_group(ReplicationGroupId=snapshot['ReplicationGroupId'])
        logging.info(f"Deleted replication group {snapshot['ReplicationGroupId']}")

def service_cleanup(config: dict):
    logging.getLogger().setLevel(logging.INFO)

    resources = BootstrapResources(
        **config
    )

    try:
        delete_sns_topic(resources.SnsTopicARN)
    except:
        logging.exception(f"Unable to delete SNS topic {resources.SnsTopicARN}")

    try:
        delete_security_group(resources.SecurityGroupID)
    except:
        logging.exception(f"Unable to delete VPC Security Group {resources.SecurityGroupID}")

    try:
        delete_user_group(resources.UserGroupID)
    except:
        logging.exception(f"Unable to delete ElastiCache User Group {resources.UserGroupID}")

    try:
        delete_kms_key(resources.KmsKeyID)
    except:
        logging.exception(f"Unable to schedule deletion for KMS key {resources.KmsKeyID}")

    try:
        delete_snapshot(resources.SnapshotName)
    except:
        logging.exception(f"Unable to delete snapshot {resources.SnapshotName}")