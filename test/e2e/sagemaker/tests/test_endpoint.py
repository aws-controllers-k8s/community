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
    service_marker,
    CONFIG_RESOURCE_PLURAL,
    MODEL_RESOURCE_PLURAL,
    ENDPOINT_RESOURCE_PLURAL,
    create_sagemaker_resource,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.aws import copy_s3_object, delete_s3_object
from common.resources import random_suffix_name
from common import k8s


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def name_suffix():
    return random_suffix_name("xgboost-endpoint", 32)


@pytest.fixture(scope="module")
def single_container_model(name_suffix):
    model_resource_name = name_suffix + "-model"
    replacements = REPLACEMENT_VALUES.copy()
    replacements["MODEL_NAME"] = model_resource_name

    model_reference, model_spec, model_resource = create_sagemaker_resource(
        resource_plural=MODEL_RESOURCE_PLURAL,
        resource_name=model_resource_name,
        spec_file="xgboost_model",
        replacements=replacements,
    )
    assert model_resource is not None

    yield (model_reference, model_resource)

    k8s.delete_custom_resource(model_reference)


@pytest.fixture(scope="module")
def multi_variant_config(name_suffix, single_container_model):
    config_resource_name = name_suffix + "-multi-variant-config"
    (_, model_resource) = single_container_model
    model_resource_name = model_resource["spec"].get("modelName", None)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CONFIG_NAME"] = config_resource_name
    replacements["MODEL_NAME"] = model_resource_name

    config_reference, config_spec, config_resource = create_sagemaker_resource(
        resource_plural=CONFIG_RESOURCE_PLURAL,
        resource_name=config_resource_name,
        spec_file="endpoint_config_multi_variant",
        replacements=replacements,
    )
    assert config_resource is not None

    yield (config_reference, config_resource)

    k8s.delete_custom_resource(config_reference)


@pytest.fixture(scope="module")
def single_variant_config(name_suffix, single_container_model):
    config_resource_name = name_suffix + "-single-variant-config"
    (_, model_resource) = single_container_model
    model_resource_name = model_resource["spec"].get("modelName", None)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CONFIG_NAME"] = config_resource_name
    replacements["MODEL_NAME"] = model_resource_name

    config_reference, config_spec, config_resource = create_sagemaker_resource(
        resource_plural=CONFIG_RESOURCE_PLURAL,
        resource_name=config_resource_name,
        spec_file="endpoint_config_single_variant",
        replacements=replacements,
    )
    assert config_resource is not None

    yield (config_reference, config_resource)

    k8s.delete_custom_resource(config_reference)


@pytest.fixture(scope="module")
def xgboost_endpoint(name_suffix, single_variant_config):
    endpoint_resource_name = name_suffix
    (_, config_resource) = single_variant_config
    config_resource_name = config_resource["spec"].get("endpointConfigName", None)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ENDPOINT_NAME"] = endpoint_resource_name
    replacements["CONFIG_NAME"] = config_resource_name

    reference, spec, resource = create_sagemaker_resource(
        resource_plural=ENDPOINT_RESOURCE_PLURAL,
        resource_name=endpoint_resource_name,
        spec_file="endpoint_base",
        replacements=replacements,
    )

    assert resource is not None

    yield (reference, resource, spec)

    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(reference):
        k8s.delete_custom_resource(reference)


