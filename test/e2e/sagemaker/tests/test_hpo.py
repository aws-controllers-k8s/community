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
"""Integration tests for the SageMaker HyperParameterTuning API.
"""

import boto3
import pytest
import logging
from typing import Dict

from sagemaker import service_marker, create_sagemaker_resource
from sagemaker.replacement_values import REPLACEMENT_VALUES
from common.resources import random_suffix_name
from common import k8s

RESOURCE_PLURAL = "hyperparametertuningjobs"
HPO_JOB_STATUS_CREATED = ("InProgress", "Completed")
HPO_JOB_STATUS_STOPPED = ("Stopped", "Stopping")


def _sagemaker_client():
    return boto3.client("sagemaker")


@pytest.fixture(scope="module")
def xgboost_hpojob():
    resource_name = random_suffix_name("xgboost-hpojob", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["HPO_JOB_NAME"] = resource_name

    reference, spec, resource = create_sagemaker_resource(
        resource_plural=RESOURCE_PLURAL,
        resource_name=resource_name,
        spec_file="xgboost_hpojob",
        replacements=replacements,
    )
    assert resource is not None
    yield (reference, resource)

    # Delete the k8s resource if not already deleted by tests
    if k8s.get_resource_exists(reference):
        k8s.delete_custom_resource(reference)


def get_sagemaker_hpo_job(hpo_job_name: str):
    try:
        hpo_desc = _sagemaker_client().describe_hyper_parameter_tuning_job(
            HyperParameterTuningJobName=hpo_job_name
        )
        return hpo_desc
    except BaseException:
        logging.error(
            f"SageMaker could not find an hpo job with the name {hpo_job_name}"
        )
        return None


@service_marker
@pytest.mark.canary
class TestHPO:
    def test_create_hpo(self, xgboost_hpojob):
        (reference, resource) = xgboost_hpojob
        assert k8s.get_resource_exists(reference)

        hpo_job_name = resource["spec"].get("hyperParameterTuningJobName", None)
        assert hpo_job_name is not None

        hpo_sm_desc = get_sagemaker_hpo_job(hpo_job_name)
        assert (
            k8s.get_resource_arn(resource) == hpo_sm_desc["HyperParameterTuningJobArn"]
        )
        assert hpo_sm_desc["HyperParameterTuningJobStatus"] in HPO_JOB_STATUS_CREATED

        # Delete the k8s resource.
        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        hpo_sm_desc = get_sagemaker_hpo_job(hpo_job_name)
        assert hpo_sm_desc["HyperParameterTuningJobStatus"] in HPO_JOB_STATUS_STOPPED
