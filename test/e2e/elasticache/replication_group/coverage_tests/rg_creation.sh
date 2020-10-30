#!/usr/bin/env bash

# replication group creation tests: testing valid and invalid inputs

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

# attempt creation of replication group with numeric name: negative test, expect failure
test_create_rg_numeric_name() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="12345"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  error_code=$?

  # kubectl apply should fail given a numeric resource name
  if [ $error_code -eq 0 ]; then
    echo "FAIL: expected creation of replication group $rg_id to have failed due to numeric name"
    exit 1
  fi

  # check that error message is the one we expect
  if [[ $output_msg != *"unable to decode \"STDIN\""*  ]]; then
    echo "FAIL: creation of replication group $rg_id failed as expected, but error message different than expected:"
    echo "$output_msg"
    exit 1
  fi
}

# attempt creation of RG with invalid name (has space): negative test, expect failure
test_create_rg_name_contains_spaces() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="new rg"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  error_code=$?

  # kubectl apply should fail given a resource name with spaces
  if [ $error_code -eq 0 ]; then
    echo "FAIL: expected creation of replication group $rg_id to have failed since name contains spaces"
    exit 1
  fi

  # check that error message is the one we expect
  if [[ $output_msg != *"a DNS-1123 subdomain must consist of"* ]]; then
    echo "FAIL: creation of replication group $rg_id failed as expected, but error message different than expected:"
    echo "$output_msg"
    exit 1
  fi
}

# attempt creation of RG with capital letters in name: negative test, expect failure
test_create_rg_mixed_case_name() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="newRG"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  error_code=$?

  # kubectl apply should fail given a mixed-case resource name
  if [ $error_code -eq 0 ]; then
    echo "FAIL: expected creation of replication group $rg_id to have failed due to mixed-case name"
    exit 1
  fi

  # check that error message is the one we expect
  if [[ $output_msg != *"a DNS-1123 subdomain must consist of"* ]]; then
    echo "FAIL: creation of replication group $rg_id failed as expected, but error message different than expected:"
    echo "$output_msg"
    exit 1
  fi
}

# create replication group with one node group (cluster mode disabled), no replicas
test_create_rg_single_shard_no_replicas() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="single-shard-no-replicas"
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available
  wait_and_assert_replication_group_available_status
}

# create RG with custom node group specification where ID isn't surrounded by quotes: negative test,
#   expect failure
test_create_rg_specify_node_group_no_quotes() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-custom-node-no-quotes"
  automatic_failover_enabled="true"
  num_node_groups=1
  replicas_per_node_group=1
  multi_az_enabled="false"
  yaml_base="$(provide_replication_group_yaml)"
  rg_yaml=$(cat <<EOF
$yaml_base
    nodeGroupConfiguration:
      - nodeGroupID: 1020
        primaryAvailabilityZone: us-east-1a
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  error_code=$?

  # kubectl apply should fail if node group ID is not placed within quotes
  if [ $error_code -eq 0 ]; then
    echo "FAIL: expected config application for replication group $rg_id to have failed"
    exit 1
  fi

  # check that error message is the one we expect
  if [[ $output_msg != *"spec.nodeGroupConfiguration.nodeGroupID: Invalid value"*  ]]; then
    echo "FAIL: creation of replication group $rg_id failed as expected, but error message different than expected:"
    echo "$output_msg"
    exit 1
  fi
}

# create replication group with custom nodeGroupConfiguration
test_create_rg_custom_node_config() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-custom-node-config"
  num_node_groups=1
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

  # ensure resource successfully created and available
  wait_and_assert_replication_group_available_status
}

# create RG with custom param group
test_create_rg_custom_param_group() {
  debug_msg "executing ${FUNCNAME[0]}"

  # create custom param group
  k8s_controller_reload_credentials "elasticache"
  daws elasticache create-cache-parameter-group --cache-parameter-group-name "pgtest" --cache-parameter-group-family "redis6.x" --description "test" 1>/dev/null 2>&1
  daws elasticache modify-cache-parameter-group --cache-parameter-group-name "pgtest" --parameter-name-values "ParameterName=reserved-memory-percent,ParameterValue=30" 1>/dev/null 2>&1

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-custom-param-group"
  num_node_groups=1
  replicas_per_node_group=2
  yaml_base="$(provide_replication_group_yaml)"
  rg_yaml=$(cat <<EOF
$yaml_base
    cacheParameterGroupName: pgtest
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  aws_assert_rg_param_group "$rg_id" "pgtest"
}

# create multiple RGs and check deletion succeeds
test_rg_deletion_multiple() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for creation of first replication group
  clear_rg_parameter_variables
  rg_id="rg-deletion-1"
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure first RG successfully created and available
  wait_and_assert_replication_group_available_status

  # generate and apply yaml for creation of second replication group
  rg_id="rg-deletion-2"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure second RG successfully created and available
  wait_and_assert_replication_group_available_status

  # delete and wait for deletion to complete
  kubectl delete ReplicationGroup --all 2>/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  aws_wait_replication_group_deleted "rg-deletion-1" "FAIL: expected replication group rg-deletion-1 to have been deleted in ${service_name}"
  aws_wait_replication_group_deleted "rg-deletion-2" "FAIL: expected replication group rg-deletion-2 to have been deleted in ${service_name}"
}

# run tests
test_create_rg_numeric_name
test_create_rg_name_contains_spaces
test_create_rg_mixed_case_name
test_create_rg_single_shard_no_replicas
test_create_rg_specify_node_group_no_quotes
test_create_rg_custom_node_config
test_create_rg_custom_param_group
test_rg_deletion_multiple

k8s_perform_rg_test_cleanup