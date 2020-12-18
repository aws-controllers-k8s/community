#!/usr/bin/env bash

# replication group scaling tests: horizontal and vertical scaling

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/elasticache.sh"

check_is_installed jq "Please install jq before running this script."
AWS_REGION=${AWS_REGION:-"us-west-2"}

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test group: $service_name/$test_name------------------------------"
debug_msg "selected AWS region: $AWS_REGION"

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
  provide_replication_group_yaml | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 0

  # update config and apply: attempt to scale out
  # config application should actually succeed in this case, but leave RG with Terminal Condition set True
  num_node_groups=2
  provide_replication_group_yaml | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure terminal condition exists, is set true, and has expected message
  check_rg_terminal_condition_true "$rg_id" "Operation is only applicable for cluster mode enabled"
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
  provide_replication_group_yaml | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.micro"

  # update config and apply: scale up to larger instance
  cache_node_type="cache.t3.small"
  provide_replication_group_yaml | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new state
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.small"
}

# create a cluster mode enabled RG, then attempt to scale out and increase replica count
test_modify_rg_cme_scale_out_add_replica() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-cme-scale-out-add-replica"
  num_node_groups="2"
  yaml_base="$(provide_replication_group_yaml_for_replica_config)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
      - nodeGroupID: "0020"
        primaryAvailabilityZone: ${AWS_REGION}b
        replicaAvailabilityZones:
        - ${AWS_REGION}a
        replicaCount: 1
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 2
  k8s_assert_replication_group_replica_count "$rg_id" 1
  k8s_assert_replication_group_total_node_count "$rg_id" 4

  # update config and apply: scale out and add replicas
  num_node_groups="3"
  yaml_base="$(provide_replication_group_yaml_for_replica_config)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        - ${AWS_REGION}a
        replicaCount: 2
      - nodeGroupID: "0020"
        primaryAvailabilityZone: ${AWS_REGION}b
        replicaAvailabilityZones:
        - ${AWS_REGION}a
        - ${AWS_REGION}b
        replicaCount: 2
      - nodeGroupID: "0030"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        - ${AWS_REGION}a
        replicaCount: 2
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new resource state
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 3
  k8s_assert_replication_group_replica_count "$rg_id" 2
  k8s_assert_replication_group_total_node_count "$rg_id" 9
}

# scale out a cluster mode enabled RG where replica count is uneven between shards (i.e. there is a replicaCount
#   specified for each node group rather than one replicasPerNodeGroup property for the entire RG)
test_modify_rg_cme_scale_out_uneven_shards() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-cme-scale-out-uneven-shards"
  yaml_base=$(provide_replication_group_yaml_basic "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    automaticFailoverEnabled: true
    cacheNodeType: cache.t3.micro
    numNodeGroups: 2
    multiAZEnabled: true
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
      - nodeGroupID: "0020"
        primaryAvailabilityZone: ${AWS_REGION}b
        replicaAvailabilityZones:
        - ${AWS_REGION}a
        - ${AWS_REGION}c
        replicaCount: 2
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 2
  k8s_assert_replication_group_total_node_count "$rg_id" 5 #skip checking each node group for now

   # update config and apply: scale out and add replicas
  yaml_base=$(provide_replication_group_yaml_basic "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    automaticFailoverEnabled: true
    cacheNodeType: cache.t3.micro
    numNodeGroups: 3
    multiAZEnabled: true
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
      - nodeGroupID: "0020"
        primaryAvailabilityZone: ${AWS_REGION}b
        replicaAvailabilityZones:
        - ${AWS_REGION}a
        - ${AWS_REGION}c
        replicaCount: 2
      - nodeGroupID: "0030"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new resource state
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 3
  k8s_assert_replication_group_total_node_count "$rg_id" 7
}

# basic scale out test for cluster mode enabled replication groups, # replicas/node group unchanged
test_modify_rg_cme_scale_out_basic() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-cme-scale-out-basic"
  num_node_groups="2"
  yaml_base="$(provide_replication_group_yaml_for_replica_config)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
      - nodeGroupID: "0020"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 2
  k8s_assert_replication_group_replica_count "$rg_id" 1
  k8s_assert_replication_group_total_node_count "$rg_id" 4

   # update config and apply: scale out
  num_node_groups="3"
  yaml_base="$(provide_replication_group_yaml_for_replica_config)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: "0010"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
      - nodeGroupID: "0020"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
      - nodeGroupID: "0030"
        primaryAvailabilityZone: ${AWS_REGION}a
        replicaAvailabilityZones:
        - ${AWS_REGION}b
        replicaCount: 1
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert resource state
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  k8s_assert_replication_group_shard_count "$rg_id" 3
  k8s_assert_replication_group_replica_count "$rg_id" 1
  k8s_assert_replication_group_total_node_count "$rg_id" 6
}

# run tests
test_modify_rg_cmd_scale_out
test_modify_rg_cmd_scale_up
test_modify_rg_cme_scale_out_add_replica # failing, terminal condition shows "2 validation errors" after new config - issue with distribution of AZs in test case?
test_modify_rg_cme_scale_out_uneven_shards
test_modify_rg_cme_scale_out_basic

k8s_perform_rg_test_cleanup