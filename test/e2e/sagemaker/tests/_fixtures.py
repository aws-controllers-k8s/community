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
"""Common SageMaker test fixtures.
"""

import pytest
import logging

from sagemaker import (
    SERVICE_NAME,
    CRD_GROUP,
    CRD_VERSION,
    CONFIG_RESOURCE_PLURAL,
    MODEL_RESOURCE_PLURAL,
    ENDPOINT_RESOURCE_PLURAL,
    DATA_QUALITY_JOB_DEFINITION_RESOURCE_PLURAL,
    create_sagemaker_resource,
    wait_sagemaker_endpoint_status
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import load_resource_file, random_suffix_name
from common import k8s

def _make_xgboost_churn_endpoint():
    """Creates a SageMaker endpoint with the XGBoost churn single-variant model
    and data capture enabled.
    """
    endpoint_resource_name = random_suffix_name("xgboost-churn", 32)
    config_resource_name = endpoint_resource_name + "-config"
    model_resource_name = config_resource_name + "-model"

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ENDPOINT_NAME"] = endpoint_resource_name
    replacements["CONFIG_NAME"] = config_resource_name
    replacements["MODEL_NAME"] = model_resource_name

    model = load_resource_file(
        SERVICE_NAME, "xgboost_churn_model", additional_replacements=replacements
    )
    logging.debug(model)

    config = load_resource_file(
        SERVICE_NAME,
        "endpoint_config_data_capture_single_variant",
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

    config_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        CONFIG_RESOURCE_PLURAL,
        config_resource_name,
        namespace="default",
    )

    endpoint_reference = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        ENDPOINT_RESOURCE_PLURAL,
        endpoint_resource_name,
        namespace="default",
    )

    return (model_reference, config_reference, endpoint_reference, \
        model, config, endpoint_spec)

@pytest.fixture(scope="module")
def xgboost_churn_endpoint():
    (model_reference, config_reference, endpoint_reference, \
        model, config, endpoint_spec) = _make_xgboost_churn_endpoint()

    model_resource = k8s.create_custom_resource(model_reference, model)
    model_resource = k8s.wait_resource_consumed_by_controller(model_reference)
    assert model_resource is not None

    config_resource = k8s.create_custom_resource(config_reference, config)
    config_resource = k8s.wait_resource_consumed_by_controller(config_reference)
    assert config_resource is not None

    endpoint_resource = k8s.create_custom_resource(endpoint_reference, endpoint_spec)
    endpoint_resource = k8s.wait_resource_consumed_by_controller(endpoint_reference)
    assert endpoint_resource is not None

    yield (endpoint_reference, endpoint_resource, endpoint_spec)

    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(endpoint_reference):
        k8s.delete_custom_resource(endpoint_reference)

@pytest.fixture(scope="module")
def xgboost_churn_data_quality_job_definition(xgboost_churn_endpoint):
    (_, _, endpoint_spec) = xgboost_churn_endpoint

    endpoint_name = endpoint_spec["spec"].get("endpointName")
    assert endpoint_name is not None

    wait_sagemaker_endpoint_status(endpoint_name, "InService")

    resource_name = random_suffix_name("data-quality-job-definition", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["JOB_DEFINITION_NAME"] = resource_name
    replacements["ENDPOINT_NAME"] = endpoint_name

    job_definition_reference, job_definition_data, resource = create_sagemaker_resource(
        resource_plural=DATA_QUALITY_JOB_DEFINITION_RESOURCE_PLURAL,
        resource_name=resource_name,
        spec_file="xgboost_churn_data_quality_job_definition",
        replacements=replacements,
    )
    assert resource is not None

    job_definition_name = resource["spec"].get("jobDefinitionName")
    assert job_definition_name is not None

    yield (job_definition_reference, resource)

    if k8s.get_resource_exists(job_definition_reference):
        k8s.delete_custom_resource(job_definition_reference)
