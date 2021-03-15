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
"""Bootstraps the resources required to run the APIGatewayV2 integration tests.
"""
import logging
import os
import time
from zipfile import ZipFile
import random
import string
import tempfile

import boto3

from apigatewayv2.bootstrap_resources import TestBootstrapResources
from common.aws import get_aws_region

RAND_TEST_SUFFIX = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
AUTHORIZER_IAM_ROLE_NAME = 'ack-apigwv2-authorizer-role-' + RAND_TEST_SUFFIX
AUTHORIZER_ASSUME_ROLE_POLICY = '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": '\
                                '"lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]} '
AUTHORIZER_POLICY_ARN = 'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole'
AUTHORIZER_FUNCTION_NAME = 'ack-apigatewayv2-authorizer-' + RAND_TEST_SUFFIX


def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)
    authorizer_role_arn = create_authorizer_role()
    time.sleep(15)
    authorizer_function_arn = create_lambda_authorizer(authorizer_role_arn)

    return TestBootstrapResources(
        AUTHORIZER_IAM_ROLE_NAME,
        AUTHORIZER_POLICY_ARN,
        authorizer_role_arn,
        AUTHORIZER_FUNCTION_NAME,
        authorizer_function_arn
    ).__dict__


def create_authorizer_role() -> str:
    region = get_aws_region()
    iam_client = boto3.client("iam", region_name=region)

    logging.debug(f"Creating authorizer iam role {AUTHORIZER_IAM_ROLE_NAME}")

    try:
        iam_client.get_role(RoleName=AUTHORIZER_IAM_ROLE_NAME)
        raise RuntimeError(f"Expected {AUTHORIZER_IAM_ROLE_NAME} role to not exist."
                           f" Did previous test cleanup successfully?")
    except iam_client.exceptions.NoSuchEntityException:
        pass

    resp = iam_client.create_role(
        RoleName=AUTHORIZER_IAM_ROLE_NAME,
        AssumeRolePolicyDocument=AUTHORIZER_ASSUME_ROLE_POLICY
    )
    iam_client.attach_role_policy(RoleName=AUTHORIZER_IAM_ROLE_NAME, PolicyArn=AUTHORIZER_POLICY_ARN)
    return resp['Role']['Arn']


def create_lambda_authorizer(authorizer_role_arn : str) -> str:
    region = get_aws_region()
    lambda_client = boto3.client("lambda", region)

    try:
        lambda_client.get_function(FunctionName=AUTHORIZER_FUNCTION_NAME)
        raise RuntimeError(f"Expected {AUTHORIZER_FUNCTION_NAME} function to not exist. Did previous test cleanup"
                           f" successfully?")
    except lambda_client.exceptions.ResourceNotFoundException:
        pass

    with tempfile.TemporaryDirectory() as tempdir:
        current_directory = os.path.dirname(os.path.realpath(__file__))
        index_zip = ZipFile(f'{tempdir}/index.zip', 'w')
        index_zip.write(f'{current_directory}/resources/index.js', 'index.js')
        index_zip.close()

        with open(f'{tempdir}/index.zip', 'rb') as f:
            b64_encoded_zip_file = f.read()

        response = lambda_client.create_function(
            FunctionName=AUTHORIZER_FUNCTION_NAME,
            Role=authorizer_role_arn,
            Handler='index.handler',
            Runtime='nodejs12.x',
            Code={'ZipFile': b64_encoded_zip_file}
        )
    
    return response['FunctionArn']
