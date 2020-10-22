#!/usr/bin/env bash

# replication group modification tests

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/elasticache.sh"

check_is_installed jq "Please install jq before running this script."

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test group: $service_name/$test_name------------------------------"

# modify replication group to enable auto failover: expect success
test_modify_rg_enable_auto_failover() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-enable-failover"
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  num_node_groups=1
  replicas_per_node_group=1
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check property as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "disabled"

  # update configuration and apply
  automatic_failover_enabled="true"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait until RG available again then check value updated
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "enabled"
}

# create 1 shard/no replica RG, enable autofailover while adding replicas
test_modify_rg_enable_af_add_replicas() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-enable-af-add-replicas"
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 0
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "disabled"
  k8s_assert_replication_group_status_property "$rg_id" ".multiAZ" "disabled"

  # update config and apply: enable autofailover and add replicas to satisfy enabling condition
  replicas_per_node_group=3
  automatic_failover_enabled="true"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new state
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 3
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "enabled"
  k8s_assert_replication_group_status_property "$rg_id" ".multiAZ" "disabled"
}

# create a cluster mode disabled RG with 3 replicas, and scale up
test_modify_rg_cmd_scale_up() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-cmd-scale-up"
  automatic_failover_enabled="true"
  cache_node_type="cache.t3.micro"
  num_node_groups=1
  replicas_per_node_group=3
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.micro"

  # update config and apply: scale up to larger instance
  cache_node_type="cache.t3.small"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new state
  wait_and_assert_replication_group_available_status
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.small"
}

# attempt to scale out a cluster mode disabled RG with no replicas: negative test, expect failure
test_modify_rg_cmd_scale_out() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-cmd-scale-out"
  automatic_failover_enabled="false"
  num_node_groups="1"
  replicas_per_node_group="0"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 0

  # update config and apply: attempt to scale out
  # config application should actually succeed in this case, but leave RG with Terminal Condition set True
  num_node_groups=2
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # check for terminal condition on resource
  conditions=$(k8s_get_rg_field "$rg_id" ".status .conditions[]")
  terminal_cond=$(echo $conditions | jq -r -e 'select(.type == "ACK.Terminal")')
  if [[ $? != 0 ]]; then
    echo "FAIL: replication group $rg_id did not have a terminal condition after attempted scale out"
    exit 1
  fi

  # ensure terminal condition properties are as expected
  status=$(echo $terminal_cond | jq -r -e ".status")
  cond_msg=$(echo $terminal_cond | jq -r -e ".message")
  if [[ $status != "True" || $cond_msg != *"Operation is only applicable for cluster mode enabled"* ]]; then
    echo "FAIL: replication group $rg_id has terminal condition, but with unexpected status or message"
    exit 1
  fi
}


test_modify_rg_enable_auto_failover # passing
test_modify_rg_enable_af_add_replicas # failing due to RG not entering "modifying" status when creating new replicas
test_modify_rg_cmd_scale_up # passing
test_modify_rg_cmd_scale_out # currently failing, terminal condition frequently toggles (is this desired behavior?)

