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
key_name="ack-test-smoke-$service_name-$AWS_ACCOUNT_ID"
resource_name="$key_name"
key_description="this little key"
user=$(echo $ACK_TEST_IAM_ROLE|tr '[:upper:]' '[:lower:]')

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
spec:
  description: $key_description 
EOF

sleep 1

key_id=$(get_field_from_status key/$resource_name 'keyID')

debug_msg "checking key $key_name created in $service_name has field description with value $key_description"
aws_return=$(daws kms describe-key --key-id $key_id | jq -r .KeyMetadata.Description)
assert_equal "$key_description" "$aws_return" "Expected $key_description to be configured and set to $aws_description" || exit 1

cat <<EOF | kubectl apply -f -
apiVersion: kms.services.k8s.aws/v1alpha1
kind: Grant
metadata:
  name: $user-$key_name
spec:
  keyID: $key_id
  granteePrincipal: $user
  operations:
    - Encrypt
    - Decrypt
EOF

sleep 1

debug_msg "checking grant creation for user $user in $service_name for kms key $key_id"
aws_return=$(daws kms list-grants --key-id $key_id | jq -r '.Grants [] .GranteePrincipal')
assert_equal "$user" "$aws_return" "Expected $user to have access to kms grant" || exit 1

sleep 1

kubectl delete "key/$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

assert_pod_not_restarted $ack_ctrl_pod_id