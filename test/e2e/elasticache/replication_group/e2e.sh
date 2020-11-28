#!/usr/bin/env bash

##############################################
# Tests for AWS ElastiCache Replication Group
##############################################

set -u

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/elasticache.sh"

check_is_installed jq "Please install jq before running this script."

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"
debug_msg "selected AWS region: $AWS_REGION"

setup_replication_group_fields() {
  # uses non local variable for later use in tests
  # cluster mode enabled replication group
  rg_id="ack-test-rg-1"
  rg_description="$rg_id description"
  num_node_groups="2"
  replicas_per_node_group="1"
}
setup_replication_group_fields

ack_apply_replication_group_yaml() {
  rg_yaml="$(provide_replication_group_yaml)"
  echo "$rg_yaml" | kubectl apply -f -
}

ack_apply_replication_group_with_node_groups_yaml() {
  rg_yaml="$(provide_replication_group_detailed_yaml)"  # helps determine node groups to retain during decrease
  echo "$rg_yaml" | kubectl apply -f -
}

k8s_controller_reload_credentials "$service_name"

#################################################
# create replication group
#################################################
ack_create_replication_group() {
  setup_replication_group_fields
  debug_msg "Testing create replication group: $rg_id."
  ack_apply_replication_group_yaml
}

ack_create_replication_group
wait_and_assert_replication_group_available_status

#################################################
# modify replication group
#################################################
k8s_assert_replication_group_status_property "$rg_id" ".description" "$rg_description"
ack_modify_replication_group() {
  # uses non local variable for later use in tests
  rg_description="$rg_id description updated"
  debug_msg "Testing modify replication group: $rg_id."
  ack_apply_replication_group_yaml
}
ack_modify_replication_group
wait_and_assert_replication_group_available_status
k8s_assert_replication_group_status_property "$rg_id" ".description" "$rg_description"

#################################################
# modify replication group shards count
#################################################
test_update_shards_count_increase() {
  k8s_assert_replication_group_shard_count "$rg_id" "$num_node_groups"  # assert current value
  # uses non local variable for later use in tests
  num_node_groups="3" # increases from 2 to 3
  debug_msg "Testing modify replication group: $rg_id shards count to new value: $num_node_groups."
  ack_apply_replication_group_yaml
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" "$num_node_groups"  # assert updated value
}
test_update_shards_count_increase

test_update_shards_count_decrease() {
  k8s_assert_replication_group_shard_count "$rg_id" "$num_node_groups"  # assert current value
  # uses non local variable for later use in tests
  num_node_groups="2" # decreases from 3 to 2
  debug_msg "Testing modify replication group: $rg_id shards count to new value: $num_node_groups."
  ack_apply_replication_group_with_node_groups_yaml
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" "$num_node_groups"  # assert updated value
}
test_update_shards_count_decrease

#################################################
# modify replication group replica count
#################################################
ack_modify_replication_group_replica_count() {
  # uses non local variable for later use in tests
  replicas_per_node_group="$1"
  debug_msg "Testing modify replication group: $rg_id replica count to new value: $replicas_per_node_group."
  ack_apply_replication_group_yaml
}
test_update_replica_count() {
  k8s_assert_replication_group_replica_count "$rg_id" "$replicas_per_node_group"  # assert current value
  ack_modify_replication_group_replica_count "$1"
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_replica_count "$rg_id" "$replicas_per_node_group"  # assert updated value
}
### increase replicas count
test_update_replica_count "2"
### decrease replicas count
test_update_replica_count "1"

#################################################
# delete replication group
#################################################
debug_msg "Testing delete replication group: $rg_id."
kubectl delete ReplicationGroup/"$rg_id" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1
sleep 5
aws_wait_replication_group_deleted  "$rg_id" "FAIL: expected replication group $rg_id to have been deleted in ${service_name}"

