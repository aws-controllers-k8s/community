#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/aws/s3.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

# AWS_ACCOUNT_ID_ALT should be configured to allow cross account resources management
# from AWS_ACCOUNT_ID
AWS_ACCOUNT_ID_ALT=${AWS_ACCOUNT_ID_ALT:-""}
AWS_PROFILE_ALT=${AWS_PROFILE_ALT:-""}
AWS_REGION_ALT="eu-west-2"
TESTING_NAMESPACE="testing-$RANDOM"
ASSUME_POLICY_ARN=${ASSUME_POLICY_ARN:="s3FullAccess"}

if [[ -z "$AWS_ACCOUNT_ID_ALT" ]] || [[ -z "$AWS_REGION_ALT" ]] || [[ -z "$AWS_PROFILE_ALT" ]]; then
    echo "skipping cross-account tests due to missing credentials"
    exit 0
fi

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="s3"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

bucket_name="ack-test-smoke-$service_name-$AWS_ACCOUNT_ID_ALT-$RANDOM"
resource_name="buckets/$bucket_name"

# Create the testing namespace.
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: $TESTING_NAMESPACE
  annotations:
    services.k8s.aws/default-region: "$AWS_REGION_ALT"
    services.k8s.aws/owner-account-id: "$AWS_ACCOUNT_ID_ALT"
EOF

# Create the ack-role-account-map ConfigMap.
# See CARM design: https://github.com/aws/aws-controllers-k8s/blob/main/docs/design/proposals/carm/cross-account-resource-management.md#storing-aws-role-arns
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: ack-role-account-map
  namespace: ack-system
data:
  "$AWS_ACCOUNT_ID_ALT": arn:aws:iam::$AWS_ACCOUNT_ID_ALT:role/$ASSUME_POLICY_ARN
EOF

sleep 5

# PRE-CHECKS
if s3_bucket_exists "$bucket_name" "$AWS_REGION_ALT" "$AWS_PROFILE_ALT"; then
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

debug_msg "checking bucket $bucket_name created in S3, in region $AWS_REGION_ALT"
if ! s3_bucket_exists "$bucket_name" "$AWS_REGION_ALT" "$AWS_PROFILE_ALT"; then
    echo "FAIL: expected $bucket_name to have been created in S3"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

list_bucket_json
if s3_bucket_exists "$bucket_name" "$AWS_REGION_ALT" "$AWS_PROFILE_ALT"; then
    echo "FAIL: expected $bucket_name to be deleted in S3"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

# Delete the testing namespace and ack-role-account-map
kubectl delete namespace $TESTING_NAMESPACE
kubectl delete configmap -n ack-system ack-role-account-map

assert_pod_not_restarted $ack_ctrl_pod_id
