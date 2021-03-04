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

"""Stores the values used by each of the integration tests for replacing the
AmazonMQ-specific test variables.
"""

from common.aws import get_aws_region
from mq.bootstrap_resources import get_bootstrap_resources

REPLACEMENT_VALUES = {
    "MQ_RABBITMQ_ENGINE_VERSION": "3.8.6",
    "MQ_HOST_INSTANCE_TYPE": "mq.t3.micro",
}
