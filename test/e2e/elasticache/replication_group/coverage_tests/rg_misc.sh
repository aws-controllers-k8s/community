#!/usr/bin/env bash

# replication group miscellaneous tests

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/elasticache.sh"

check_is_installed jq "Please install jq before running this script."

# test if an otherwise working modification works when engine version 6.x is explicitly specified
test_modify_rg_redis_6x() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-six-x"
  num_node_groups="1"
  replicas_per_node_group="1"
  yaml_base=$(provide_replication_group_yaml "$rg_id")
  cache_node_type="cache.t3.micro"
  rg_yaml=$(cat <<EOF
$yaml_base
    cacheParameterGroupName: default.redis6.x.cluster.on
    engineVersion: 6.x
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 1
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.micro"
  aws_assert_rg_param_group "$rg_id" "default.redis6.x.cluster.on"

  # update config and apply: attempt scale up
  cache_node_type="cache.t3.small"
  yaml_base=$(provide_replication_group_yaml "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    cacheParameterGroupName: default.redis6.x.cluster.on
    engineVersion: 6.x
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new resource state
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 1
  aws_assert_replication_group_property "$rg_id" ".CacheNodeType" "cache.t3.small"
  aws_assert_rg_param_group "$rg_id" "default.redis6.x.cluster.on"
}

# create replication group specifying auth token, modify to change auth token
test_modify_rg_change_auth_token() {
  debug_msg "executing ${FUNCNAME[0]}"

  # generate and apply yaml for replication group creation
  clear_rg_parameter_variables
  rg_id="rg-auth-token"
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  yaml_base=$(provide_replication_group_yaml "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    transitEncryptionEnabled: true
    authToken: this-is-an-example-token
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 0
  k8s_assert_replication_group_status_property "$rg_id" ".authTokenEnabled" "true"

  # update config and apply: change auth token
  yaml_base=$(provide_replication_group_yaml "$rg_id")
  rg_yaml=$(cat <<EOF
$yaml_base
    transitEncryptionEnabled: true
    authToken: example-token-modified
EOF
)
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # wait and assert new resource state
  wait_and_assert_replication_group_available_status
  k8s_assert_replication_group_shard_count "$rg_id" 1
  k8s_assert_replication_group_replica_count "$rg_id" 0
  k8s_assert_replication_group_status_property "$rg_id" ".authTokenEnabled" "true"
}

# test creation of replication group with explicit security group
test_create_rg_specify_sg() {

  # create security group in default VPC to use
  k8s_controller_reload_credentials "elasticache"
  local vpc_json=$(daws ec2 describe-vpcs | jq -r -e '.Vpcs[] | select( .IsDefault == true )')
  local default_vpc_id=$(echo "$vpc_json" | jq -r -e '.VpcId')
  daws ec2 create-security-group --group-name "test-sg-default" --description "sg for automated elasticache ACK test" --vpc-id "$default_vpc_id" 1>/dev/null 2>&1

  # retrieve security group ID from newly created security group
  local sg_id=$(daws ec2 describe-security-groups | jq -r -e '.SecurityGroups[] | select( .GroupName == "test-sg-default" ) | .GroupId')

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
  output_msg=$(echo "$rg_yaml" | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"

  # ensure resource successfully created and available, check resource is as expected
  wait_and_assert_replication_group_available_status
  local primary_cluster=$(aws_get_replication_group_json "$rg_id" | jq -r -e ".MemberClusters[0]")
  daws elasticache describe-cache-clusters --cache-cluster-id "$primary_cluster" | jq -r -e '.CacheClusters[0]' | grep "$sg_id"
  if [[ $? != 0 ]]; then
    echo "FAIL: expected replication group $rg_id to have security group $sg_id"
    exit 1
  fi
}

# run tests
test_modify_rg_redis_6x # failing (expected due to known issue with 6.x)
test_modify_rg_change_auth_token # failing
test_create_rg_specify_sg # failing

k8s_perform_rg_test_cleanup