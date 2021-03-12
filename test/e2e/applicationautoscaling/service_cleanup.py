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
"""Cleans up the resources created by the Application Auto Scaling bootstrapping process.
"""

import re
import boto3
import logging
from common.aws import get_aws_region
from applicationautoscaling.bootstrap_resources import TestBootstrapResources

def delete_dynamodb_table(table_name: str):
    region = get_aws_region()
    dynamodb = boto3.client("dynamodb", region_name=region)

    dynamodb.delete_table(TableName=table_name)

    logging.info(f"Deleted DynamoDB table {table_name}")

def service_cleanup(config: dict):
    logging.getLogger().setLevel(logging.INFO)

    resources = TestBootstrapResources(
        **config
    )

    try:
        delete_dynamodb_table(resources.ScalableDynamoTableName)
    except:
        logging.exception(f"Unable to delete DynamoDB table {resources.ScalableDynamoTableName}")

    try:
        delete_dynamodb_table(resources.RegisteredDynamoTableName)
    except:
        logging.exception(f"Unable to delete DynamoDB table {resources.RegisteredDynamoTableName}")