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

from sagemaker import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from sagemaker.replacement_values import REPLACEMENT_VALUES
from sagemaker.tests._fixtures import xgboost_churn_data_quality_job_definition, xgboost_churn_endpoint
from sagemaker.tests._helpers import _wait_sagemaker_endpoint_status, _get_job_definition_arn, _sagemaker_client
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'dataqualityjobdefinitions'

# Access variable so it is loaded as a fixture
_accessed = xgboost_churn_data_quality_job_definition, xgboost_churn_endpoint

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
        (job_definition_reference, resource) = xgboost_churn_data_quality_job_definition
        assert k8s.get_resource_exists(job_definition_reference)
    
        job_definition_name = resource["spec"].get("jobDefinitionName")
        assert job_definition_name is not None

        description = get_sagemaker_data_quality_job_definition(job_definition_name)
        assert _get_job_definition_arn(resource) == description["JobDefinitionArn"]

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(job_definition_reference)
        assert deleted is True

        description = get_sagemaker_data_quality_job_definition(job_definition_name)
        assert description is None