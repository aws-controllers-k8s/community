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
"""Integration tests for the SageMaker Endpoint API.
"""

import boto3
import pytest
import logging
import time
from typing import Dict

from sagemaker import (
    SERVICE_NAME,
    service_marker,
    CRD_GROUP,
    CRD_VERSION,
    CONFIG_RESOURCE_PLURAL,
    MODEL_RESOURCE_PLURAL,
    ENDPOINT_RESOURCE_PLURAL,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import load_resource_file, random_suffix_name
from common import k8s


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def single_variant_xgboost_endpoint():
    endpoint_resource_name = random_suffix_name("single-variant-endpoint", 32)
    config1_resource_name = endpoint_resource_name + "-config"
    model_resource_name = config1_resource_name + "-model"

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ENDPOINT_NAME"] = endpoint_resource_name
    replacements["CONFIG_NAME"] = config1_resource_name
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

    endpoint_spec = load_resource_file(
        SERVICE_NAME, "endpoint_base", additional_replacements=replacements
    )
    logging.debug(endpoint_spec)

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

    config1_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        CONFIG_RESOURCE_PLURAL,
        config1_resource_name,
        namespace="default",
    )
    config1_resource = k8s.create_custom_resource(config1_reference, config)
    config1_resource = k8s.wait_resource_consumed_by_controller(config1_reference)
    assert config1_resource is not None

    config2_resource_name = random_suffix_name("2-single-variant-endpoint", 32)
    config["metadata"]["name"] = config["spec"][
        "endpointConfigName"
    ] = config2_resource_name
    logging.debug(config)
    config2_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        CONFIG_RESOURCE_PLURAL,
        config2_resource_name,
        namespace="default",
    )
    config2_resource = k8s.create_custom_resource(config2_reference, config)
    config2_resource = k8s.wait_resource_consumed_by_controller(config2_reference)
    assert config2_resource is not None

    endpoint_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        ENDPOINT_RESOURCE_PLURAL,
        endpoint_resource_name,
        namespace="default",
    )
    endpoint_resource = k8s.create_custom_resource(endpoint_reference, endpoint_spec)
    endpoint_resource = k8s.wait_resource_consumed_by_controller(endpoint_reference)
    assert endpoint_resource is not None

    yield (endpoint_reference, endpoint_resource, endpoint_spec, config2_resource_name)

    # Delete the k8s resource if not already deleted by tests
    for cr in (model_reference, config1_reference, config2_reference, endpoint_reference):
        try:
            k8s.delete_custom_resource(cr)
        except:
            pass


@service_marker
@pytest.mark.canary
class TestEndpoint:
    status_creating: str = "Creating"
    status_inservice: str = "InService"
    status_udpating: str = "Updating"

    def _get_resource_endpoint_arn(self, resource: Dict):
        assert (
            "ackResourceMetadata" in resource["status"]
            and "arn" in resource["status"]["ackResourceMetadata"]
        )
        return resource["status"]["ackResourceMetadata"]["arn"]

    def _describe_sagemaker_endpoint(self, sagemaker_client, endpoint_name: str):
        try:
            return sagemaker_client.describe_endpoint(EndpointName=endpoint_name)
        except BaseException:
            logging.error(
                f"SageMaker could not find a endpoint with the name {endpoint_name}"
            )
            return None

    def _wait_resource_endpoint_status(
        self,
        reference: k8s.CustomResourceReference,
        expected_status: str,
        wait_periods: int = 18,
    ):
        resource_status = None
        for _ in range(wait_periods):
            time.sleep(30)
            resource = k8s.get_resource(reference)
            assert "endpointStatus" in resource["status"]
            resource_status = resource["status"]["endpointStatus"]
            if resource_status == expected_status:
                break
        else:
            logging.error(
                f"Wait for endpoint resource status: {expected_status} timed out. Actual status: {resource_status}"
            )

        return resource_status

    def _wait_sagemaker_endpoint_status(
        self,
        sagemaker_client,
        endpoint_name,
        expected_status: str,
        wait_periods: int = 18,
    ):
        actual_status = None
        for _ in range(wait_periods):
            time.sleep(30)
            actual_status = sagemaker_client.describe_endpoint(
                EndpointName=endpoint_name
            )["EndpointStatus"]
            if actual_status == expected_status:
                break
        else:
            logging.error(
                f"Wait for sagemaker endpoint status: {expected_status} timed out. Actual status: {actual_status}"
            )

        return actual_status

    def _assert_endpoint_status_in_sync(
        self, sagemaker_client, endpoint_name, reference, expected_status
    ):
        assert (
            self._wait_sagemaker_endpoint_status(
                sagemaker_client, endpoint_name, expected_status
            )
            == self._wait_resource_endpoint_status(reference, expected_status)
            == expected_status
        )

    def test_create_endpoint(self, single_variant_xgboost_endpoint):
        assert k8s.get_resource_exists(single_variant_xgboost_endpoint[0])

    def test_endpoint_has_correct_arn_and_status(
        self, sagemaker_client, single_variant_xgboost_endpoint
    ):
        (reference, _, _, _) = single_variant_xgboost_endpoint
        resource = k8s.get_resource(reference)
        endpoint_name = resource["spec"].get("endpointName", None)

        assert endpoint_name is not None

        assert (
            self._get_resource_endpoint_arn(resource)
            == self._describe_sagemaker_endpoint(sagemaker_client, endpoint_name)[
                "EndpointArn"
            ]
        )

        self._assert_endpoint_status_in_sync(
            sagemaker_client, endpoint_name, reference, self.status_creating
        )
        self._assert_endpoint_status_in_sync(
            sagemaker_client, endpoint_name, reference, self.status_inservice
        )

    def test_update_endpoint(self, sagemaker_client, single_variant_xgboost_endpoint):
        (
            reference,
            resource,
            endpoint_spec,
            config2_resource_name,
        ) = single_variant_xgboost_endpoint
        endpoint_spec["spec"]["endpointConfigName"] = config2_resource_name
        resource = k8s.patch_custom_resource(reference, endpoint_spec)
        resource = k8s.wait_resource_consumed_by_controller(reference)
        assert resource is not None

        self._assert_endpoint_status_in_sync(
            sagemaker_client, reference.name, reference, self.status_udpating
        )
        self._assert_endpoint_status_in_sync(
            sagemaker_client, reference.name, reference, self.status_inservice
        )

    def test_delete_endpoint(self, sagemaker_client, single_variant_xgboost_endpoint):
        (reference, _, _, _) = single_variant_xgboost_endpoint
        resource = k8s.get_resource(reference)
        endpoint_name = resource["spec"].get("endpointName", None)

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        assert (
            self._describe_sagemaker_endpoint(sagemaker_client, endpoint_name) is None
        )
