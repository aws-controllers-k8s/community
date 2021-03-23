# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Contains helper methods used across multiple SageMaker tests.
"""
import logging
import time
import boto3

__sagemaker_client = None

def _sagemaker_client():
    global __sagemaker_client
    if __sagemaker_client is None:
        __sagemaker_client = boto3.client('sagemaker')
    return __sagemaker_client

def _get_job_definition_arn(resource: object):
    if 'status' not in resource:
        return None
    return resource['status'].get('jobDefinitionARN')

def _wait_sagemaker_endpoint_status(
    endpoint_name,
    expected_status: str,
    wait_periods: int = 18,
):
    actual_status = None
    for _ in range(wait_periods):
        time.sleep(30)
        actual_status = _sagemaker_client().describe_endpoint(
            EndpointName=endpoint_name
        )["EndpointStatus"]
        if actual_status == expected_status:
            break
    else:
        logging.error(
            f"Wait for sagemaker endpoint status: {expected_status} timed out. Actual status: {actual_status}"
        )

    return actual_status