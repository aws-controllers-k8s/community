#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

check_is_installed jq

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="apigatewayv2"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

api_name="ack-test-$service_name-api"
resource_name="api/$api_name"

#PRE-CHECKS
## httpapi-sample api should not be existing.
if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS

## create httpapi-sample api resource
cat <<EOF | kubectl apply -f -
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: API
metadata:
  name: $api_name
spec:
  name: $api_name
  protocolType: HTTP
EOF

sleep 10

## validate that api-id was populated in resource status
debug_msg "retrieve api-id from $resource_name resource's status"
api_id=$(kubectl get $resource_name -o=json | jq -r .status.apiID)

if [[ -z "$api_id" ]];then
	echo "FAIL: $resource_name resource's status does not have apiID"
	exit 1
fi

## validate that api was created using apigatewayv2 get-api operation
debug_msg "apigatewayv2 get-api with api-id $api_id"
aws apigatewayv2 get-api --api-id=$api_id > /dev/null 2>&1
assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-api --api-id=$api_id' but got $?" || exit 1

## delete resource
debug_msg "delete $resource_name resource"
kubectl delete $resource_name >/dev/null 2>&1
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

## validate that api was deleted using apigatewayv2 get-api operation
debug_msg "get-api with api-id $api_id"
aws apigatewayv2 get-api --api-id=$api_id > /dev/null 2>&1
assert_equal "254" "$?" "Expected not-found status code from 'apigatewayv2 get-api --api-id=$api_id' but got $?" || exit 1

echo "Successful smoke test"

assert_pod_not_restarted $ack_ctrl_pod_id