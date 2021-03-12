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
"""Bootstraps the resources required to run the Application Auto Scaling
integration tests.
"""

import boto3
import json
import logging

from applicationautoscaling.bootstrap_resources import TestBootstrapResources


def service_bootstrap() -> dict:
    logging.getLogger().setLevel(logging.INFO)

    return TestBootstrapResources(
        
    ).__dict__
