#!/usr/bin/env bash

# snapshot: basic e2e tests

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

# basic test covering the four snapshot APIs
test_snapshot_CRUD() {
  debug_msg "executing ${FUNCNAME[0]}"

  # delete snapshots if they already exist - no need to wait due to upcoming replication group wait
  local snapshot_name="snapshot-test"
  local copied_snapshot_name="snapshot-copy"
  daws elasticache delete-snapshot --snapshot-name "$snapshot_name" 1>/dev/null 2>&1
  daws elasticache delete-snapshot --snapshot-name "$copied_snapshot_name" 1>/dev/null 2>&1

  # delete replication group if it already exists (we want it to be created to below specification)
  clear_rg_parameter_variables
  rg_id="rg-snapshot-test" # non-local because for now, provide_replication_group_yaml uses unscoped variables
  daws elasticache describe-replication-groups --replication-group-id "$rg_id" 1>/dev/null 2>&1
  if [[ "$?" == "0" ]]; then
    daws elasticache delete-replication-group --replication-group-id "$rg_id" 1>/dev/null 2>&1
    aws_wait_replication_group_deleted  "$rg_id" "FAIL: expected replication group $rg_id to have been deleted in ${service_name}"
  fi

  # create replication group for snapshot
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"
  wait_and_assert_replication_group_available_status

  # proceed to CRUD test: create first snapshot
  local cc_id="$rg_id-001"
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $snapshot_name
spec:
  snapshotName: $snapshot_name
  cacheClusterID: $cc_id
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $snapshot_name to succeed" || exit 1
  k8s_wait_resource_synced "snapshots/$snapshot_name" 10

  # create second snapshot from first to trigger copy-snapshot API
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $copied_snapshot_name
spec:
  snapshotName: $copied_snapshot_name
  sourceSnapshotName: $snapshot_name
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $copied_snapshot_name to succeed" || exit 1
  k8s_wait_resource_synced "snapshots/$snapshot_name" 20

  # test deletion
  kubectl delete snapshots/"$snapshot_name"
  kubectl delete snapshots/"$copied_snapshot_name"
  aws_wait_snapshot_deleted "$snapshot_name"
  aws_wait_snapshot_deleted "$copied_snapshot_name"
}

# tests creation of snapshots for cluster mode disabled
test_snapshot_CMD_creates() {
  debug_msg "executing ${FUNCNAME[0]}"

  # delete replication group if it already exists (we want it to be created to below specification)
  clear_rg_parameter_variables
  rg_id="snapshot-test-cmd"
  daws elasticache describe-replication-groups --replication-group-id "$rg_id" 1>/dev/null 2>&1
  if [[ "$?" == "0" ]]; then
    daws elasticache delete-replication-group --replication-group-id "$rg_id" 1>/dev/null 2>&1
    aws_wait_replication_group_deleted  "$rg_id" "FAIL: expected replication group $rg_id to have been deleted in ${service_name}"
  fi

  # create cluster mode disabled replication group for snapshot
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"
  wait_and_assert_replication_group_available_status

  # case 1: specify only the replication group - should fail as RG snapshot not permitted for CMD RG
  local snapshot_name="snapshot-cmd"
  daws elasticache delete-snapshot --snapshot-name "$snapshot_name" 1>/dev/null 2>&1
  sleep 10
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $snapshot_name
spec:
  snapshotName: $snapshot_name
  replicationGroupID: $rg_id
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $snapshot_name to succeed" || exit 1
  sleep 10 # give time for server validation
  k8s_check_resource_terminal_condition_true "snapshots/$snapshot_name" "Cannot snapshot a replication group with cluster-mode disabled"

  # case 1 test cleanup
  kubectl delete snapshots/"$snapshot_name"
  aws_wait_snapshot_deleted "$snapshot_name"

  # case 2: specify both RG and cache cluster ID (should succeed)
  local snapshot_name="snapshot-cmd"
  daws elasticache delete-snapshot --snapshot-name "$snapshot_name" 1>/dev/null 2>&1
  sleep 10
  local cc_id="$rg_id-001"
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $snapshot_name
spec:
  snapshotName: $snapshot_name
  replicationGroupID: $rg_id
  cacheClusterID: $cc_id
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $snapshot_name to succeed" || exit 1
  k8s_wait_resource_synced "snapshots/$snapshot_name" 20

  # delete snapshot for case 2 if creation succeeded
  kubectl delete snapshots/"$snapshot_name"
  aws_wait_snapshot_deleted "$snapshot_name"
}

