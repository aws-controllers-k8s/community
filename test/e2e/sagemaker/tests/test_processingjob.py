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
"""Integration tests for the SageMaker ProcessingJob API.
"""

import boto3
import pytest
import logging
from typing import Dict

from sagemaker import (
    service_marker,
    create_sagemaker_resource,
)
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import random_suffix_name
from common import k8s

RESOURCE_PLURAL = "processingjobs"


@pytest.fixture(scope="module")
def sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def kmeans_processing_job():
    resource_name = random_suffix_name("kmeans-processingjob", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["PROCESSING_JOB_NAME"] = resource_name

    reference, spec, resource = create_sagemaker_resource(
        resource_plural=RESOURCE_PLURAL,
        resource_name=resource_name,
        spec_file="kmeans_processingjob",
        replacements=replacements,
    )
    assert resource is not None

    yield (reference, resource)

    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(reference):
        k8s.delete_custom_resource(reference)


@service_marker
@pytest.mark.canary
class TestProcessingJob:
    def _get_created_processing_job_status_list(self):
        return ["InProgress", "Completed"]

    def _get_stopped_processing_job_status_list(self):
        return ["Stopped", "Stopping"]

    def _get_sagemaker_processing_job_arn(
        self, sagemaker_client, processing_job_name: str
    ):
        try:
            processing_job = sagemaker_client.describe_processing_job(
                ProcessingJobName=processing_job_name
            )
            return processing_job["ProcessingJobArn"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a processing job with the name {processing_job_name}"
            )
            return None

    def _get_sagemaker_processing_job_status(
        self, sagemaker_client, processing_job_name: str
    ):
        try:
            processing_job = sagemaker_client.describe_processing_job(
                ProcessingJobName=processing_job_name
            )
            return processing_job["ProcessingJobStatus"]
        except BaseException:
            logging.error(
                f"SageMaker could not find a processing job with the name {processing_job_name}"
            )
            return None

    def test_create_processing_job(self, kmeans_processing_job):
        (reference, resource) = kmeans_processing_job
        assert k8s.get_resource_exists(reference)

    def test_processing_job_has_correct_arn(
        self, sagemaker_client, kmeans_processing_job
    ):
        (reference, _) = kmeans_processing_job
        resource = k8s.get_resource(reference)
        processing_job_name = resource["spec"].get("processingJobName", None)

        assert processing_job_name is not None

        resource_processing_job_arn = k8s.get_resource_arn(resource)
        expected_processing_job_arn = self._get_sagemaker_processing_job_arn(
            sagemaker_client, processing_job_name
        )

        assert resource_processing_job_arn == expected_processing_job_arn

    def test_processing_job_has_created_status(
        self, sagemaker_client, kmeans_processing_job
    ):
        (reference, _) = kmeans_processing_job
        resource = k8s.get_resource(reference)
        processing_job_name = resource["spec"].get("processingJobName", None)

        assert processing_job_name is not None

        current_processing_job_status = self._get_sagemaker_processing_job_status(
            sagemaker_client, processing_job_name
        )
        expected_processing_job_status_list = (
            self._get_created_processing_job_status_list()
        )
        assert current_processing_job_status in expected_processing_job_status_list

    def test_processing_job_has_stopped_status(
        self, sagemaker_client, kmeans_processing_job
    ):
        (reference, _) = kmeans_processing_job
        resource = k8s.get_resource(reference)
        processing_job_name = resource["spec"].get("processingJobName", None)

        assert processing_job_name is not None

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        current_processing_job_status = self._get_sagemaker_processing_job_status(
            sagemaker_client, processing_job_name
        )
        expected_processing_job_status_list = (
            self._get_stopped_processing_job_status_list()
        )
        assert current_processing_job_status in expected_processing_job_status_list
