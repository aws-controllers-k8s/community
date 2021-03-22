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
"""Integration tests for the SageMaker DataQualityJobDefinition API.
"""

import boto3
import pytest
import logging
from typing import Dict
import time

from sagemaker import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from sagemaker.replacement_values import REPLACEMENT_VALUES
from sagemaker.tests._fixtures import _make_monitoring_schedule
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'dataqualityjobdefinitions'

def _sagemaker_client():
    return boto3.client('sagemaker')

def _make_job_definition(endpoint_name):
    resource_name = random_suffix_name("data-quality-job-definition", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["JOB_DEFINITION_NAME"] = resource_name
    replacements["ENDPOINT_NAME"] = endpoint_name

    data = load_resource_file(
        SERVICE_NAME, "xgboost_churn_data_quality_job_definition", additional_replacements=replacements
    )
    logging.debug(data)

    reference = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, resource_name, namespace="default"
    )

    return reference, data

@pytest.fixture(scope="module")
def xgboost_churn_data_quality_job_definition(xgboost_churn_endpoint):
    (_, _, endpoint_spec) = xgboost_churn_endpoint

    job_definition_reference, job_definition_data = _make_job_definition()
    resource = k8s.create_custom_resource(job_definition_reference, job_definition_data)
    resource = k8s.wait_resource_consumed_by_controller(job_definition_reference)

    # Create a monitoring schedule to attach the job definition to the endpoint
    endpoint_name = endpoint_spec.get("endpointName")
    assert endpoint_name is not None
    job_definition_name = resource["spec"].get("jobDefinitionName")
    assert job_definition_name is not None

    monitoring_schedule_reference, monitoring_schedule_data = \
        _make_monitoring_schedule("DataQuality", job_definition_name)
    monitoring_schedule_resource = k8s.create_custom_resource(monitoring_schedule_reference, monitoring_schedule_data)
    monitoring_schedule_resource = k8s.wait_resource_consumed_by_controller(monitoring_schedule_reference)

    yield (job_definition_reference, resource, monitoring_schedule_reference, monitoring_schedule_resource)

    for cr in (monitoring_schedule_reference, job_definition_reference):
        if k8s.get_resource_exists(cr):
            k8s.delete_custom_resource(cr)

def get_sagemaker_data_quality_job_definition(job_definition_name: str):
    try:
        hpo_desc = _sagemaker_client().describe_data_quality_job_definition(
            JobDefinitionName=job_definition_name
        )
        return hpo_desc
    except BaseException:
        logging.error(
            f"Could not find Data Quality Job Definition with name {job_definition_name}"
        )
        return None

@service_marker
@pytest.mark.canary
class TestDataQualityJobDefinition:
    def test_create_definition(self, xgboost_churn_data_quality_job_definition):
        (job_definition_reference, resource, monitoring_schedule_reference, monitoring_schedule_resource) = \
            xgboost_churn_data_quality_job_definition
        assert k8s.get_resource_exists(job_definition_reference)
        assert k8s.get_resource_exists(monitoring_schedule_reference)
    
        job_definition_name = resource["spec"].get("jobDefinitionName")
        assert job_definition_name is not None

        description = get_sagemaker_data_quality_job_definition(job_definition_name)
        assert k8s.get_resource_arn(resource) == description["JobDefinitionArn"]

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(job_definition_reference)
        assert deleted is True

        description = get_sagemaker_data_quality_job_definition(job_definition_name)
        assert description