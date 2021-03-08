#!/usr/bin/env bash

##############################################
# Tests for AWS ElastiCache Cache Subnet Group
##############################################

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/elasticache.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="elasticache"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

aws_resource_name="ack-test-smoke-${service_name}-subnet-gp"
k8s_resource_name="cachesubnetgroups/$aws_resource_name"

# pre-req: a subnet id to create cache subnet group
if ! aws_subnet_ids_json="$(get_default_subnets)"; then
  echo "FAIL: No default subnet id found to run the test. Ensure that subnet id is available to run the test."
  exit 1
else
  aws_subnet_id="$(echo "$aws_subnet_ids_json" | jq -r -e '.[0]')"
fi

if [ -z "$aws_subnet_id" ]; then
    echo "no subnet id found to run the test. Ensure that subnet id is available to run the test."
    exit 1
else
    echo "using subnet id: ${aws_subnet_id} for test."
fi


describe_subnet_json() {
    daws elasticache describe-cache-subnet-groups --cache-subnet-group-name "$aws_resource_name"  --output json >/dev/null 2>&1
}
get_subnet_group_description() {
  if [[ $# -ne 1 ]]; then
    echo "FAIL: Wrong number of arguments passed to get_subnet_group_description"
    echo "Usage: get_subnet_group_description $subnet_group_name"
    exit 1
  fi
  local subnet_group_name="$1"
  local subnet_group_desc="$(aws elasticache describe-cache-subnet-groups --cache-subnet-group-name "$subnet_group_name" --output json | jq -r -e '.CacheSubnetGroups[] | .CacheSubnetGroupDescription')"
  echo "$subnet_group_desc"
}

# PRE-CHECKS
describe_subnet_json
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $aws_resource_name to not exist in ${service_name}. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$k8s_resource_name"; then
    echo "FAIL: expected $k8s_resource_name to not exist on K8s cluster. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

## Create
debug_msg "Creating subnet group $aws_resource_name in ${service_name}"
cat <<EOF | kubectl apply -f -
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: CacheSubnetGroup
metadata:
  name: $aws_resource_name
spec:
  cacheSubnetGroupDescription: "ack test subnet group test description"
  cacheSubnetGroupName: $aws_resource_name
  subnetIDs:
    - $aws_subnet_id
EOF

sleep 20

describe_subnet_json
if [[ $? -eq 255 || $? -eq 254 ]]; then
    echo "FAIL: expected $aws_resource_name to have been created in ${service_name}"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi
debug_msg "checking subnet group $aws_resource_name created in ${service_name}"

## Modify
debug_msg "Modifying subnet group $aws_resource_name in ${service_name}"
expected_subnet_group_desc='ack test subnet group test description updated'

cat <<EOF | kubectl apply -f -
apiVersion: elasticache.services.k8s.aws/v1alpha1
kind: CacheSubnetGroup
metadata:
  name: $aws_resource_name
spec:
  cacheSubnetGroupDescription: "$expected_subnet_group_desc"
  cacheSubnetGroupName: $aws_resource_name
  subnetIDs:
    - $aws_subnet_id
EOF

sleep 20

actual_subnet_group_desc="$(get_subnet_group_description $aws_resource_name)"
if [ "$expected_subnet_group_desc" != "$actual_subnet_group_desc" ]; then
  echo "FAIL: expected $aws_resource_name to have been modified in ${service_name}"
  log_and_exit "cachesubnetgroups/$aws_resource_name"
fi
debug_msg "subnet group $aws_resource_name modified in ${service_name}"

## Delete
debug_msg "Deleting subnet group $aws_resource_name in ${service_name}"
kubectl delete "$k8s_resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

sleep 20

describe_subnet_json
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $aws_resource_name to be deleted in ${service_name}"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

# pod may restart upon refresh of credentials, remove this check for now
# assert_pod_not_restarted $ack_ctrl_pod_id
