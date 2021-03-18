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
"""Bootstraps the resources required to run the SageMaker integration tests.
"""

import boto3
import json
import logging
import time

from common.aws import get_aws_account_id, get_aws_region, duplicate_s3_contents
from common.resources import random_suffix_name
from sagemaker.bootstrap_resources import TestBootstrapResources, SAGEMAKER_SOURCE_DATA_BUCKET


def create_execution_role() -> str:
    region = get_aws_region()
    role_name = random_suffix_name(f"ack-sagemaker-execution-role", 63)
    iam = boto3.client("iam", region_name=region)

    iam.create_role(
        RoleName=role_name,
        AssumeRolePolicyDocument=json.dumps({"Version": "2012-10-17", "Statement": [
                                            {"Effect": "Allow", "Principal": {"Service": "sagemaker.amazonaws.com"}, "Action": "sts:AssumeRole"}]}),
        Description="SageMaker execution role for ACK integration and canary tests"
    )

    iam.attach_role_policy(
        RoleName=role_name,
        PolicyArn='arn:aws:iam::aws:policy/AmazonSageMakerFullAccess'
    )
    iam.attach_role_policy(
        RoleName=role_name,
        PolicyArn='arn:aws:iam::aws:policy/AmazonS3FullAccess'
    )

    iam_resource = iam.get_role(RoleName=role_name)
    resource_arn = iam_resource['Role']['Arn']

    # There appears to be a delay in role availability after role creation
    # resulting in failure that role is not present. So adding a delay
    # to allow for the role to become available
    time.sleep(10)
    logging.info(f"Created SageMaker execution role {resource_arn}")

    return resource_arn


def create_data_bucket() -> str:
    region = get_aws_region()
    account_id = get_aws_account_id()
    bucket_name = random_suffix_name(f"ack-data-bucket-{region}-{account_id}", 63)

    s3 = boto3.client("s3", region_name=region)
    if region == "us-east-1":
        s3.create_bucket(Bucket=bucket_name)
    else:
        s3.create_bucket(Bucket=bucket_name, CreateBucketConfiguration={
            'LocationConstraint': region
        })

    logging.info(f"Created SageMaker data bucket {bucket_name}")

    s3_resource = boto3.resource("s3", region_name=region)

    source_bucket = s3_resource.Bucket(SAGEMAKER_SOURCE_DATA_BUCKET)
    destination_bucket = s3_resource.Bucket(bucket_name)
    duplicate_s3_contents(source_bucket, destination_bucket)

    logging.info(f"Synced data bucket")

    return bucket_name


def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)

    return TestBootstrapResources(
        create_data_bucket(),
        create_execution_role(),
    ).__dict__
