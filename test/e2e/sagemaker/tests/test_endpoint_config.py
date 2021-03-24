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
"""Integration tests for the SageMaker EndpointConfig API.
"""

import boto3
import pytest
import logging
from typing import Dict

from sagemaker import (
    service_marker,
    CONFIG_RESOURCE_PLURAL,
    MODEL_RESOURCE_PLURAL,
    create_sagemaker_resource,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import random_suffix_name
from common import k8s


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def single_variant_config():
    config_resource_name = random_suffix_name("single-variant-config", 32)
    model_resource_name = config_resource_name + "-model"

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CONFIG_NAME"] = config_resource_name
    replacements["MODEL_NAME"] = model_resource_name

    model_reference, model_spec, model_resource = create_sagemaker_resource(
        resource_plural=MODEL_RESOURCE_PLURAL,
        resource_name=model_resource_name,
        spec_file="xgboost_model",
        replacements=replacements,
    )
    assert model_resource is not None

    config_reference, config_spec, config_resource = create_sagemaker_resource(
        resource_plural=CONFIG_RESOURCE_PLURAL,
        resource_name=config_resource_name,
        spec_file="endpoint_config_single_variant",
        replacements=replacements,
    )
    assert config_resource is not None

    yield (config_reference, config_resource)

    k8s.delete_custom_resource(model_reference)
    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(config_reference):
        k8s.delete_custom_resource(config_reference)


@service_marker
@pytest.mark.canary
class TestEndpointConfig:
    def _get_resource_endpoint_config_arn(self, resource: Dict):
        assert (
            "ackResourceMetadata" in resource["status"]
            and "arn" in resource["status"]["ackResourceMetadata"]
        )
        return resource["status"]["ackResourceMetadata"]["arn"]

    def _get_sagemaker_endpoint_config_arn(self, sagemaker_client, config_name: str):
        try:
            response = sagemaker_client.describe_endpoint_config(
                EndpointConfigName=config_name
            )
            return response["EndpointConfigArn"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a config with the name {config_name}"
            )
            return None

    def test_create_endpoint_config(self, single_variant_config):
        (reference, resource) = single_variant_config
        assert k8s.get_resource_exists(reference)

    def test_config_has_correct_arn(self, sagemaker_client, single_variant_config):
        (reference, _) = single_variant_config
        resource = k8s.get_resource(reference)
        config_name = resource["spec"].get("endpointConfigName", None)

        assert config_name is not None

        assert self._get_resource_endpoint_config_arn(
            resource
        ) == self._get_sagemaker_endpoint_config_arn(sagemaker_client, config_name)

    def test_config_is_deleted(self, sagemaker_client, single_variant_config):
        (reference, _) = single_variant_config
        resource = k8s.get_resource(reference)
        config_name = resource["spec"].get("endpointConfigName", None)

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        assert (
            self._get_sagemaker_endpoint_config_arn(sagemaker_client, config_name)
            is None
        )
