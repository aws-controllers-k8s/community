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
"""Integration tests for the SageMaker Model API.
"""

import boto3
import pytest
import logging
from typing import Dict

from sagemaker import (
    service_marker,
    create_sagemaker_resource,
    MODEL_RESOURCE_PLURAL,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import random_suffix_name
from common import k8s


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def xgboost_model():
    resource_name = random_suffix_name("xgboost-model", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["MODEL_NAME"] = resource_name

    reference, spec, resource = create_sagemaker_resource(
        resource_plural=MODEL_RESOURCE_PLURAL,
        resource_name=resource_name,
        spec_file="xgboost_model",
        replacements=replacements,
    )
    assert resource is not None

    yield (reference, resource)

    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(reference):
        k8s.delete_custom_resource(reference)


@service_marker
@pytest.mark.canary
class TestModel:
    def _get_resource_model_arn(self, resource: Dict):
        assert (
            "ackResourceMetadata" in resource["status"]
            and "arn" in resource["status"]["ackResourceMetadata"]
        )
        return resource["status"]["ackResourceMetadata"]["arn"]

    def _get_sagemaker_model_arn(self, sagemaker_client, model_name: str):
        try:
            model = sagemaker_client.describe_model(ModelName=model_name)
            return model["ModelArn"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a model with the name {model_name}"
            )
            return None

    def test_create_model(self, xgboost_model):
        (reference, resource) = xgboost_model
        assert k8s.get_resource_exists(reference)

    def test_model_has_correct_arn(self, sagemaker_client, xgboost_model):
        (reference, _) = xgboost_model
        resource = k8s.get_resource(reference)
        model_name = resource["spec"].get("modelName", None)

        assert model_name is not None

        assert self._get_resource_model_arn(resource) == self._get_sagemaker_model_arn(
            sagemaker_client, model_name
        )

    def test_model_is_deleted(self, sagemaker_client, xgboost_model):
        (reference, _) = xgboost_model
        resource = k8s.get_resource(reference)
        model_name = resource["spec"].get("modelName", None)

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        assert self._get_sagemaker_model_arn(sagemaker_client, model_name) is None
