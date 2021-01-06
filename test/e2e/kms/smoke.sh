#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"
service_name="kms"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/aws/$service_name.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"
key_name=ack-test-smoke-kms-413689294325
#key_name="ack-test-smoke-$service_name-$AWS_ACCOUNT_ID"
resource_name="$key_name"

# PRE-CHECKS
if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

cat <<EOF | kubectl apply -f -
apiVersion: kms.services.k8s.aws/v1alpha1
kind: Key
metadata:
  name: $key_name
EOF

sleep 1

debug_msg "checking key $key_name created in $service_name"
key_id=$(get_field_from_status key/$resource_name 'keyID')
if ! kms_key_exists "$key_id"; then
    echo "FAIL: expected $key_name to have been created in $service_name"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

kubectl delete "key/$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

assert_pod_not_restarted $ack_ctrl_pod_id
