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
"""Integration tests for the SageMaker TransformJob API.
"""

import boto3
import pytest
import logging
from typing import Dict

from sagemaker import (
    service_marker,
    create_sagemaker_resource,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import random_suffix_name
from common import k8s
from common.aws import get_aws_region
from sagemaker.bootstrap_resources import get_bootstrap_resources

RESOURCE_PLURAL = "transformjobs"


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def xgboost_transformjob(sagemaker_client):
    # Create model using boto3 for TransformJob
    transform_model_file = (
        f"s3://{get_bootstrap_resources().DataBucketName}/sagemaker/batch/model.tar.gz"
    )
    model_name = random_suffix_name("xgboost-model", 32)

    create_response = sagemaker_client.create_model(
        ModelName=model_name,
        PrimaryContainer={
            "Image": REPLACEMENT_VALUES["XGBOOST_IMAGE_URI"],
            "ModelDataUrl": transform_model_file,
            "Environment": {},
        },
        ExecutionRoleArn=REPLACEMENT_VALUES["SAGEMAKER_EXECUTION_ROLE_ARN"],
    )
    logging.debug(create_response)

    # Check if the model is created successfully
    describe_model_response = sagemaker_client.describe_model(ModelName=model_name)
    assert describe_model_response["ModelName"] is not None

    resource_name = random_suffix_name("xgboost-transformjob", 32)

    # Use the model created above
    replacements = REPLACEMENT_VALUES.copy()
    replacements["MODEL_NAME"] = model_name
    replacements["TRANSFORM_JOB_NAME"] = resource_name

    reference, spec, resource = create_sagemaker_resource(
        resource_plural=RESOURCE_PLURAL,
        resource_name=resource_name,
        spec_file="xgboost_transformjob",
        replacements=replacements,
    )
    assert resource is not None

    yield (reference, resource)

    # Delete the model created
    sagemaker_client.delete_model(ModelName=model_name)

    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(reference):
        k8s.delete_custom_resource(reference)


@service_marker
@pytest.mark.canary
class TestTransformJob:
    def _get_created_transformjob_status_list(self):
        return ["InProgress"]

    def _get_stopped_transformjob_status_list(self):
        return ["Stopped", "Stopping", "Completed"]

    def _get_resource_transformjob_arn(self, resource: Dict):
        assert (
            "ackResourceMetadata" in resource["status"]
            and "arn" in resource["status"]["ackResourceMetadata"]
        )
        return resource["status"]["ackResourceMetadata"]["arn"]

    def _get_sagemaker_transformjob_arn(self, sagemaker_client, transformjob_name: str):
        try:
            transformjob = sagemaker_client.describe_transform_job(
                TransformJobName=transformjob_name
            )
            return transformjob["TransformJobArn"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a transformJob with the name {transformjob_name}"
            )
            return None

    def _get_sagemaker_transformjob_status(
        self, sagemaker_client, transformjob_name: str
    ):
        try:
            transformjob = sagemaker_client.describe_transform_job(
                TransformJobName=transformjob_name
            )
            return transformjob["TransformJobStatus"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a transformJob with the name {transformjob_name}"
            )
            return None

    def test_create_transformjob(self, xgboost_transformjob):
        (reference, _) = xgboost_transformjob
        assert k8s.get_resource_exists(reference)

    def test_transformjob_has_correct_arn(self, sagemaker_client, xgboost_transformjob):
        (reference, resource) = xgboost_transformjob
        transformjob_name = resource["spec"].get("transformJobName", None)

        assert transformjob_name is not None

        resource_transformjob_arn = self._get_resource_transformjob_arn(resource)
        assert (
            self._get_sagemaker_transformjob_arn(sagemaker_client, transformjob_name)
        ) == resource_transformjob_arn

    def test_transformjob_has_created_status(
        self, sagemaker_client, xgboost_transformjob
    ):
        (reference, resource) = xgboost_transformjob
        transformjob_name = resource["spec"].get("transformJobName", None)

        assert transformjob_name is not None

        assert (
            self._get_sagemaker_transformjob_status(sagemaker_client, transformjob_name)
        ) in self._get_created_transformjob_status_list()

    def test_transformjob_has_stopped_status(
        self, sagemaker_client, xgboost_transformjob
    ):
        (reference, resource) = xgboost_transformjob
        transformjob_name = resource["spec"].get("transformJobName", None)

        assert transformjob_name is not None

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        assert (
            self._get_sagemaker_transformjob_status(sagemaker_client, transformjob_name)
        ) in self._get_stopped_transformjob_status_list()
