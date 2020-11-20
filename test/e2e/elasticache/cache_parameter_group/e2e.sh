#!/usr/bin/env bash

##############################################
# Tests for AWS ElastiCache Cache Parameter Group
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

setup_cache_parameter_group_fields() {
  # uses non local variable for later use in tests
  cpg_name="ack-test-cpg-1"
  cpg_description="$cpg_name description"
  cpg_parameter_1_name="activedefrag"
  cpg_parameter_1_value="yes"
  cpg_parameter_2_name="active-defrag-cycle-max"
  cpg_parameter_2_value="74"
  cpg_parameter_3_name="active-defrag-cycle-min"
  cpg_parameter_3_value="10"
}
setup_cache_parameter_group_fields

#################################################
# create cache parameter group
#################################################

ack_create_cache_parameter_group() {
  local cpg_yaml="$(provide_cache_parameter_group_yaml)"
  echo "$cpg_yaml" | kubectl apply -f -
  sleep 10
}
assert_k8s_status_cache_parameters_defaults() {
  assert_k8s_status_parameters_value_source "$cpg_name" "activedefrag" "no" "system"
  assert_k8s_status_parameters_value_source "$cpg_name" "active-defrag-cycle-max" "75" "system"
  assert_k8s_status_parameters_value_source "$cpg_name" "active-defrag-cycle-min" "5" "system"
}

debug_msg "Testing create Cache Parameter Group: $cpg_name."
assert_cache_parameter_group_does_not_exist "$cpg_name"
ack_create_cache_parameter_group
assert_cache_parameter_group_exists "$cpg_name"
assert_k8s_status_cache_parameters_defaults

#################################################
# modify cache parameter group
#################################################
debug_msg "Testing modify Cache Parameter Group: $cpg_name."
#########################
## Add parameters
#########################
debug_msg "Testing Add Parameters to Cache Parameter Group: $cpg_name."

assert_no_custom_cache_parameters() {
  local actual_value=$(aws_get_cache_parameters_property "$cpg_name" ".Parameters" "user" | jq length)
  assert_equal "0" "$actual_value" "Expected: 0 actual: $actual_value found for user parameters in cache parameter group $cpg_name" || exit 1
}

ack_set_custom_cache_parameters() {
  local cpg_yaml="$(provide_custom_cache_parameters_group_yaml)"
  echo "$cpg_yaml" | kubectl apply -f -
  sleep 10
}

assert_custom_cache_parameters() {
  local actual_parameters=$(aws_get_cache_parameters_property "$cpg_name" ".Parameters" "user")
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_1_name" "$cpg_parameter_1_value"
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_2_name" "$cpg_parameter_2_value"
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_3_name" "$cpg_parameter_3_value"
}

assert_k8s_status_cache_parameters_custom() {
  assert_k8s_status_parameters_value_source "$cpg_name" "$cpg_parameter_1_name" "$cpg_parameter_1_value" "user"
  assert_k8s_status_parameters_value_source "$cpg_name" "$cpg_parameter_2_name" "$cpg_parameter_2_value" "user"
  assert_k8s_status_parameters_value_source "$cpg_name" "$cpg_parameter_3_name" "$cpg_parameter_3_value" "user"
}

assert_no_custom_cache_parameters
ack_set_custom_cache_parameters
assert_custom_cache_parameters
assert_k8s_status_cache_parameters_custom

#########################
## Update parameter
#########################
debug_msg "Testing Update Parameters to Cache Parameter Group: $cpg_name."
update_cache_parameter_group_fields() {
  # uses non local variable for later use in tests
  cpg_parameter_1_value="no"
  cpg_parameter_2_value="70"
  cpg_parameter_3_value="15"
}

update_cache_parameter_group_fields
ack_set_custom_cache_parameters
assert_custom_cache_parameters

#########################
## Remove parameter
#########################
debug_msg "Testing Remove Parameters to Cache Parameter Group: $cpg_name."
ack_remove_custom_cache_parameters() {
  # keeps only parameter1. removes parameter2, sets parameter 3 to ""
  local cpg_yaml="$(provide_custom_remove_cache_parameters_group_yaml)"
  echo "$cpg_yaml" | kubectl apply -f -
  sleep 10
}

assert_remove_custom_cache_parameters() {
  # verify only parameter 1 is of source type 'user'
  local actual_parameters=$(aws_get_cache_parameters_property "$cpg_name" ".Parameters" "user")
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_1_name" "$cpg_parameter_1_value"
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_2_name" ""
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_3_name" ""

  # validate that the parameter 2 and 3 are now system default
  local actual_parameters=$(aws_get_cache_parameters_property "$cpg_name" ".Parameters" "system")
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_2_name" "75"
  assert_parameters_name_value "$actual_parameters" "$cpg_parameter_3_name" "5"
}

assert_k8s_status_cache_parameters_remove_custom() {
  assert_k8s_status_parameters_value_source "$cpg_name" "$cpg_parameter_1_name" "$cpg_parameter_1_value" "user"
  assert_k8s_status_parameters_value_source "$cpg_name" "$cpg_parameter_2_name" "75" "system"
  assert_k8s_status_parameters_value_source "$cpg_name" "$cpg_parameter_3_name" "5" "system"
}

ack_remove_custom_cache_parameters
assert_remove_custom_cache_parameters
assert_k8s_status_cache_parameters_remove_custom

#################################################
# reset cache parameter group
# (remove all parameters)
#################################################
debug_msg "Testing Reset Parameters to Cache Parameter Group: $cpg_name."

reset_all_custom_cache_parameters() {
  # yaml has no parameters
  local cpg_yaml="$(provide_cache_parameter_group_yaml)"
  echo "$cpg_yaml" | kubectl apply -f -
  sleep 10
}
reset_all_custom_cache_parameters
assert_no_custom_cache_parameters
assert_k8s_status_cache_parameters_defaults

#################################################
# delete cache parameter group
#################################################
debug_msg "Testing delete Cache Parameter Group: $cpg_name."

ack_delete_cache_parameter_group() {
  kubectl delete CacheParameterGroup/"$cpg_name" 2>/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1
  sleep 5
}
ack_delete_cache_parameter_group
assert_cache_parameter_group_does_not_exist "$cpg_name"

debug_msg "Test completed."
