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
"""Bootstraps the resources required to run the Application Auto Scaling
integration tests.
"""

import boto3
import logging

from common.resources import random_suffix_name
from common.aws import get_aws_account_id, get_aws_region
from applicationautoscaling.bootstrap_resources import TestBootstrapResources

def create_dynamodb_table() -> str:
    region = get_aws_region()
    account_id = get_aws_account_id()
    table_name = random_suffix_name(f"ack-autoscaling-table-{region}-{account_id}", 63)

    dynamodb = boto3.client("dynamodb", region_name=region)
    table = dynamodb.create_table(
        TableName=table_name,
        KeySchema=[
            {
                "AttributeName": "TablePrimaryAttribute",
                "KeyType": "HASH"
            }
        ],
        AttributeDefinitions=[
            {
                "AttributeName": "TablePrimaryAttribute",
                "AttributeType": "N"
            }
        ],
        ProvisionedThroughput={
            'ReadCapacityUnits': 10,
            'WriteCapacityUnits': 10
        }
    )

    assert table['TableDescription']['TableName'] == table_name
    logging.info(f"Created DynamoDB table {table_name}")

    return table_name

def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)

    return TestBootstrapResources(
        create_dynamodb_table()
    ).__dict__