@pytest.fixture(scope="module")
def faulty_config(name_suffix, single_container_model):
    replacements = REPLACEMENT_VALUES.copy()

    # copy model data to a temp S3 location and delete it after model is created on SageMaker
    model_bucket = replacements["SAGEMAKER_DATA_BUCKET"]
    copy_source = {
        "Bucket": model_bucket,
        "Key": "sagemaker/model/xgboost-mnist-model.tar.gz",
    }
    model_destination_key = "sagemaker/model/delete/xgboost-mnist-model.tar.gz"
    copy_s3_object(model_bucket, copy_source, model_destination_key)

    model_resource_name = name_suffix + "faulty-model"
    replacements["MODEL_NAME"] = model_resource_name
    replacements["MODEL_LOCATION"] = f"s3://{model_bucket}/{model_destination_key}"
    model_reference, model_spec, model_resource = create_sagemaker_resource(
        resource_plural=MODEL_RESOURCE_PLURAL,
        resource_name=model_resource_name,
        spec_file="xgboost_model_with_model_location",
        replacements=replacements,
    )
    assert model_resource is not None
    model_resource = k8s.get_resource(model_reference)
    assert (
        "ackResourceMetadata" in model_resource["status"]
        and "arn" in model_resource["status"]["ackResourceMetadata"]
    )
    delete_s3_object(model_bucket, model_destination_key)

    config_resource_name = name_suffix + "-faulty-config"
    (_, model_resource) = single_container_model
    model_resource_name = model_resource["spec"].get("modelName", None)

    replacements["CONFIG_NAME"] = config_resource_name

    config_reference, config_spec, config_resource = create_sagemaker_resource(
        resource_plural=CONFIG_RESOURCE_PLURAL,
        resource_name=config_resource_name,
        spec_file="endpoint_config_multi_variant",
        replacements=replacements,
    )
    assert config_resource is not None

    yield (config_reference, config_resource)

    k8s.delete_custom_resource(model_reference)
    k8s.delete_custom_resource(config_reference)


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
        wait_periods: int = 30,
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
        wait_periods: int = 60,
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
            == self._wait_resource_endpoint_status(reference, expected_status, 2)
            == expected_status
        )

    def create_endpoint_test(self, sagemaker_client, xgboost_endpoint):
        (reference, resource, _) = xgboost_endpoint
        assert k8s.get_resource_exists(reference)

        # endpoint has correct arn and status
        endpoint_name = resource["spec"].get("endpointName", None)
        assert endpoint_name is not None

        assert (
            self._get_resource_endpoint_arn(resource)
            == self._describe_sagemaker_endpoint(sagemaker_client, endpoint_name)[
                "EndpointArn"
            ]
        )

        # endpoint transitions Creating -> InService state
        self._assert_endpoint_status_in_sync(
            sagemaker_client, endpoint_name, reference, self.status_creating
        )
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "False")

        self._assert_endpoint_status_in_sync(
            sagemaker_client, endpoint_name, reference, self.status_inservice
        )
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True")

    def update_endpoint_failed_test(
        self, sagemaker_client, single_variant_config, faulty_config, xgboost_endpoint
    ):
        (endpoint_reference, _, endpoint_spec) = xgboost_endpoint
        (_, faulty_config_resource) = faulty_config
        faulty_config_name = faulty_config_resource["spec"].get(
            "endpointConfigName", None
        )
        endpoint_spec["spec"]["endpointConfigName"] = faulty_config_name
        endpoint_resource = k8s.patch_custom_resource(endpoint_reference, endpoint_spec)
        endpoint_resource = k8s.wait_resource_consumed_by_controller(endpoint_reference)
        assert endpoint_resource is not None

        # endpoint transitions Updating -> InService state
        self._assert_endpoint_status_in_sync(
            sagemaker_client,
            endpoint_reference.name,
            endpoint_reference,
            self.status_udpating,
        )
        assert k8s.wait_on_condition(endpoint_reference, "ACK.ResourceSynced", "False")
        endpoint_resource = k8s.get_resource(endpoint_reference)
        assert (
            endpoint_resource["status"].get("lastEndpointConfigNameForUpdate", None)
            == faulty_config_name
        )

        self._assert_endpoint_status_in_sync(
            sagemaker_client,
            endpoint_reference.name,
            endpoint_reference,
            self.status_inservice,
        )

        assert k8s.wait_on_condition(endpoint_reference, "ACK.ResourceSynced", "True")
        assert k8s.assert_condition_state_message(
            endpoint_reference,
            "ACK.Terminal",
            "True",
            "Unable to update Endpoint. Check FailureReason",
        )

        endpoint_resource = k8s.get_resource(endpoint_reference)
        assert endpoint_resource["status"].get("failureReason", None) is not None

        # additional check: endpoint using old endpoint config
        (_, old_config_resource) = single_variant_config
        current_config_name = endpoint_resource["status"].get(
            "latestEndpointConfigName"
        )
        assert (
            current_config_name is not None
            and current_config_name
            == old_config_resource["spec"].get("endpointConfigName", None)
        )

    def update_endpoint_successful_test(
        self, sagemaker_client, multi_variant_config, xgboost_endpoint
    ):
        (endpoint_reference, endpoint_resource, endpoint_spec) = xgboost_endpoint

        endpoint_name = endpoint_resource["spec"].get("endpointName", None)
        production_variants = self._describe_sagemaker_endpoint(
            sagemaker_client, endpoint_name
        )["ProductionVariants"]
        old_variant_instance_count = production_variants[0]["CurrentInstanceCount"]
        old_variant_name = production_variants[0]["VariantName"]

        (_, new_config_resource) = multi_variant_config
        new_config_name = new_config_resource["spec"].get("endpointConfigName", None)
        endpoint_spec["spec"]["endpointConfigName"] = new_config_name
        endpoint_resource = k8s.patch_custom_resource(endpoint_reference, endpoint_spec)
        endpoint_resource = k8s.wait_resource_consumed_by_controller(endpoint_reference)
        assert endpoint_resource is not None

        # endpoint transitions Updating -> InService state
        self._assert_endpoint_status_in_sync(
            sagemaker_client,
            endpoint_reference.name,
            endpoint_reference,
            self.status_udpating,
        )

        assert k8s.wait_on_condition(endpoint_reference, "ACK.ResourceSynced", "False")
        assert k8s.assert_condition_state_message(
            endpoint_reference, "ACK.Terminal", "False", None
        )
        endpoint_resource = k8s.get_resource(endpoint_reference)
        assert (
            endpoint_resource["status"].get("lastEndpointConfigNameForUpdate", None)
            == new_config_name
        )

        self._assert_endpoint_status_in_sync(
            sagemaker_client,
            endpoint_reference.name,
            endpoint_reference,
            self.status_inservice,
        )
        assert k8s.wait_on_condition(endpoint_reference, "ACK.ResourceSynced", "True")
        assert k8s.assert_condition_state_message(
            endpoint_reference, "ACK.Terminal", "False", None
        )
        endpoint_resource = k8s.get_resource(endpoint_reference)
        assert endpoint_resource["status"].get("failureReason", None) is None

        # RetainAllVariantProperties - variant properties were retained + is a multi-variant endpoint
        new_production_variants = self._describe_sagemaker_endpoint(
            sagemaker_client, endpoint_name
        )["ProductionVariants"]
        assert len(new_production_variants) > 1
        new_variant_instance_count = None
        for variant in new_production_variants:
            if variant["VariantName"] == old_variant_name:
                new_variant_instance_count = variant["CurrentInstanceCount"]

        assert new_variant_instance_count == old_variant_instance_count

    def delete_endpoint_test(self, sagemaker_client, xgboost_endpoint):
        (reference, resource, _) = xgboost_endpoint
        endpoint_name = resource["spec"].get("endpointName", None)

        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        # resource is removed from management from controller side if call to deleteEndpoint succeeds.
        # Sagemaker also removes a 'Deleting' endpoint pretty quickly, but there might be a lag
        # If we see errors in this part of test, can add a loop in future or consider changing controller
        # to wait for SageMaker
        time.sleep(10)
        assert (
            self._describe_sagemaker_endpoint(sagemaker_client, endpoint_name) is None
        )

    def test_driver(
        self,
        sagemaker_client,
        single_variant_config,
        faulty_config,
        multi_variant_config,
        xgboost_endpoint,
    ):
        self.create_endpoint_test(sagemaker_client, xgboost_endpoint)
        self.update_endpoint_failed_test(
            sagemaker_client, single_variant_config, faulty_config, xgboost_endpoint
        )
        # Note: the test has been intentionally ordered to run a successful update after a failed update
        # check that controller updates the endpoint, removes the terminal condition and clears the failure reason
        self.update_endpoint_successful_test(
            sagemaker_client, multi_variant_config, xgboost_endpoint
        )
        self.delete_endpoint_test(sagemaker_client, xgboost_endpoint)
