#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

check_is_installed jq

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="dynamodb"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

table_name="ack-test-smoke-$service_name"
resource_name="tables/$table_name"

get_table() {
    aws dynamodb describe-table --table-name "$table_name" --output json >/dev/null 2>&1
}

# PRE-CHECKS
get_table
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $table_name to not exist in DynamoDB. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

cat <<EOF | kubectl apply -f -
apiVersion: dynamodb.services.k8s.aws/v1alpha1
kind: Table
metadata:
  name: $table_name
spec:
  tableName: $table_name
  attributeDefinitions:
    - attributeName: ForumName
      attributeType: S
    - attributeName: Subject
      attributeType: S
    - attributeName: LastPostDateTime
      attributeType: S
  keySchema:
    - attributeName: ForumName
      keyType: HASH
    - attributeName: Subject
      keyType: RANGE
  localSecondaryIndexes:
    - indexName: LastPostIndex
      keySchema:
        - attributeName: ForumName
          keyType: HASH
        - attributeName: LastPostDateTime
          keyType: RANGE
      projection:
        projectionType: KEYS_ONLY
  provisionedThroughput:
    readCapacityUnits: 5
    writeCapacityUnits: 5
EOF

sleep 20

debug_msg "checking table $table_name created in DynamoDB"
get_table
if [[ $? -eq 255 || $? -eq 254 ]]; then
    echo "FAIL: expected $table_name to have been created in DynamoDB"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

# Table creation in DynamoDB is not fast. Need to sleep here for an extended
# period of time otherwise attempting to delete the table will result in the
# following error in the service controller:
#
# 2020-09-28T15:50:49.923Z	ERROR	controller-runtime.controller	Reconciler
# error	{"controller": "table", "request": "default/ack-test-smoke-dynamodb",
# "error": "ResourceInUseException: Attempt to change a resource which is still
# in use: Table is being created: ack-test-smoke-dynamodb"}

sleep 120

kubectl delete "$resource_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

sleep 60

get_table
if [[ $? -ne 255 && $? -ne 254 ]]; then
    echo "FAIL: expected $table_name to be deleted in DynamoDB"
    kubectl logs -n ack-system "$ack_ctrl_pod_id"
    exit 1
fi

assert_pod_not_restarted $ack_ctrl_pod_id