test_snapshot_CME_creates() {
  debug_msg "executing ${FUNCNAME[0]}"

  # delete replication group if it already exists (we want it to be created to below specification)
  clear_rg_parameter_variables
  rg_id="snapshot-test-cme"
  daws elasticache describe-replication-groups --replication-group-id "$rg_id" 1>/dev/null 2>&1
  if [[ "$?" == "0" ]]; then
    daws elasticache delete-replication-group --replication-group-id "$rg_id" 1>/dev/null 2>&1
    aws_wait_replication_group_deleted  "$rg_id" "FAIL: expected replication group $rg_id to have been deleted in ${service_name}"
  fi

  # create cluster mode enabled replication group for snapshot
  num_node_groups=2
  replicas_per_node_group=1
  automatic_failover_enabled="true"
  multi_az_enabled="true"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"
  wait_and_assert_replication_group_available_status

  # case 1: specify only RG
  local snapshot_name="snapshot-cme"
  daws elasticache delete-snapshot --snapshot-name "$snapshot_name" 1>/dev/null 2>&1
  sleep 10
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $snapshot_name
spec:
  snapshotName: $snapshot_name
  replicationGroupID: $rg_id
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $snapshot_name to succeed" || exit 1
  k8s_wait_resource_synced "snapshots/$snapshot_name" 20

  # delete snapshot for case 1 if creation succeeded
  kubectl delete snapshots/"$snapshot_name"
  aws_wait_snapshot_deleted "$snapshot_name"

  # case 2: specify both RG and cache cluster ID
  local snapshot_name="snapshot-cme"
  daws elasticache delete-snapshot --snapshot-name "$snapshot_name" 1>/dev/null 2>&1
  sleep 10
  local cc_id="$rg_id-001"
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $snapshot_name
spec:
  snapshotName: $snapshot_name
  replicationGroupID: $rg_id
  cacheClusterID: $cc_id
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $snapshot_name to succeed" || exit 1
  k8s_wait_resource_synced "snapshots/$snapshot_name" 20

  # delete snapshot for case 2 if creation succeeded
  kubectl delete snapshots/"$snapshot_name"
  aws_wait_snapshot_deleted "$snapshot_name"
}

# test snapshot creation while specifying KMS key
test_snapshot_create_KMS() {
  debug_msg "executing ${FUNCNAME[0]}"

  # create KMS key and get key ID
  local output=$(daws kms create-key --output json)
  assert_equal "0" "$?" "Expected creation of KMS key to succeed" || exit 1

  local key_id=$(echo "$output" | jq -r -e ".KeyMetadata.KeyId")
  assert_equal "0" "$?" "Key ID does not exist for KMS key" || exit 1

  # delete replication group if it already exists (we want it to be created to below specification)
  clear_rg_parameter_variables
  rg_id="snapshot-test-kms"
  daws elasticache describe-replication-groups --replication-group-id "$rg_id" 1>/dev/null 2>&1
  if [[ "$?" == "0" ]]; then
    daws elasticache delete-replication-group --replication-group-id "$rg_id" 1>/dev/null 2>&1
    aws_wait_replication_group_deleted  "$rg_id" "FAIL: expected replication group $rg_id to have been deleted in ${service_name}"
  fi

  # create cluster mode disabled replication group for snapshot
  num_node_groups=1
  replicas_per_node_group=0
  automatic_failover_enabled="false"
  multi_az_enabled="false"
  output_msg=$(provide_replication_group_yaml | kubectl apply -f - 2>&1)
  exit_if_rg_config_application_failed $? "$rg_id"
  wait_and_assert_replication_group_available_status

  # create snapshot while specifying KMS key
  local snapshot_name="snapshot-kms"
  daws elasticache delete-snapshot --snapshot-name "$snapshot_name" 1>/dev/null 2>&1
  sleep 10
  local cc_id="$rg_id-001"
  local snapshot_yaml=$(cat <<EOF
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: Snapshot
metadata:
  name: $snapshot_name
spec:
  snapshotName: $snapshot_name
  cacheClusterID: $cc_id
  kmsKeyID: $key_id
EOF)
  echo "$snapshot_yaml" | kubectl apply -f -
  assert_equal "0" "$?" "Expected application of $snapshot_name to succeed" || exit 1
  k8s_wait_resource_synced "snapshots/$snapshot_name" 20

  # delete snapshot for case 1 if creation succeeded
  kubectl delete snapshots/"$snapshot_name"
  aws_wait_snapshot_deleted "$snapshot_name"
}

# run tests
test_snapshot_CRUD
test_snapshot_CMD_creates #issue: second snapshot doesn't have "status" property - problem with yaml or something else?
test_snapshot_CME_creates #same issue as above
test_snapshot_create_KMS #IAM role needs KMS access for this to work