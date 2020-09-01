#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

check_is_installed jq

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="sns"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

topic_name="ack-test-smoke-$service_name"
resource_name="topics/$topic_name"

get_topic_attributes() {
    topic_arn="arn:aws:sns:$AWS_REGION:$AWS_ACCOUNT_ID:$topic_name"
    aws sns get-topic-attributes --topic-arn "$topic_arn" --output json >/dev/null 2>&1
}

# PRE-CHECKS
get_topic_attributes
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $topic_name to not exist in SNS. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

cat <<EOF | kubectl apply -f -
apiVersion: sns.services.k8s.aws/v1alpha1
kind: Topic
metadata:
  name: $topic_name
spec:
  name: $topic_name
EOF

sleep 20

debug_msg "checking topic $topic_name created in SNS"
get_topic_attributes
if [[ $? -eq 255 || $? -eq 254 ]]; then
    echo "FAIL: expected $topic_name to have been created in SNS"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

get_topic_attributes
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $topic_name to be deleted in SNS"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

assert_pod_not_restarted $ack_ctrl_pod_id
