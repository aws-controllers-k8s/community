#!/usr/bin/env bash

# tests covering "replication", i.e. adding/removing replicas, auto failover, etc.

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

# create cluster mode disabled replication group without replicas, attempt to modify to
#   negative replica count: negative test, expect failure
test_modify_rg_negative_replicas() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-rg-modify-to-negative-replicas"
  automatic_failover_enabled="false"
  num_node_groups="1"
  replicas_per_node_group="0"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_replica_count "$rg_id" 0

  # update config and apply: attempt to change to negative replica count
  replicas_per_node_group="-1"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  check_rg_terminal_condition_true "$rg_id" "New replica count must be between"
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

# modify replication group to enable auto failover
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

# create cluster mode disabled replication group with one replica and auto failover enabled, attempt to remove
#   replica while keeping auto failover enabled: negative test, expect failure
test_modify_rg_remove_replica_with_af_enabled() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-rg-remove-last-replica-af-enabled"
  automatic_failover_enabled="true"
  num_node_groups="1"
  replicas_per_node_group="1"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_replica_count "$rg_id" 1
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "enabled"

  # update config and apply: attempt to remove replica while keeping auto failover enabled
  replicas_per_node_group="0"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  check_rg_terminal_condition_true "$rg_id" "Must have at least 1 replica when cluster mode is disabled with auto failover enabled"
}

# create cluster mode disabled replication group with one replica and auto failover enabled, then remove the
#   replica while disabling auto failover
test_modify_rg_remove_replica_disable_af() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="test-rg-remove-last-replica-disable-af"
  automatic_failover_enabled="true"
  num_node_groups="1"
  replicas_per_node_group="1"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_replica_count "$rg_id" 1
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "enabled"
  k8s_assert_replication_group_total_node_count "$rg_id" 2

  # update config and apply: remove replica while disabling auto failover
  replicas_per_node_group="0"
  automatic_failover_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new state
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_replica_count "$rg_id" 0
  k8s_assert_replication_group_status_property "$rg_id" ".automaticFailover" "disabled"
  k8s_assert_replication_group_total_node_count "$rg_id" 1
}

# create replication group with one shard and one replica with custom node group configuration. Increase
#   replicas per node group and specify preferred AZ
test_modify_rg_increase_replica_specify_az() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-inc-replica-specify-az"
  num_node_groups=1
  replicas_per_node_group=1
  yaml_base="$(provide_replication_group_yaml)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: us-east-1a
        replicaAvailabilityZones:
          - us-east-1b
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 1

  # update config and apply: increase replica count, specify additional AZ
  replicas_per_node_group=2
  yaml_base="$(provide_replication_group_yaml)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: us-east-1a
        replicaAvailabilityZones:
          - us-east-1b
          - us-east-1a
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert resource state
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 2
}

# create a cluster mode enabled RG, then scale up while adding replicas
test_modify_rg_cme_scale_up_add_replicas() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-cme-scale-up-add-replicas"
  cache_node_type="cache.t3.micro"
  num_node_groups="2"
  replicas_per_node_group="1"
  yaml_base=$(provide_replication_group_yaml)
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: us-east-1a
        replicaAvailabilityZones:
        - us-east-1b
      - nodeGroupID: "0020"
        primaryAvailabilityZone: us-east-1b
        replicaAvailabilityZones:
        - us-east-1a
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 2
  k8s_assert_replication_group_replica_count "$rg_id" 1
  k8s_assert_replication_group_total_node_count "$rg_id" 4
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.micro"

  echo "past initial creation and assertion"

  # update config and apply: scale out and add replicas
  replicas_per_node_group="2"
  cache_node_type="cache.t3.small"
  yaml_base=$(provide_replication_group_yaml "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: us-east-1a
        replicaAvailabilityZones:
        - us-east-1b
        - us-east-1a
      - nodeGroupID: "0020"
        primaryAvailabilityZone: us-east-1b
        replicaAvailabilityZones:
        - us-east-1a
        - us-east-1b
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  echo "past application of modified config"

  # wait and assert new resource state
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 2
  k8s_assert_replication_group_replica_count "$rg_id" 2
  k8s_assert_replication_group_total_node_count "$rg_id" 6
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.small"

  echo "passed test"
}

