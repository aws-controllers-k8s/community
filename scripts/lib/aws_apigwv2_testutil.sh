#!/usr/bin/env bash

###########################################
# API
###########################################
create_http_api_and_validate() {
## create api resource
cat <<EOF | kubectl apply -f - >/dev/null 2>&1
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
debug_msg "retrieve api-id from api/$api_name resource's status"
api_id=$(kubectl get $api_resource_name -o=json | jq -r .status.apiID)

if [[ -z "$api_id" ]];then
	echo "FAIL: $api_resource_name resource's status does not have apiID"
	exit 1
fi

## validate that api was created using apigatewayv2 get-api operation
debug_msg "apigatewayv2 get-api with api-id $api_id"
aws apigatewayv2 get-api --api-id="$api_id" > /dev/null 2>&1
assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-api --api-id=$api_id' but got $?" || exit 1
}

delete_http_api_and_validate() {
#delete api resource
debug_msg "delete api/$api_name resource"
kubectl delete $api_resource_name >/dev/null 2>&1
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

#validate that api was deleted using apigatewayv2 get-api operation
debug_msg "get-api with api-id $api_id"
aws apigatewayv2 get-api --api-id="$api_id" > /dev/null 2>&1
assert_equal "254" "$?" "Expected not-found status code from 'apigatewayv2 get-api --api-id=$api_id' but got $?" || exit 1
}

###########################################
# INTEGRATION
###########################################

create_integration_and_validate() {
## create integration resource
cat <<EOF | kubectl apply -f -  >/dev/null 2>&1
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Integration
metadata:
  name: $integration_name
spec:
  apiID: $api_id
  integrationType: HTTP_PROXY
  integrationURI: "https://httpbin.org/get"
  integrationMethod: GET
  payloadFormatVersion: "1.0"
EOF

sleep 10

## validate that integration-id was populated in resource status
debug_msg "retrieve integration-id from integration/$integration_name resource's status"
integration_id=$(kubectl get $integration_resource_name -o=json | jq -r .status.integrationID)

if [[ -z "$integration_id" ]];then
	echo "FAIL: $integration_resource_name resource's status does not have integrationID"
	exit 1
fi

## validate that integration was created using apigatewayv2 get-integration operation
debug_msg "apigatewayv2 get-integration with api-id $api_id and integration_id $integration_id"
aws apigatewayv2 get-integration --api-id="$api_id" --integration-id="$integration_id" > /dev/null 2>&1
assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-integration --api-id=$api_id --integration-id=$integration_id' but got $?" || exit 1
}

delete_integration_and_validate() {
#delete integration resource
debug_msg "delete $integration_resource_name resource"
kubectl delete $integration_resource_name >/dev/null 2>&1
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

#validate that integration was deleted using apigatewayv2 get-integration operation
debug_msg "get-integration with api-id $api_id and integration_id $integration_id"
aws apigatewayv2 get-integration --api-id="$api_id" --integration-id="$integration_id" >/dev/null 2>&1
assert_equal "254" "$?" "Expected not-found status code from 'apigatewayv2 get-integration --api-id=$api_id --integration-id=$integration_id' but got $?" || exit 1
}

###########################################
# ROUTE
###########################################
create_route_and_validate() {
## create route resource
cat <<EOF | kubectl apply -f -  >/dev/null 2>&1
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Route
metadata:
  name: $route_name
spec:
  apiID: $api_id
  routeKey: "GET /$route_key"
  target: integrations/$integration_id
EOF

sleep 10

## validate that route-id was populated in resource status
debug_msg "retrieve route-id from $route_resource_name resource's status"
route_id=$(kubectl get $route_resource_name -o=json | jq -r .status.routeID)

if [[ -z "$route_id" ]];then
	echo "FAIL: $route_resource_name resource's status does not have routeID"
	exit 1
fi

## validate that route was created using apigatewayv2 get-route operation
debug_msg "apigatewayv2 get-route with api-id $api_id and route_id $route_id"
aws apigatewayv2 get-route --api-id="$api_id" --route-id="$route_id" > /dev/null 2>&1
assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-route --api-id=$api_id --route-id=$route_id' but got $?" || exit 1
}

delete_route_and_validate() {
#delete route resource
debug_msg "delete $route_resource_name resource"
kubectl delete $route_resource_name >/dev/null 2>&1
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

#validate that route was deleted using apigatewayv2 get-route operation
debug_msg "get-route with api-id $api_id and route_id $route_id"
aws apigatewayv2 get-route --api-id="$api_id" --route-id="$route_id" >/dev/null 2>&1
assert_equal "254" "$?" "Expected not-found status code from 'apigatewayv2 get-route --api-id=$api_id --route-id=$route_id' but got $?" || exit 1
}

###########################################
# STAGE
###########################################
create_stage_and_validate() {
## create stage resource
cat <<EOF | kubectl apply -f -  >/dev/null 2>&1
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Stage
metadata:
  name: $stage_name
spec:
  apiID: $api_id
  stageName: $stage_name
  autoDeploy: true
EOF

sleep 10

## validate that stage was created using apigatewayv2 get-stage operation
debug_msg "apigatewayv2 get-stage with api-id $api_id and stage-name $stage_name"
aws apigatewayv2 get-stage --api-id="$api_id" --stage-name="$stage_name" > /dev/null 2>&1
assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-stage --api-id=$api_id --stage-name=$stage_name' but got $?" || exit 1
}

delete_stage_and_validate() {
#delete stage resource
debug_msg "delete $stage_resource_name resource"
kubectl delete $stage_resource_name >/dev/null 2>&1
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

#validate that stage was deleted using apigatewayv2 get-stage operation
debug_msg "get-stage with api-id $api_id and stage_name $stage_name"
aws apigatewayv2 get-stage --api-id="$api_id" --stage-name="$stage_name" >/dev/null 2>&1
assert_equal "254" "$?" "Expected not-found status code from 'apigatewayv2 get-stage --api-id=$api_id --stage-name=$stage_name' but got $?" || exit 1
}

###########################################
# INVOCATION
###########################################
perform_invocation_and_validate() {
local api_endpoint=$(kubectl get "$api_resource_name" -o=json | jq -r .status.apiEndpoint)
local invocation_endpoint=$api_endpoint/$stage_name/$route_key
local test_header="Testheader"
local input_value="InputValue"
local output_value=$(curl -s -H "$test_header: $input_value" "$invocation_endpoint" | jq -r .headers.$test_header)
if [[ $input_value != $output_value ]]; then
  echo "FAIL: expected invocation result: $input_value but received; $output_value"
  exit 1
fi
}
