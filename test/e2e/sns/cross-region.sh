#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/aws/sns.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

AWS_ACCOUNT_ID=$( aws_account_id )
TESTING_NAMESPACE=${TESTING_NAMESPACE:-"testing-$RANDOM"}
AWS_REGION_OVERRIDE=${AWS_REGION_OVERRIDE:-"eu-west-3"}
AWS_REGION_DEFAULT=${AWS_REGION_DEFAULT:-"eu-central-1"}

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="sns"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

topic_name="ack-test-smoke-$service_name"
resource_name="topics/$topic_name"
topic_arn_override_region="arn:aws:sns:$AWS_REGION_OVERRIDE:$AWS_ACCOUNT_ID:$topic_name"
topic_arn_default_region="arn:aws:sns:$AWS_REGION_DEFAULT:$AWS_ACCOUNT_ID:$topic_name"


# PRE-CHECKS
if sns_topic_exists "$topic_arn_override_region" "$AWS_REGION_OVERRIDE"; then
    echo "FAIL: expected $topic_name to not exist in SNS in region $AWS_REGION_OVERRIDE. Did previous test run cleanup?"
    exit 1
fi

if sns_topic_exists "$topic_arn_default_region" "$AWS_REGION_DEFAULT"; then
    echo "FAIL: expected $topic_name to not exist in SNS in region $AWS_REGION_DEFAULT. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# Create the testing namespace
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: $TESTING_NAMESPACE
  annotations:
    services.k8s.aws/default-region: "$AWS_REGION_DEFAULT"
EOF

# TEST ACTIONS and ASSERTIONS

# TEST Creating resource in the namespace default region
cat <<EOF | kubectl apply -f -
apiVersion: sns.services.k8s.aws/v1alpha1
kind: Topic
metadata:
  name: $topic_name
  namespace: $TESTING_NAMESPACE
spec:
  name: $topic_name
EOF

sleep 15

debug_msg "checking topic $topic_name created in SNS in region $AWS_REGION_DEFAULT"

if ! sns_topic_exists "$topic_arn_default_region" "$AWS_REGION_DEFAULT"; then
    echo "FAIL: expected $topic_name to have been created in SNS in region $AWS_REGION_DEFAULT"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl -n "$TESTING_NAMESPACE" delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

sleep 15

if sns_topic_exists "$topic_arn_default_region" "$AWS_REGION_DEFAULT"; then
    echo "FAIL: expected $topic_name to be deleted in SNS in region $AWS_REGION_DEFAULT"
    exit 1
fi

# TEST Overriding the resource region using CR Annotations

cat <<EOF | kubectl apply -f -
apiVersion: sns.services.k8s.aws/v1alpha1
kind: Topic
metadata:
  name: $topic_name
  namespace: $TESTING_NAMESPACE
  annotations:
    services.k8s.aws/region: $AWS_REGION_OVERRIDE
spec:
  name: $topic_name
EOF

sleep 15

debug_msg "checking topic $topic_name created in SNS in region $AWS_REGION_OVERRIDE"
if ! sns_topic_exists "$topic_arn_override_region" "$AWS_REGION_OVERRIDE"; then
    echo "FAIL: expected $topic_name to have been created in SNS in region $AWS_REGION_OVERRIDE"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl -n "$TESTING_NAMESPACE" delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

sleep 15

if sns_topic_exists "$topic_arn_override_region" "$AWS_REGION_OVERRIDE"; then
    echo "FAIL: expected $topic_name to be deleted in SNS in region $AWS_REGION_OVERRIDE"
    exit 1
fi

kubectl delete namespace $TESTING_NAMESPACE 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete namespace but got $?" || exit 1

assert_pod_not_restarted $ack_ctrl_pod_id
