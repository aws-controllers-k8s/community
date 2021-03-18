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

"""Integration tests for the Elasticache ReplicationGroup resource
"""

import pytest
import boto3
import logging

from common.resources import read_bootstrap_config, random_suffix_name, load_resource_file
from common import k8s
from time import sleep
from elasticache import SERVICE_NAME, service_marker, CRD_GROUP, CRD_VERSION
from elasticache.service_bootstrap import BootstrapResources

RESOURCE_PLURAL = "replicationgroups"
DEFAULT_WAIT_SECS = 30

@pytest.fixture(scope="module")
def rg_deletion_waiter():
    ec = boto3.client("elasticache")
    return ec.get_waiter('replication_group_deleted')

# retrieve resources created in the bootstrap step
@pytest.fixture(scope="module")
def bootstrap_resources():
    return BootstrapResources(**read_bootstrap_config(SERVICE_NAME))

# factory for replication group names
@pytest.fixture(scope="module")
def make_rg_name():
    def _make_rg_name(base):
        return random_suffix_name(base, 32)

    return _make_rg_name

# factory for replication groups
@pytest.fixture(scope="module")
def make_replication_group():
    def _make_replication_group(yaml_name, input_dict, rg_name):
        rg = load_resource_file(
            SERVICE_NAME, yaml_name, additional_replacements=input_dict)
        logging.debug(rg)

        reference = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL, rg_name, namespace="default")
        _ = k8s.create_custom_resource(reference, rg)
        resource = k8s.wait_resource_consumed_by_controller(reference, wait_periods=10)
        assert resource is not None
        return (reference, resource)

    return _make_replication_group

@pytest.fixture(scope="module")
def rg_input_coverage(bootstrap_resources, make_rg_name, make_replication_group, rg_deletion_waiter):
    input_dict = {
        "RG_ID": make_rg_name("rg-input-coverage"),
        "KMS_KEY_ID": bootstrap_resources.KmsKeyID,
        "SNS_TOPIC_ARN": bootstrap_resources.SnsTopicARN,
        "SG_ID": bootstrap_resources.SecurityGroupID,
        "USERGROUP_ID": bootstrap_resources.UserGroupID
    }

    (reference, resource) = make_replication_group("replicationgroup_input_coverage", input_dict, input_dict["RG_ID"])
    yield (reference, resource)

    # teardown
    k8s.delete_custom_resource(reference)
    sleep(DEFAULT_WAIT_SECS)
    rg_deletion_waiter.wait(ReplicationGroupId=input_dict["RG_ID"]) #throws exception if wait fails

@pytest.fixture(scope="module")
def first_secret():
    k8s.create_opaque_secret("default", "first", "secret1", "securetoken123456")
    yield
    k8s.delete_secret("default", "first")

@pytest.fixture(scope="module")
def second_secret():
    k8s.create_opaque_secret("default", "second", "secret2", "newsecuretoken123456")
    yield
    k8s.delete_secret("default", "second")


@pytest.fixture(scope="module")
def rg_auth_token(make_rg_name, make_replication_group, rg_deletion_waiter, first_secret, second_secret):
    input_dict = {
        "RG_ID": make_rg_name("rg-auth-token"),
        "NAME": "first",
        "KEY": "secret1"
    }
    (reference, resource) = make_replication_group("replicationgroup_authtoken", input_dict, input_dict["RG_ID"])
    yield (reference, resource)
    k8s.delete_custom_resource(reference)
    sleep(DEFAULT_WAIT_SECS)
    rg_deletion_waiter.wait(ReplicationGroupId=input_dict["RG_ID"]) #throws exception if wait fails

@pytest.fixture(scope="module")
def rg_cmd_fromsnapshot(bootstrap_resources, make_rg_name, make_replication_group, rg_deletion_waiter):
    input_dict = {
        "RG_ID": make_rg_name("rg-cmd-fromsnapshot"),
        "SNAPSHOT_NAME": bootstrap_resources.SnapshotName
    }

    (reference, resource) = make_replication_group("replicationgroup_cmd_fromsnapshot", input_dict, input_dict["RG_ID"])
    yield (reference, resource)

    # teardown
    k8s.delete_custom_resource(reference)
    sleep(DEFAULT_WAIT_SECS)
    rg_deletion_waiter.wait(ReplicationGroupId=input_dict["RG_ID"])


@service_marker
class TestReplicationGroup:

    def test_rg_input_coverage(self, rg_input_coverage):
        (reference, _) = rg_input_coverage
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True", wait_periods=30)

    def test_rg_cmd_fromsnapshot(self, rg_cmd_fromsnapshot):
        (reference, _) = rg_cmd_fromsnapshot
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True", wait_periods=30)

    def test_rg_auth_token(self, rg_auth_token):
        (reference, _) = rg_auth_token
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True", wait_periods=30)

        update_dict = {
            "RG_ID": reference.name,
            "NAME": "second",
            "KEY": "secret2"
        }

        updated_spec = load_resource_file(
            SERVICE_NAME, "replicationgroup_authtoken", additional_replacements=update_dict)

        k8s.patch_custom_resource(reference, updated_spec)
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True", wait_periods=30)
