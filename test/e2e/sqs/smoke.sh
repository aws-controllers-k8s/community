#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="sqs"
ack_ctrl_pod_id=$( controller_pod_id "sqs")
debug_msg "executing test: $service_name/$test_name"

queue_name="ack-test-smoke-$service_name"
resource_name="queues/$queue_name"

get_queue_url() {
    aws sqs get-queue-url --queue-name "$queue_name" --output json >/dev/null 2>&1
}

# PRE-CHECKS
get_queue_url
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $queue_name to not exist in SQS. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

cat <<EOF | kubectl apply -f -
apiVersion: sqs.services.k8s.aws/v1alpha1
kind: Queue
metadata:
  name: $queue_name
spec:
  name: $queue_name
EOF

sleep 20

debug_msg "checking queue $queue_name created in SQS"
get_queue_url
if [[ $? -eq 255 || $? -eq 254 ]]; then
    echo "FAIL: expected $queue_name to have been created in SQS"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

get_queue_url
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $queue_name to be deleted in SQS"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi
