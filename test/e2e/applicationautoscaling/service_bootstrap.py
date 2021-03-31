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
from time import sleep

from common.resources import random_suffix_name
from common.aws import get_aws_account_id, get_aws_region
from applicationautoscaling.bootstrap_resources import TestBootstrapResources

def create_dynamodb_table() -> str:
    """Create a DynamoDB table with a randomised table name.

    Returns:
        str: The name of the DynamoDB table
    """
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
            "ReadCapacityUnits": 10,
            "WriteCapacityUnits": 10
        }
    )

    assert table["TableDescription"]["TableName"] == table_name
    logging.info(f"Created DynamoDB table {table_name}")

    return table_name

def wait_for_dynamodb_table_active(table_name: str, wait_periods: int = 6, period_length: int = 5) -> bool:
    """Wait for the given DynamoDB table to reach ACTIVE status.

    Args:
        table_name: The DynamoDB table to poll.
        wait_periods: The number of times to poll for the status.
        period_length: The delay between polling calls.

    Returns:
        bool: True if the table reached ACTIVE status.
    """
    region = get_aws_region()
    dynamodb = boto3.client("dynamodb", region_name=region)

    for _ in range(wait_periods):
        sleep(period_length)
        table = dynamodb.describe_table(
            TableName=table_name
        )

        status = table['Table']['TableStatus']
        if status == 'ACTIVE':
            logging.info(f"DynamoDB table {table_name} has reached status {status}")
            return True

        logging.debug(f"DynamoDB table {table_name} is in status {status}. Waiting...")

    logging.error(
        f"Wait for DynamoDB table {table_name} to become ACTIVE timed out")
    return False

def register_scalable_dynamodb_table(table_name: str) -> str:
    """Registers a DynamoDB table as a scalable target.

    Args:
        table_name: The DynamoDB table to register.

    Returns:
        str: The name of the DynamoDB table
    """
    region = get_aws_region()

    applicationautoscaling_client = boto3.client("application-autoscaling", region_name=region)
    applicationautoscaling_client.register_scalable_target(
        ServiceNamespace="dynamodb",
        ResourceId=f"table/{table_name}",
        ScalableDimension="dynamodb:table:WriteCapacityUnits",
        MinCapacity=50,
        MaxCapacity=100,
    )

    logging.info(f"Registered DynamoDB table {table_name} as scalable target")

    return table_name

def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)

    scalable_table = create_dynamodb_table()
    registered_table = create_dynamodb_table()

    if not wait_for_dynamodb_table_active(scalable_table) \
        or not wait_for_dynamodb_table_active(registered_table):
        raise Exception("DynamoDB tables did not become ACTIVE")

    return TestBootstrapResources(
        scalable_table,
        register_scalable_dynamodb_table(registered_table)
    ).__dict__
