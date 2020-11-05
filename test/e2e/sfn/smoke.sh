#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/aws/sfn.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="sfn"
ack_ctrl_pod_id=$( controller_pod_id )

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: executing test: $service_name/$test_name"



######## STATE MACHINES ########

statemachine_name="ack-test-smoke-$service_name"
statemachine_resource_name="statemachines/$statemachine_name"
statemachine_arn="arn:aws:states:$AWS_REGION:$AWS_ACCOUNT_ID:stateMachine:$statemachine_name"
sfn_executionrole_name="ack-sfn-execution-role"
sfn_executionrole_arn="arn:aws:iam::$AWS_ACCOUNT_ID:role/$sfn_executionrole_name"

debug_msg ""
debug_msg "Resources for Step Functions state machine smoke test:"
debug_msg "statemachine_name: $statemachine_name"
debug_msg "statemachine_resource_name: $statemachine_resource_name"
debug_msg "statemachine_arn: $statemachine_arn"
debug_msg "sfn_executionrole_name: $sfn_executionrole_name"
debug_msg "sfn_executionrole_arn: $sfn_executionrole_arn"
debug_msg ""

# PRE-CHECKS
if sfn_statemachine_exists "$statemachine_arn"; then
    echo "FAIL: expected state machine $statemachine_name to not exist in Step Functions. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$statemachine_resource_name"; then
    echo "FAIL: expected state machine $statemachine_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS
currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Creating execution role in IAM"
sfn_setup_iam_resources_for_executionrole $sfn_executionrole_name
sleep 5

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Creating state machine $statemachine_name"
cat <<EOF | kubectl apply -f -
apiVersion: sfn.services.k8s.aws/v1alpha1
kind: StateMachine
metadata:
  name: $statemachine_name
spec:
  name: $statemachine_name
  roleARN: $sfn_executionrole_arn
  definition: "{ \"StartAt\": \"HelloWorld\", \"States\": { \"HelloWorld\": { \"Type\": \"Pass\", \"Result\": \"Hello World!\", \"End\": true }}}"
EOF

sleep 5

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Checking state machine $statemachine_arn created in Step Functions"
if ! sfn_statemachine_exists "$statemachine_arn"; then
    echo "FAIL: expected state machine $statemachine_name to have been created in Step Functions"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Deleting state machine $statemachine_resource_name resource from Kubernetes"
kubectl delete "$statemachine_resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

# Deleting can take a little while.
sleep 90

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Checking state machine $statemachine_arn no longer exists"
if sfn_statemachine_exists "$statemachine_arn"; then
    echo "FAIL: expected state machine $statemachine_name to be deleted in Step Functions"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Cleaning up execution role in IAM"
sfn_clean_up_iam_resources_for_executionrole $sfn_executionrole_name



######## ACTIVITIES ########

activity_name="ack-test-smoke-$service_name"
activity_resource_name="activities/$activity_name"
activity_arn="arn:aws:states:$AWS_REGION:$AWS_ACCOUNT_ID:activity:$activity_name"

debug_msg ""
debug_msg "Resources for Step Functions activity smoke test:"
debug_msg "activity_name: $activity_name"
debug_msg "activity_resource_name: $activity_resource_name"
debug_msg "activity_arn: $activity_arn"
debug_msg ""

# PRE-CHECKS
if sfn_activity_exists "activity_arn"; then
    echo "FAIL: expected activity $activity_name to not exist in Step Functions. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "activity_resource_name"; then
    echo "FAIL: expected activity_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS
currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Creating activity $activity_name"
cat <<EOF | kubectl apply -f -
apiVersion: sfn.services.k8s.aws/v1alpha1
kind: Activity
metadata:
  name: $activity_name
spec:
  name: $activity_name
EOF

sleep 5

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Checking activity $activity_arn created in Step Functions"
if ! sfn_activity_exists "$activity_arn"; then
    echo "FAIL: expected activity $activity_name to have been created in Step Functions"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Deleting activity $activity_resource_name resource from Kubernetes"
kubectl delete "$activity_resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

sleep 5

currtime=`date +%Y-%m-%dT%H:%M:%S%z`
debug_msg "$currtime: Checking activity $activity_arn no longer exists"
if sfn_activity_exists "$activity_arn"; then
    echo "FAIL: expected activity $activity_name to be deleted in Step Functions"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

assert_pod_not_restarted $ack_ctrl_pod_id
