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
"""Integration tests for the SageMaker HyperParameterTuning API.
"""

import boto3
import pytest
import logging
from typing import Dict
import time

from sagemaker import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'hyperparametertuningjobs'


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client('sagemaker')


@pytest.fixture(scope="module")
def xgboost_hpojob():
    resource_name = random_suffix_name("xgboost-hpojob", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["HPO_JOB_NAME"] = resource_name

    hpojob = load_resource_file(
        SERVICE_NAME, "xgboost_hpojob", additional_replacements=replacements)
    logging.debug(hpojob)

    # Create the k8s resource
    reference = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, resource_name, namespace="default")
    resource = k8s.create_custom_resource(reference, hpojob)
    resource = k8s.wait_resource_consumed_by_controller(reference)

    assert resource is not None

    yield (reference, resource)    

    # Delete the k8s resource if not already deleted by tests
    try:
        k8s.delete_custom_resource(reference)
    except:
        pass


@service_marker
@pytest.mark.canary
class TestHPO:
    def _get_created_hpo_job_status_list(self):
        return ["InProgress", "Completed"]

    def _get_stopped_hpo_job_status_list(self):
        return ["Stopped", "Stopping"]

    def _get_sagemaker_hpo_job_arn(self, sagemaker_client, hpo_job_name: str):
        try:
            hpo_desc = sagemaker_client.describe_hyper_parameter_tuning_job(
                HyperParameterTuningJobName=hpo_job_name
            )
            return hpo_desc["HyperParameterTuningJobArn"]
        except BaseException:
            logging.error(
                f"SageMaker could not find an hpo job with the name {hpo_job_name}"
            )
            return None

    def _get_sagemaker_hpo_job_status(
        self, sagemaker_client, hpo_job_name: str
    ):
        try:
            hpo_job = sagemaker_client.describe_hyper_parameter_tuning_job(
                HyperParameterTuningJobName=hpo_job_name
            )
            return hpo_job["HyperParameterTuningJobStatus"]
        except BaseException:
            logging.error(
                f"SageMaker could not find an hpo job with the name {hpo_job_name}"
            )
            return None

    def test_create_hpo(self, xgboost_hpojob):
        (reference, resource) = xgboost_hpojob
        assert k8s.get_resource_exists(reference)

    def test_hpo_has_correct_arn(self, sagemaker_client, xgboost_hpojob):
        (reference, _) = xgboost_hpojob
        resource = k8s.get_resource(reference)
        hpo_job_name = resource["spec"].get("hyperParameterTuningJobName", None)

        assert hpo_job_name is not None

        assert k8s.get_resource_arn(resource) == self._get_sagemaker_hpo_job_arn(
            sagemaker_client, hpo_job_name
        )

    def test_hpo_job_has_created_status(
        self, sagemaker_client, xgboost_hpojob
    ):
        (reference, _) = xgboost_hpojob
        resource = k8s.get_resource(reference)
        hpo_job_name = resource["spec"].get("hyperParameterTuningJobName", None)

        assert hpo_job_name is not None

        current_hpo_job_status = self._get_sagemaker_hpo_job_status(
            sagemaker_client, hpo_job_name
        )
        expected_hpo_job_status_list = (
            self._get_created_hpo_job_status_list()
        )
        assert current_hpo_job_status in expected_hpo_job_status_list

    def test_hpo_job_has_stopped_status(
        self, sagemaker_client, xgboost_hpojob
    ):
        (reference, _) = xgboost_hpojob
        resource = k8s.get_resource(reference)
        hpo_job_name = resource["spec"].get("hyperParameterTuningJobName", None)

        assert hpo_job_name is not None

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        current_hpo_job_status = self._get_sagemaker_hpo_job_status(
            sagemaker_client, hpo_job_name
        )
        expected_hpo_job_status_list = (
            self._get_stopped_hpo_job_status_list()
        )
        assert current_hpo_job_status in expected_hpo_job_status_list


