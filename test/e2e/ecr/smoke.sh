#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
ack_ctrl_pod_id=$( controller_pod_id "ecr")
debug_msg "executing test: $test_name"

repo_name="ack-test-smoke-ecr"
resource_name="repositories/$repo_name"

# PRE-CHECKS

aws ecr describe-repositories --repository-names "$repo_name" --output json >/dev/null 2>&1
if [ $? -ne 255 ]; then
    echo "FAIL: expected $repo_name to not exist in ECR. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

cat <<EOF | kubectl apply -f -
apiVersion: ecr.services.k8s.aws/v1alpha1
kind: Repository
metadata:
  name: $repo_name
spec:
  repositoryName: $repo_name
EOF

sleep 5

debug_msg "checking repository $repo_name created in ECR"
aws ecr describe-repositories --repository-names "$repo_name" --output table
if [ $? -eq 255 ]; then
    echo "FAIL: expected $repo_name to have been created in ECR"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

aws ecr describe-repositories --repository-names "$repo_name" --output json >/dev/null 2>&1
if [ $? -ne 255 ]; then
    echo "FAIL: expected $repo_name to deleted in ECR"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi
