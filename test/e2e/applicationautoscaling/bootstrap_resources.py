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
"""Declares the structure of the bootstrapped resources and provides a loader
for them.
"""

from applicationautoscaling import SERVICE_NAME
from common.resources import read_bootstrap_config
from dataclasses import dataclass

@dataclass
class TestBootstrapResources:
    ScalableDynamoTableName: str # To be used for testing RegisterScalableTarget
    RegisteredDynamoTableName: str # Already registered, for testing scaling policies

_bootstrap_resources = None

def get_bootstrap_resources():
    global _bootstrap_resources
    if _bootstrap_resources is None:
        _bootstrap_resources = TestBootstrapResources(**read_bootstrap_config(SERVICE_NAME))
    return _bootstrap_resources