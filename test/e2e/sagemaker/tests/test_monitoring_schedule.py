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
"""Integration tests for the SageMaker MonitoringSchedule API.
"""

import time
import pytest
import logging

from sagemaker import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from sagemaker.replacement_values import REPLACEMENT_VALUES
from sagemaker.tests._fixtures import xgboost_churn_data_quality_job_definition, xgboost_churn_endpoint
from sagemaker.tests._helpers import _sagemaker_client
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'monitoringschedules'

# Access variable so it is loaded as a fixture
_accessed = xgboost_churn_data_quality_job_definition, xgboost_churn_endpoint

def _make_monitoring_schedule(monitoring_type, job_definition_name):
    resource_name = random_suffix_name("monitoring-schedule", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["SCHEDULE_NAME"] = resource_name
    replacements["JOB_DEFINITION_NAME"] = job_definition_name
    replacements["MONITORING_TYPE"] = monitoring_type

    data = load_resource_file(
        SERVICE_NAME, "monitoring_schedule_base", additional_replacements=replacements
    )
    logging.debug(data)

    reference = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, resource_name, namespace="default"
    )

    return reference, data

@pytest.fixture(scope="module")
def xgboost_churn_data_quality_monitoring_schedule(xgboost_churn_data_quality_job_definition):
    (_, job_definition_resource) = xgboost_churn_data_quality_job_definition

    job_definition_name = job_definition_resource["spec"].get("jobDefinitionName")
    assert job_definition_name is not None

    reference, data = _make_monitoring_schedule("DataQuality", job_definition_name)
    resource = k8s.create_custom_resource(reference, data)
    resource = k8s.wait_resource_consumed_by_controller(reference)

    yield (reference, resource)

    if k8s.get_resource_exists(reference):
        k8s.delete_custom_resource(reference)

def get_sagemaker_monitoring_schedule(monitoring_schedule_name: str):
    try:
        hpo_desc = _sagemaker_client().describe_monitoring_schedule(
            MonitoringScheduleName=monitoring_schedule_name
        )
        return hpo_desc
    except BaseException:
        logging.error(
            f"Could not find Monitoring Schedule with name {monitoring_schedule_name}"
        )
        return None

@service_marker
@pytest.mark.canary
class TestMonitoringSchedule:
    def test_create_definition(self, xgboost_churn_data_quality_monitoring_schedule):
        (reference, resource) = xgboost_churn_data_quality_monitoring_schedule
        assert k8s.get_resource_exists(reference)
    
        monitoring_schedule_name = resource["spec"].get("monitoringScheduleName")
        assert monitoring_schedule_name is not None

        description = get_sagemaker_monitoring_schedule(monitoring_schedule_name)
        assert k8s.get_resource_arn(resource) == description["MonitoringScheduleArn"]

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        # Arbitrary wait for server-side acknowledgement
        time.sleep(10)

        description = get_sagemaker_monitoring_schedule(monitoring_schedule_name)
        assert description is None