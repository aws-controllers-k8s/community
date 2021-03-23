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

"""Integration tests for the RDS API DBSecurityGroup resource
"""

import boto3
import datetime
import logging
import time
from typing import Dict

import pytest

from rds import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from rds.bootstrap_resources import get_bootstrap_resources
from rds.replacement_values import REPLACEMENT_VALUES
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = 'dbsecuritygroups'

DELETE_WAIT_AFTER_SECONDS = 10


@pytest.fixture(scope="module")
def rds_client():
    return boto3.client('rds')


@service_marker
@pytest.mark.canary
class TestDBSecurityGroup:
    def test_create_delete_simple(self, rds_client):
        resource_name = "my-db-security-group"
        resource_desc = "my-db-security-group description"

        br_resources = get_bootstrap_resources()

        replacements = REPLACEMENT_VALUES.copy()
        replacements["DB_SECURITY_GROUP_NAME"] = resource_name
        replacements["DB_SECURITY_GROUP_DESC"] = resource_desc

        resource_data = load_resource_file(
            SERVICE_NAME,
            "db_security_group_simple",
            additional_replacements=replacements,
        )
        logging.debug(resource_data)

        # Create the k8s resource
        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
            resource_name, namespace="default",
        )
        k8s.create_custom_resource(ref, resource_data)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        # Let's check that the DB security group appears in RDS
        aws_res = rds_client.describe_db_security_groups(DBSecurityGroupName=resource_name)
        assert aws_res is not None
        assert len(aws_res['DBSecurityGroups']) == 1

        # Delete the k8s resource on teardown of the module
        k8s.delete_custom_resource(ref)

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # DB security group should no longer appear in RDS
        try:
            aws_res = rds_client.describe_db_security_groups(DBSecurityGroupName=resource_name)
            assert False
        except rds_client.exceptions.DBSecurityGroupNotFoundFault:
            pass

