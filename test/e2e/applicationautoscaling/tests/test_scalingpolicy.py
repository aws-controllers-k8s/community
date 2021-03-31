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
"""Integration tests for the Application Auto Scaling ScalingPolicy API.
"""

import boto3
import pytest
import logging
from typing import Dict, Tuple
import time

from applicationautoscaling import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from applicationautoscaling.replacement_values import REPLACEMENT_VALUES
from applicationautoscaling.bootstrap_resources import TestBootstrapResources, get_bootstrap_resources
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = "scalingpolicies"


@pytest.fixture(scope="module")
def applicationautoscaling_client():
    return boto3.client("application-autoscaling")

@service_marker
@pytest.mark.canary
class TestScalingPolicy:
    def _generate_dynamodb_policy(self, bootstrap_resources: TestBootstrapResources) -> Tuple[k8s.CustomResourceReference, Dict]:
        resource_name = random_suffix_name("dynamodb-scaling-policy", 32)

        replacements = REPLACEMENT_VALUES.copy()
        replacements["SCALINGPOLICY_NAME"] = resource_name
        replacements["DYNAMODB_TABLE"] = bootstrap_resources.RegisteredDynamoTableName

        policy = load_resource_file(
            SERVICE_NAME, "dynamodb_scalingpolicy", additional_replacements=replacements
        )
        logging.debug(policy)

        # Create the k8s resource
        reference = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, resource_name, namespace="default"
        )

        return (reference, policy)
    
    def _get_dynamodb_scaling_policy_exists(self, applicationautoscaling_client, policy_name: str) -> bool:
        targets = applicationautoscaling_client.describe_scaling_policies(
            ServiceNamespace="dynamodb",
            PolicyNames=[policy_name]
        )
        
        return len(targets["ScalingPolicies"]) == 1

    def test_smoke(self, applicationautoscaling_client):
        (reference, policy) = self._generate_dynamodb_policy(get_bootstrap_resources())
        resource = k8s.create_custom_resource(reference, policy)
        resource = k8s.wait_resource_consumed_by_controller(reference)
        assert k8s.get_resource_exists(reference)

        policyName = policy["spec"].get("policyName")
        assert policyName is not None

        exists = self._get_dynamodb_scaling_policy_exists(applicationautoscaling_client, policyName)
        assert exists

        _, deleted = k8s.delete_custom_resource(reference)
        assert deleted is True

        exists = self._get_dynamodb_scaling_policy_exists(applicationautoscaling_client, policyName)
        assert not exists

