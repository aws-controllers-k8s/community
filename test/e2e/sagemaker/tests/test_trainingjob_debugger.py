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
"""Integration tests for the SageMaker TrainingJob API with the Debugger Feature.
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

RESOURCE_PLURAL = "trainingjobs"


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def xgboost_trainingjob_debugger():
    resource_name = random_suffix_name("xgboost-trainingjob-debugger", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["TRAINING_JOB_NAME"] = resource_name

    trainingjob = load_resource_file(
        SERVICE_NAME, "xgboost_trainingjob_debugger", additional_replacements=replacements
    )
    logging.debug(trainingjob)

    # Create the k8s resource
    reference = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, resource_name, namespace="default"
    )
    resource = k8s.create_custom_resource(reference, trainingjob)
    resource = k8s.wait_resource_consumed_by_controller(reference)

    assert resource is not None

    yield (reference, resource)

    # # Delete the k8s resource if not already deleted by tests
    # try:
    #     k8s.delete_custom_resource(reference)
    # except:
    #     pass


@service_marker
@pytest.mark.canary
class TestTrainingJobDebugger:
    def _get_created_trainingjob_status_list(self):
        return ["InProgress", "Completed"]

    def _get_stopped_trainingjob_status_list(self):
        return ["Stopped", "Stopping"]

    def _get_created_trainingjob_debugger_status_list(self):
        return ["InProgress", "NoIssuesFound"]

    def _get_stopped_trainingjob_debugger_status_list(self):
        return ["Stopped", "Stopping"]

    def _get_resource_trainingjob_arn(self, resource: Dict):
        assert (
            "ackResourceMetadata" in resource["status"]
            and "arn" in resource["status"]["ackResourceMetadata"]
        )
        return resource["status"]["ackResourceMetadata"]["arn"]

    def _get_sagemaker_trainingjob_arn(self, sagemaker_client, trainingjob_name: str):
        try:
            trainingjob = sagemaker_client.describe_training_job(
                TrainingJobName=trainingjob_name
            )
            return trainingjob["TrainingJobArn"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a trainingJob with the name {trainingjob_name}"
            )
            return None

    def _get_sagemaker_trainingjob_status(
        self, sagemaker_client, trainingjob_name: str
    ):
        try:
            trainingjob = sagemaker_client.describe_training_job(
                TrainingJobName=trainingjob_name
            )
            return trainingjob["TrainingJobStatus"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a trainingJob with the name {trainingjob_name}"
            )
            return None

    def _get_sagemaker_trainingjob_debugger_status(
        self, sagemaker_client, trainingjob_name: str
    ):
        try:
            trainingjob = sagemaker_client.describe_training_job(
                TrainingJobName=trainingjob_name
            )
            return trainingjob["DebugRuleEvaluationStatuses"][0]["RuleEvaluationStatus"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a debugger trainingJob with the name {trainingjob_name}"
            )
            return None

    def test_create_trainingjob_debugger(self, xgboost_trainingjob_debugger):
        (reference, resource) = xgboost_trainingjob_debugger
        assert k8s.get_resource_exists(reference)

    def test_trainingjob_debugger_has_correct_arn(self, sagemaker_client, xgboost_trainingjob_debugger):
        (reference, _) = xgboost_trainingjob_debugger
        resource = k8s.get_resource(reference)
        trainingjob_name = resource["spec"].get("trainingJobName", None)

        assert trainingjob_name is not None

        resource_trainingjob_arn = k8s.get_resource_arn(resource)
        expected_trainingjob_arn = self._get_sagemaker_trainingjob_arn(
            sagemaker_client, trainingjob_name
        )

        assert resource_trainingjob_arn == expected_trainingjob_arn

    def test_trainingjob_debugger_has_created_status(
        self, sagemaker_client, xgboost_trainingjob_debugger
    ):
        (reference, _) = xgboost_trainingjob_debugger
        resource = k8s.get_resource(reference)
        trainingjob_name = resource["spec"].get("trainingJobName", None)

        assert trainingjob_name is not None

        current_trainingjob_status = self._get_sagemaker_trainingjob_status(
            sagemaker_client, trainingjob_name
        )
        expected_trainingjob_status_list = self._get_created_trainingjob_status_list()
        assert current_trainingjob_status in expected_trainingjob_status_list

    def test_trainingjob__debugger_has_debugger_status(
        self, sagemaker_client, xgboost_trainingjob_debugger
    ):
        (reference, _) = xgboost_trainingjob_debugger
        resource = k8s.get_resource(reference)
        trainingjob_name = resource["spec"].get("trainingJobName", None)

        assert trainingjob_name is not None

        current_trainingjob_debugger_status = self._get_sagemaker_trainingjob_debugger_status(
            sagemaker_client, trainingjob_name
        )
        expected_trainingjob_debugger_status_list = self._get_created_trainingjob_debugger_status_list()
        assert current_trainingjob_debugger_status in expected_trainingjob_debugger_status_list

    def test_trainingjob_debugger_has_stopped_status(
        self, sagemaker_client, xgboost_trainingjob_debugger
    ):
        (reference, _) = xgboost_trainingjob_debugger
        resource = k8s.get_resource(reference)
        trainingjob_name = resource["spec"].get("trainingJobName", None)

        assert trainingjob_name is not None

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        current_trainingjob_debugger_status = self._get_sagemaker_trainingjob_status(
            sagemaker_client, trainingjob_name
        )
        expected_trainingjob_debugger_status_list = self._get_stopped_trainingjob_debugger_status_list()
        assert current_trainingjob_debugger_status in expected_trainingjob_debugger_status_list
