#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

check_is_installed jq

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="elasticache"
ack_ctrl_pod_id=$( controller_pod_id "$service_name")
debug_msg "executing test: $service_name/$test_name"

aws_resource_name="ack-test-smoke-${service_name}-subnet-gp"
k8s_resource_name="cachesubnetgroups/$aws_resource_name"

# pre-req: a subnet id to create cache subnet group
aws_subnet_id="$(aws ec2 describe-subnets --output json | jq -e '.Subnets[0] | .SubnetId')"

if [ $? -eq 4 ] || [ -z $aws_subnet_id ]; then
    echo "no subnet id found to run the test. Ensure that subnet id is available to run the test."
    exit 1
else
    echo "using subnet id: ${aws_subnet_id} for test."
fi


describe_subnet_json() {
    aws elasticache describe-cache-subnet-groups --cache-subnet-group-name "$aws_resource_name"  --output json >/dev/null 2>&1
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

debug_msg "checking subnet group $aws_resource_name created in ${service_name}"
describe_subnet_json
if [[ $? -eq 255 || $? -eq 254 ]]; then
    echo "FAIL: expected $aws_resource_name to have been created in ${service_name}"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "$k8s_resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

sleep 20

describe_subnet_json
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $aws_resource_name to be deleted in ${service_name}"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi
