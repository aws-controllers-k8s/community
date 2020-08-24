#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

check_is_installed jq

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="s3"
ack_ctrl_pod_id=$( controller_pod_id "s3")
debug_msg "executing test: $service_name/$test_name"

bucket_name="ack-test-smoke-$service_name-$AWS_ACCOUNT_ID"
resource_name="buckets/$bucket_name"

list_bucket_json() {
    jq_expr='.Buckets[] | select(.Name | contains($BUCKET_NAME))'
    aws s3api list-buckets --output json | jq -e --arg BUCKET_NAME "$bucket_name" "$jq_expr"
}

# PRE-CHECKS
list_bucket_json
if [ $? -ne 4 ]; then
    echo "FAIL: expected $bucket_name to not exist in S3. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

cat <<EOF | kubectl apply -f -
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: $bucket_name
spec:
  name: $bucket_name
EOF

sleep 20

debug_msg "checking bucket $bucket_name created in S3"
list_bucket_json
if [ $? -eq 4 ]; then
    echo "FAIL: expected $bucket_name to have been created in S3"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

list_bucket_json
if [ $? -ne 4 ]; then
    echo "FAIL: expected $bucket_name to be deleted in S3"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi
