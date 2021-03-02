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
    SERVICE_NAME,
    service_marker,
    CRD_GROUP,
    CRD_VERSION,
    CONFIG_RESOURCE_PLURAL,
    MODEL_RESOURCE_PLURAL,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import load_resource_file, random_suffix_name
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

    model = load_resource_file(
        SERVICE_NAME, "xgboost_model", additional_replacements=replacements
    )
    logging.debug(model)

    config = load_resource_file(
        SERVICE_NAME,
        "endpoint_config_single_variant",
        additional_replacements=replacements,
    )
    logging.debug(config)

    # Create the k8s resources
    model_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        MODEL_RESOURCE_PLURAL,
        model_resource_name,
        namespace="default",
    )
    model_resource = k8s.create_custom_resource(model_reference, model)
    model_resource = k8s.wait_resource_consumed_by_controller(model_reference)
    assert model_resource is not None

    config_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        CONFIG_RESOURCE_PLURAL,
        config_resource_name,
        namespace="default",
    )
    config_resource = k8s.create_custom_resource(config_reference, config)
    config_resource = k8s.wait_resource_consumed_by_controller(config_reference)
    assert config_resource is not None

    yield (config_reference, config_resource)

    # Delete the k8s resource if not already deleted by tests
    try:
        k8s.delete_custom_resource(model_reference)
        k8s.delete_custom_resource(config_reference)
    except:
        pass


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
