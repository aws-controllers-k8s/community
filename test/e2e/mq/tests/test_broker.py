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

"""Integration tests for the AmazonMQ API Broker resource
"""

import boto3
import datetime
import logging
import time
from typing import Dict

import pytest

from mq import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from mq.bootstrap_resources import get_bootstrap_resources
from mq.replacement_values import REPLACEMENT_VALUES
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'brokers'

DELETE_WAIT_AFTER_SECONDS = 20
CREATE_INTERVAL_SLEEP_SECONDS = 15
# Time to wait before we get to an expected RUNNING state.
# In my experience, it regularly takes more than 6 minutes to create a
# single-instance RabbitMQ broker...
CREATE_TIMEOUT_SECONDS = 600


@pytest.fixture(scope="module")
def amq_client():
    return boto3.client('mq')


# TODO(jaypipes): Move to k8s common library
def get_resource_arn(self, resource: Dict):
    assert 'ackResourceMetadata' in resource['status'] and \
        'arn' in resource['status']['ackResourceMetadata']
    return resource['status']['ackResourceMetadata']['arn']


@service_marker
@pytest.mark.canary
class TestRabbitMQBroker:
    def test_create_delete_non_public(self, amq_client):
        resource_name = "my-rabbit-broker-non-public"

        replacements = REPLACEMENT_VALUES.copy()
        replacements["BROKER_NAME"] = resource_name

        resource_data = load_resource_file(
            SERVICE_NAME,
            "broker_rabbitmq_non_public",
            additional_replacements=replacements,
        )
        logging.error(resource_data)

        # Create the k8s resource
        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
            resource_name, namespace="default",
        )
        k8s.create_custom_resource(ref, resource_data)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        broker_id = cr['status']['brokerID']

        # Let's check that the Broker appears in AmazonMQ
        aws_res = amq_client.describe_broker(BrokerId=broker_id)
        assert aws_res is not None

        now = datetime.datetime.now()
        timeout = now + datetime.timedelta(seconds=CREATE_TIMEOUT_SECONDS)

        # TODO(jaypipes): Move this into generic AWS-side waiter
        while aws_res['BrokerState'] != "RUNNING":
            if datetime.datetime.now() >= timeout:
                raise Exception("failed to find running Broker before timeout")
            time.sleep(CREATE_INTERVAL_SLEEP_SECONDS)
            aws_res = amq_client.describe_broker(BrokerId=broker_id)
            assert aws_res is not None

        # Delete the k8s resource on teardown of the module
        k8s.delete_custom_resource(ref)

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Broker should no longer appear in AmazonMQ
        res_found = False
        try:
            amq_client.describe_broker(BrokerId=broker_id)
            res_found = True
        except amq_client.exceptions.NotFoundException:
            pass

        assert res_found is False
