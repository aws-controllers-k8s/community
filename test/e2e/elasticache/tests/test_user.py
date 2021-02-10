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

"""CRUD tests for the Elasticache User resource
"""

import boto3
import botocore
import pytest

from time import sleep
from elasticache import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from common.resources import load_resource_file, random_suffix_name
from common import k8s

RESOURCE_PLURAL = "users"
DEFAULT_WAIT_SECS = 90

@pytest.fixture(scope="module")
def elasticache_client():
    return boto3.client("elasticache")

# set up input parameters for User
@pytest.fixture(scope="module")
def input_dict():
    resource_name = random_suffix_name("test-user", 32)
    input_dict = {
        "USER_ID": resource_name,
        "ACCESS_STRING": "on ~app::* -@all +@read"
    }
    return input_dict

@pytest.fixture(scope="module")
def user(input_dict, elasticache_client):

    # inject parameters into yaml; create User in cluster
    user = load_resource_file(
        SERVICE_NAME, "user", additional_replacements=input_dict)
    reference = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, input_dict["USER_ID"], namespace="default")
    _ = k8s.create_custom_resource(reference, user)
    resource = k8s.wait_resource_consumed_by_controller(reference)
    assert resource is not None
    yield (reference, resource)

    # teardown: delete in k8s, assert user does not exist in AWS
    k8s.delete_custom_resource(reference)
    sleep(DEFAULT_WAIT_SECS)
    with pytest.raises(botocore.exceptions.ClientError, match="UserNotFound"):
        _ = elasticache_client.describe_users(UserId=input_dict["USER_ID"])


@service_marker
class TestUser:

    # CRUD test for User; "create" and "delete" operations implicit in "user" fixture
    def test_CRUD(self, user, input_dict):
        (reference, resource) = user
        assert k8s.get_resource_exists(reference)

        resource = k8s.get_resource(reference)
        assert resource["status"]["lastAppliedAccessString"] == input_dict["ACCESS_STRING"]
        assert resource["status"]["status"] == "active"

        new_access_string = "on ~app::* -@all +@read +@write"
        user_patch = {"spec": {"accessString": new_access_string}}
        _ = k8s.patch_custom_resource(reference, user_patch)
        assert k8s.wait_resource_status_desired(reference, "active", wait_periods=3)

        resource = k8s.get_resource(reference)
        assert resource["status"]["lastAppliedAccessString"] == new_access_string
        assert resource["status"]["status"] == "active"