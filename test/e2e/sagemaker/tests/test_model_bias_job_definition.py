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
"""Integration tests for the SageMaker ModelBiasJobDefinition API.
"""

import boto3
import pytest
import logging

from sagemaker import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from sagemaker.replacement_values import REPLACEMENT_VALUES
from sagemaker.tests._fixtures import xgboost_churn_endpoint
from sagemaker.tests._helpers import _wait_sagemaker_endpoint_status, _get_job_definition_arn, _sagemaker_client
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'modelbiasjobdefinitions'

# Access variable so it is loaded as a fixture
_accessed = xgboost_churn_endpoint

def _make_job_definition(endpoint_name):
    resource_name = random_suffix_name("model-bias-job-definition", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["JOB_DEFINITION_NAME"] = resource_name
    replacements["ENDPOINT_NAME"] = endpoint_name

    data = load_resource_file(
        SERVICE_NAME, "xgboost_churn_model_bias_job_definition", additional_replacements=replacements
    )
    logging.debug(data)

    reference = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, resource_name, namespace="default"
    )

    return reference, data

@pytest.fixture(scope="module")
def xgboost_churn_model_bias_job_definition(xgboost_churn_endpoint):
    (_, _, endpoint_spec) = xgboost_churn_endpoint

    endpoint_name = endpoint_spec["spec"].get("endpointName")
    assert endpoint_name is not None

    _wait_sagemaker_endpoint_status(endpoint_name, "InService")

    job_definition_reference, job_definition_data = _make_job_definition(endpoint_name)
    resource = k8s.create_custom_resource(job_definition_reference, job_definition_data)
    resource = k8s.wait_resource_consumed_by_controller(job_definition_reference)

    job_definition_name = resource["spec"].get("jobDefinitionName")
    assert job_definition_name is not None

    yield (job_definition_reference, resource)

    if k8s.get_resource_exists(job_definition_reference):
        k8s.delete_custom_resource(job_definition_reference)

def get_sagemaker_model_bias_job_definition(job_definition_name: str):
    try:
        hpo_desc = _sagemaker_client().describe_model_bias_job_definition(
            JobDefinitionName=job_definition_name
        )
        return hpo_desc
    except BaseException:
        logging.error(
            f"Could not find Model Bias Job Definition with name {job_definition_name}"
        )
        return None

@service_marker
@pytest.mark.canary
class TestModelBiasJobDefinition:
    def test_create_definition(self, xgboost_churn_model_bias_job_definition):
        (job_definition_reference, resource) = xgboost_churn_model_bias_job_definition
        assert k8s.get_resource_exists(job_definition_reference)
    
        job_definition_name = resource["spec"].get("jobDefinitionName")
        assert job_definition_name is not None

        description = get_sagemaker_model_bias_job_definition(job_definition_name)
        assert _get_job_definition_arn(resource) == description["JobDefinitionArn"]

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(job_definition_reference)
        assert deleted is True

        description = get_sagemaker_model_bias_job_definition(job_definition_name)
        assert description is None