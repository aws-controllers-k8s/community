#!/usr/bin/env bash

# replication group miscellaneous tests

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
debug_msg "executing test group: $service_name/$test_name------------------------------"
debug_msg "selected AWS region: $AWS_REGION"

# test creation of replication group with explicit security group
test_create_rg_specify_sg() {

  # create security group in default VPC to use
  local vpc_json=$(daws ec2 describe-vpcs | jq -r '.Vpcs[] | select( .IsDefault == true )')
  local default_vpc_id=$(echo "$vpc_json" | jq -r '.VpcId')
  daws ec2 create-security-group --group-name "test-sg-default" --description "sg for automated elasticache ACK test" --vpc-id "$default_vpc_id" 1>/dev/null 2>&1

  # retrieve security group ID from newly created security group
  local sg_id=$(daws ec2 describe-security-groups | jq -r -e '.SecurityGroups[] | select( .GroupName == "test-sg-default" ) | .GroupId')
  assert_equal "0" "$?" "Could not find security group ID for security group test-sg-default" || exit 1

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-specify-sg"
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  yaml_base=$(provide_replication_group_yaml "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    securityGroupIDs:
      - $sg_id
EOF
)
  echo "$rg_yaml" | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_synced_and_available "$rg_id"
  local primary_cluster=$(aws_get_replication_group_json "$rg_id" | jq -r -e ".MemberClusters[0]")
  assert_equal "0" "$?" "Could not find cache cluster for replication group $rg_id" || exit 1
  daws elasticache describe-cache-clusters --cache-cluster-id "$primary_cluster" | jq -r '.CacheClusters[0]' | grep "$sg_id"
  assert_equal "0" "$?" "Expected replication group $rg_id to have security group $sg_id" || exit 1
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
  provide_replication_group_yaml | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure first RG successfully created and available
  wait_and_assert_replication_group_synced_and_available "$rg_id"

  # generate and apply yaml for creation of second replication group
  rg_id="rg-deletion-2"
  provide_replication_group_yaml | kubectl apply -f - 2>&1
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure second RG successfully created and available
  wait_and_assert_replication_group_synced_and_available "$rg_id"

  # delete and wait for deletion to complete
  kubectl delete ReplicationGroup --all 2>/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  aws_wait_replication_group_deleted "rg-deletion-1" "FAIL: expected replication group rg-deletion-1 to have been deleted in ${service_name}"
  aws_wait_replication_group_deleted "rg-deletion-2" "FAIL: expected replication group rg-deletion-2 to have been deleted in ${service_name}"
}

# run tests
test_create_rg_specify_sg # failing
test_rg_deletion_multiple

k8s_perform_rg_test_cleanup