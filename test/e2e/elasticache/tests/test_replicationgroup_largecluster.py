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

"""Large cluster test for replication group resource
"""

import pytest

from time import sleep
from common import k8s
from elasticache.tests.test_replicationgroup import make_replication_group, rg_deletion_waiter, make_rg_name, DEFAULT_WAIT_SECS
from elasticache.util import provide_node_group_configuration

@pytest.fixture(scope="module")
def rg_largecluster_input(make_rg_name):
    return {
        "RG_ID": make_rg_name("rg-large-cluster"),
        "NUM_NODE_GROUPS": "125",
        "REPLICAS_PER_NODE_GROUP": "3"
    }

@pytest.fixture(scope="module")
def rg_largecluster(rg_largecluster_input, make_replication_group, rg_deletion_waiter):
    input_dict = rg_largecluster_input

    (reference, resource) = make_replication_group("replicationgroup_largecluster", input_dict, input_dict["RG_ID"])
    yield (reference, resource)

    # teardown
    k8s.delete_custom_resource(reference)
    sleep(DEFAULT_WAIT_SECS)
    rg_deletion_waiter.wait(ReplicationGroupId=input_dict["RG_ID"])

class TestReplicationGroupLargeCluster:

    @pytest.mark.slow
    def test_rg_largecluster(self, rg_largecluster_input, rg_largecluster):
        (reference, _) = rg_largecluster
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True", wait_periods=240)

        # assertions after initial creation
        desired_node_groups = int(rg_largecluster_input['NUM_NODE_GROUPS'])
        desired_replica_count = int(rg_largecluster_input['REPLICAS_PER_NODE_GROUP'])
        desired_total_nodes = (desired_node_groups * (1 + desired_replica_count))
        resource = k8s.get_resource(reference)
        assert resource['status']['status'] == "available"
        assert len(resource['status']['nodeGroups']) == desired_node_groups
        assert len(resource['status']['memberClusters']) == desired_total_nodes

        # update, wait for resource to sync
        desired_node_groups = desired_node_groups - 10
        desired_total_nodes = (desired_node_groups * (1 + desired_replica_count))
        patch = {"spec": {"numNodeGroups": desired_node_groups,
                          "nodeGroupConfiguration": provide_node_group_configuration(desired_node_groups)}}
        _ = k8s.patch_custom_resource(reference, patch)
        sleep(DEFAULT_WAIT_SECS) # required as controller has likely not placed the resource in modifying
        assert k8s.wait_on_condition(reference, "ACK.ResourceSynced", "True", wait_periods=240)

        # assert new state after scaling in
        resource = k8s.get_resource(reference)
        assert resource['status']['status'] == "available"
        assert len(resource['status']['nodeGroups']) == desired_node_groups
        assert len(resource['status']['memberClusters']) == desired_total_nodes