# ensure node roles are correct after failover: create a cluster mode disabled RG with one replica and
#   invoke the test-failover API. Ensure node roles from k8s are in sync with node roles from AWS CLI
test_rg_failover_roles() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-failover-roles"
  num_node_groups=1
  replicas_per_node_group=1
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure second RG successfully created and available, assert initial node roles
  wait_and_assert_replication_group_available_status
  local shard_json=$(aws_get_replication_group_json "$rg_id" | jq -r -e '.NodeGroups[0]')
  local node1_role=$(echo "$shard_json" | jq -r -e '.NodeGroupMembers[] | select(.CacheClusterId=="rg-failover-roles-001") | .CurrentRole')
  local node2_role=$(echo "$shard_json" | jq -r -e '.NodeGroupMembers[] | select(.CacheClusterId=="rg-failover-roles-002") | .CurrentRole')
  assert_equal "primary" "$node1_role" "Node $rg_id-001 has role $node1_role, but expected primary before failover" || exit 1
  assert_equal "replica" "$node2_role" "Node $rg_id-002 has role $node2_role, but expected replica before failover" || exit 1

  # call test-failover API to trigger failover to replica
  daws elasticache test-failover --replication-group-id "$rg_id" --node-group-id "0001" 1>/dev/null 2>&1
  local err_code=$?
  assert_equal "0" "$err_code" "Expected success from test-failover call but got $err_code" || exit 1

  # wait for failover to complete (initial primary node takes role "replica")
  local wait_failed="true"
  for i in $(seq 0 9); do
    sleep 30
    k8s_controller_reload_credentials "$service_name"
    local shard_json=$(aws_get_replication_group_json "$rg_id" | jq -r -e '.NodeGroups[0]')
    local node1_role=$(echo "$shard_json" | jq -r -e '.NodeGroupMembers[] | select(.CacheClusterId=="rg-failover-roles-001") | .CurrentRole')
    if [[ "$node1_role" == "replica" ]]; then
      wait_failed="false"
      break
    fi
  done
  if [[ $wait_failed == "true" ]]; then
    echo "FAIL: node $rg_id-001 should have role replica after failover operation"
    exit 1
  fi

  # roles updated in service at this point, ensure roles in k8s status match
  local shard_k8s=$(k8s_get_rg_field "$rg_id" ".status .nodeGroups[0]")
  local node1_role_k8s=$(echo "$shard_k8s" | jq -r -e '.nodeGroupMembers[] | select(.cacheClusterID=="rg-failover-roles-001") | .currentRole')
  local node2_role_k8s=$(echo "$shard_k8s" | jq -r -e '.nodeGroupMembers[] | select(.cacheClusterID=="rg-failover-roles-002") | .currentRole')

  assert_equal "replica" "$node1_role_k8s" "Node $rg_id-001 has role $node1_role, but expected replica after failover" || exit 1
  assert_equal "primary" "$node2_role_k8s" "Node $rg_id-002 has role $node2_role, but expected primary after failover" || exit 1
}

# run tests
test_modify_rg_negative_replicas # failing, same terminal condition "toggling" issue
test_modify_rg_enable_af_add_replicas # failing due to RG not entering "modifying" status when creating new replicas
test_modify_rg_enable_auto_failover
test_modify_rg_remove_replica_with_af_enabled # failing due to same reason
test_modify_rg_remove_replica_disable_af # failing due to available/modifying status and member cluster count check
test_modify_rg_increase_replica_specify_az # failing due to available during replica creation issue
test_modify_rg_cme_scale_up_add_replicas # failing (probably similar reason to other tests that add replicas)
test_rg_failover_roles

k8s_perform_rg_test_cleanup