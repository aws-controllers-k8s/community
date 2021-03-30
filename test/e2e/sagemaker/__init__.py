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

import pytest
import logging
from common import k8s

SERVICE_NAME = "sagemaker"
CRD_GROUP = "sagemaker.services.k8s.aws"
CRD_VERSION = "v1alpha1"

CONFIG_RESOURCE_PLURAL = "endpointconfigs"
MODEL_RESOURCE_PLURAL = "models"
ENDPOINT_RESOURCE_PLURAL = "endpoints"

# PyTest marker for the current service
service_marker = pytest.mark.service(arg=SERVICE_NAME)


def create_sagemaker_resource(
    resource_plural, resource_name, spec_file, replacements, namespace="default"
):
    """
    Wrapper around k8s.load_and_create_resource to create a SageMaker resource
    """

    reference, spec, resource = k8s.load_and_create_resource(
        SERVICE_NAME,
        CRD_GROUP,
        CRD_VERSION,
        resource_plural,
        resource_name,
        spec_file,
        replacements,
        namespace,
    )

    return reference, spec, resource
