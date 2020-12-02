#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

###########################################
# API
###########################################
# apigwv2_create_http_api_and_validate creates an http-api and validates that api
# exists in AWS.
# create_http_api_and_validate accepts one required parameter api_name
apigwv2_create_http_api_and_validate() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_http_api_and_validate"
    echo "Usage: apigwv2_create_http_api_and_validate api_name"
    exit 1
  fi

  local __api_name="$1"
  local __api_resource_name=api/"$1"

  ## create api resource
  cat <<EOF | kubectl apply -f - >/dev/null
  apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
  kind: API
  metadata:
    name: "$__api_name"
  spec:
    name: "$__api_name"
    protocolType: HTTP
EOF

  sleep 10

  ## validate that api-id was populated in resource status
  debug_msg "retrieve api-id from $__api_resource_name resource's status"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")

  ## validate that api was created using apigatewayv2 get-api operation
  debug_msg "apigatewayv2 get-api with api-id $__api_id"
  daws apigatewayv2 get-api --api-id="$__api_id" > /dev/null
  assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-api --api-id=$__api_id' but got $?" || exit 1
}

# apigwv2_update_http_api_and_validate updates an http-api and validates that api
# got updated in AWS.
# update_http_api_and_validate accepts one required parameter api_name
apigwv2_update_http_api_and_validate() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_http_api_and_validate"
    echo "Usage: apigwv2_update_http_api_and_validate api_name"
    exit 1
  fi

  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __updated_name="$__api_name"V2

  ## create api resource
  cat <<EOF | kubectl apply -f - >/dev/null
  apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
  kind: API
  metadata:
    name: "$__api_name"
  spec:
    name: "$__updated_name"
    protocolType: HTTP
EOF

  sleep 10

  ## retreive api-id from resource status
  debug_msg "retrieve api-id from $__api_resource_name resource's status"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")

  ## validate that api name was updated using apigatewayv2 get-api operation
  debug_msg "apigatewayv2 get-api with api-id $__api_id"
  local __name_in_aws=$(daws apigatewayv2 get-api --api-id="$__api_id" | jq -r .Name)
  assert_equal "$__updated_name" "$__name_in_aws" "Expected api name to be updated to $__updated_name but got $__name_in_aws" || exit 1
}

# apigwv2_delete_http_api_and_validate deletes an http-api and validates that api
# does not exist anymore in AWS.
# delete_http_api_and_validate accepts one required parameter api_name
apigwv2_delete_http_api_and_validate() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to delete_http_api_and_validate"
    echo "Usage: apigwv2_delete_http_api_and_validate api_name"
    exit 1
  fi

  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  #delete api resource
  debug_msg "delete $__api_resource_name resource"
  kubectl delete "$__api_resource_name" >/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  #validate that api was deleted using apigatewayv2 get-api operation
  debug_msg "get-api with api-id $__api_id"
  daws apigatewayv2 get-api --api-id="$__api_id" > /dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected not-found status code from 'apigatewayv2 get-api --api-id=$__api_id' but got $__status"
    exit 1
  fi
}

###########################################
# INTEGRATION
###########################################
# apigwv2_create_integration_and_validate creates an http-api-integration and validates that integration
# is existing in AWS.
# create_integration_and_validate accepts two required parameters. api_name and integration_name
apigwv2_create_integration_and_validate() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_integration_and_validate"
    echo "Usage: apigwv2_create_integration_and_validate api_name integration_name"
    exit 1
  fi
  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  local __integration_name="$2"
  local __integration_resource_name=integration/"$2"

  ## create integration resource
  cat <<EOF | kubectl apply -f -  >/dev/null
  apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
  kind: Integration
  metadata:
    name: "$__integration_name"
  spec:
    apiID: "$__api_id"
    integrationType: HTTP_PROXY
    integrationURI: "https://httpbin.org/get"
    integrationMethod: GET
    payloadFormatVersion: "1.0"
EOF

  sleep 10

  ## validate that integration-id was populated in resource status
  debug_msg "retrieve integration-id from $__integration_resource_name resource's status"
  local __integration_id=$(get_field_from_status "$__integration_resource_name" "integrationID")

  ## validate that integration was created using apigatewayv2 get-integration operation
  debug_msg "apigatewayv2 get-integration with api-id $__api_id and integration_id $__integration_id"
  daws apigatewayv2 get-integration --api-id="$__api_id" --integration-id="$__integration_id" > /dev/null
  assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-integration --api-id=$__api_id --integration-id=$__integration_id' but got $?" || exit 1
}

# apigwv2_delete_integration_and_validate deletes an http-api-integration and validates that integration
# does not exist in AWS anymore.
# delete_integration_and_validate accepts two required parameters. api_name and integration_name
apigwv2_delete_integration_and_validate() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to delete_integration_and_validate"
    echo "Usage: apigwv2_delete_integration_and_validate api_name integration_name"
    exit 1
  fi
  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  local __integration_name="$2"
  local __integration_resource_name=integration/"$2"
  local __integration_id=$(get_field_from_status "$__integration_resource_name" "integrationID")

  #delete integration resource
  debug_msg "delete $__integration_resource_name resource"
  kubectl delete "$__integration_resource_name" >/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  #validate that integration was deleted using apigatewayv2 get-integration operation
  debug_msg "get-integration with api-id $__api_id and integration_id $__integration_id"
  daws apigatewayv2 get-integration --api-id="$__api_id" --integration-id="$__integration_id" >/dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected not-found status code from 'apigatewayv2 get-integration --api-id=$__api_id --integration-id=$__integration_id' but got $__status"
    exit 1
  fi
}

###########################################
# ROUTE
###########################################
# apigwv2_create_route_and_validate creates an http-api-route and validates that route
# is existing in AWS.
# create_integration_and_validate accepts four required parameters. api_name, route_name
# route_key, integration_name and authorizer_name
apigwv2_create_route_and_validate() {
  if [[ $# -ne 5 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_route_and_validate"
    echo "Usage: apigwv2_create_route_and_validate api_name route_name route_key integration_name authorizer_name"
    exit 1
  fi

  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")

  local __route_name="$2"
  local __route_resource_name=route/"$2"
  local __route_key="$3"

  local __integration_name="$4"
  local __integration_resource_name=integration/"$4"
  local __integration_id=$(get_field_from_status "$__integration_resource_name" "integrationID")

  local __authorizer_name="$5"
  local __authorizer_resource_name=authorizer/"$5"
  local __authorizer_id=$(get_field_from_status "$__authorizer_resource_name" "authorizerID")

  ## create route resource
  cat <<EOF | kubectl apply -f -  >/dev/null
  apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
  kind: Route
  metadata:
    name: "$__route_name"
  spec:
    apiID: "$__api_id"
    routeKey: "GET /$__route_key"
    target: integrations/$__integration_id
    authorizationType: CUSTOM
    authorizerID: "$__authorizer_id"
EOF

  sleep 10

  ## validate that route-id was populated in resource status
  debug_msg "retrieve route-id from $__route_resource_name resource's status"
  local __route_id=$(get_field_from_status "$__route_resource_name" "routeID")

  ## validate that route was created using apigatewayv2 get-route operation
  debug_msg "apigatewayv2 get-route with api-id $__api_id and route_id $__route_id"
  daws apigatewayv2 get-route --api-id="$__api_id" --route-id="$__route_id" > /dev/null
  assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-route --api-id=$__api_id --route-id=$__route_id' but got $?" || exit 1
}

# apigwv2_delete_route_and_validate deletes an http-api-route and validates that route
# is NOT existing in AWS anymore.
# delete_route_and_validate accepts two required parameters. api_name and route_name
apigwv2_delete_route_and_validate() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to delete_route_and_validate"
    echo "Usage: apigwv2_delete_route_and_validate api_name route_name"
    exit 1
  fi
  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  local __route_name="$2"
  local __route_resource_name=route/"$2"
  local __route_id=$(get_field_from_status "$__route_resource_name" "routeID")
  
  #delete route resource
  debug_msg "delete $__route_resource_name resource"
  kubectl delete "$__route_resource_name" >/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  #validate that route was deleted using apigatewayv2 get-route operation
  debug_msg "get-route with api-id $__api_id and route_id $__route_id"
  daws apigatewayv2 get-route --api-id="$__api_id" --route-id="$__route_id" >/dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected not-found status code from 'apigatewayv2 get-route --api-id=$__api_id --route-id=$__route_id' but got $__status"
    exit 1
  fi
}

###########################################
# STAGE
###########################################
# apigwv2_create_stage_and_validate creates an http-api-stage and validates that stage
# is existing in AWS.
# create_stage_and_validate accepts two required parameters. api_name and stage_name
apigwv2_create_stage_and_validate() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_stage_and_validate"
    echo "Usage: apigwv2_create_stage_and_validate api_name stage_name"
    exit 1
  fi
  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  local __stage_name="$2"
  local __stage_resource_name=stage/"$2"
  ## create stage resource
  cat <<EOF | kubectl apply -f -  >/dev/null
  apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
  kind: Stage
  metadata:
    name: "$__stage_name"
  spec:
    apiID: "$__api_id"
    stageName: "$__stage_name"
    autoDeploy: true
EOF

  sleep 10

  ## validate that stage was created using apigatewayv2 get-stage operation
  debug_msg "apigatewayv2 get-stage with api-id $__api_id and stage-name $__stage_name"
  daws apigatewayv2 get-stage --api-id="$__api_id" --stage-name="$__stage_name" > /dev/null
  assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-stage --api-id=$__api_id --stage-name=$__stage_name' but got $?" || exit 1
}

# apigwv2_delete_stage_and_validate deletes an http-api-stage and validates that stage
# does not exist in AWS anymore.
# delete_stage_and_validate accepts two required parameters. api_name and stage_name
apigwv2_delete_stage_and_validate() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_stage_and_validate"
    echo "Usage: apigwv2_delete_stage_and_validate api_name stage_name"
    exit 1
  fi
  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  local __stage_name="$2"
  local __stage_resource_name=stage/"$2"

  #delete stage resource
  debug_msg "delete $__stage_resource_name resource"
  kubectl delete "$__stage_resource_name" >/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  #validate that stage was deleted using apigatewayv2 get-stage operation
  debug_msg "get-stage with api-id $__api_id and stage_name $__stage_name"
  daws apigatewayv2 get-stage --api-id="$__api_id" --stage-name="$__stage_name" >/dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected not-found status code from 'apigatewayv2 get-stage --api-id=$__api_id --stage-name=$__stage_name' but got $__status"
    exit 1
  fi
}

###########################################
# AUTHORIZER
###########################################
# apigwv2_setup_iam_resources_for_authorizer creates an iam-role and attaches AWSLambdaBasicExecutionRole policy to it.
# apigwv2_setup_iam_resources_for_authorizer accepts only one required parameter role_name
apigwv2_setup_iam_resources_for_authorizer() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to setup_iam_resources_for_authorizer"
    echo "Usage: apigwv2_setup_iam_resources_for_authorizer role_name"
    exit 1
  fi

  local __role_name="$1"

  daws iam get-role --role-name "$__role_name" >/dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected IAM role $__role_name to not exist. Did previous test run cleanup?"
    exit 1
  fi

  daws iam create-role --role-name "$__role_name" --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam create-role --role-name $__role_name but got $?" || exit 1
  daws iam attach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam attach-role-policy --role-name $__role_name but got $?" || exit 1
}

# apigwv2_create_lambda_authorizer creates a lambda function that will be used as authorizer for http-api
# apigwv2_create_lambda_authorizer accepts two one required parameter function_name execution_role_name
apigwv2_create_lambda_authorizer() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_lambda_authorizer"
    echo "Usage: apigwv2_create_lambda_authorizer function_name execution_role_name"
    exit 1
  fi

  local __function_name="$1"
  local __role_name="$2"
  local __role_arn=$(daws iam get-role --role-name "$__role_name" | jq -r ".Role.Arn")
  if [[ -z "$__role_arn" ]];then
    echo "FATAL: Expected iam role $__role_name to exist."
    exit 1
  fi

  #Authorizer code
  cat <<EOF > index.js
exports.handler = async(event) => {
    let response = {
        "isAuthorized": false,
        "context": {
            "stringKey": "value",
            "numberKey": 1,
            "booleanKey": true,
            "arrayKey": ["value1", "value2"],
            "mapKey": {"value1": "value2"}
        }
    };

    if (event.headers.authorization === "SecretToken") {
        response = {
            "isAuthorized": true,
            "context": {
                "stringKey": "value",
                "numberKey": 1,
                "booleanKey": true,
                "arrayKey": ["value1", "value2"],
                "mapKey": {"value1": "value2"}
            }
        };
    }

    return response;
};
EOF
  zip authorizer.zip index.js >/dev/null

  daws lambda get-function --function-name "$__function_name" >/dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected lambda function $__function_name to not exist. Did previous test run cleanup?"
    exit 1
  fi

  daws lambda create-function --function-name "$__function_name" --runtime nodejs12.x --role "$__role_arn" --handler index.handler --zip-file fileb://authorizer.zip > /dev/null
  assert_equal "0" "$?" "Expected success from aws lambda create-function --function-name $__function_name but got $?" || exit 1

  #delete redundant files
  rm index.js
  rm authorizer.zip
}

# apigwv2_create_authorizer_and_validate creates an http-api-authorizer and validates that authorizer
# is existing in AWS.
# apigwv2_create_authorizer_and_validate accepts three required parameters. api_name, authorizer_name and lambda_function_name
apigwv2_create_authorizer_and_validate() {
  if [[ $# -ne 3 ]]; then
    echo "FATAL: Wrong number of arguments passed to create_lambda_authorizer"
    echo "Usage: apigwv2_create_authorizer_and_validate api_name authorizer_name function_name"
    exit 1
  fi

  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")

  local __authorizer_name="$2"
  local __authorizer_resource_name=authorizer/"$2"

  local __function_name="$3"
  local __function_arn=$(daws lambda get-function --function-name "$__function_name" | jq -r ".Configuration.FunctionArn")
  if [[ -z "$__function_arn" ]]; then
    echo "FATAL: Expected lambda function $__function_name to exist"
    exit 1
  fi
  local __region=$(echo "$__function_arn" | cut -d':' -f4)
  local __account=$(echo "$__function_arn" | cut -d':' -f5)

  local __authorizer_uri=arn:aws:apigateway:"$__region":lambda:path/2015-03-31/functions/arn:aws:lambda:"$__region":"$__account":function:"$__function_name"/invocations
  local __identitySource='$request.header.Authorization'

  #create-resource
  cat <<EOF | kubectl apply -f -  >/dev/null
  apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
  kind: Authorizer
  metadata:
    name: "$__authorizer_name"
  spec:
    apiID: "$__api_id"
    authorizerType: REQUEST
    identitySource:
      - "$__identitySource"
    name: "$__authorizer_name"
    authorizerURI: "$__authorizer_uri"
    authorizerPayloadFormatVersion: '2.0'
    enableSimpleResponses: true
EOF

  sleep 10

  ## validate that authorizer-id was populated in resource status
  debug_msg "retrieve authorizer-id from $__authorizer_resource_name resource's status"
  local __authorizer_id=$(get_field_from_status "$__authorizer_resource_name" "authorizerID")

  ## validate that authorizer was created using apigatewayv2 get-authorizer operation
  debug_msg "apigatewayv2 get-authorizer with api-id $__api_id and authorizer_id $__authorizer_id"
  daws apigatewayv2 get-authorizer --api-id="$__api_id" --authorizer-id="$__authorizer_id" > /dev/null
  assert_equal "0" "$?" "Expected success from 'apigatewayv2 get-authorizer --api-id=$__api_id --authorizer-id=$__authorizer_id' but got $?" || exit 1

  daws lambda add-permission --function-name "$__function_name" --statement-id "apigatewayv2-authorizer-invoke-permissions" --action "lambda:InvokeFunction" --principal "apigateway.amazonaws.com" --source-arn "arn:aws:execute-api:$__region:$__account:$__api_id/authorizers/$__authorizer_id" >/dev/null
  assert_equal "0" "$?" "Expected success from lambda add-permission but got $?" || exit 1
}

# apigwv2_delete_authorizer_and_validate deletes an http-api-authorizer and validates that authorizer
# does not exist in AWS anymore.
# apigwv2_delete_authorizer_and_validate accepts two required parameters. api_name and authorizer_name
apigwv2_delete_authorizer_and_validate() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to delete_authorizer_and_validate"
    echo "Usage: apigwv2_delete_authorizer_and_validate api_name authorizer_name"
    exit 1
  fi
  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __api_id=$(get_field_from_status "$__api_resource_name" "apiID")
  local __authorizer_name="$2"
  local __authorizer_resource_name=authorizer/"$2"
  local __authorizer_id=$(get_field_from_status "$__authorizer_resource_name" "authorizerID")

  #delete authorizer resource
  debug_msg "delete $__authorizer_resource_name resource"
  kubectl delete "$__authorizer_resource_name" >/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

  #validate that authorizer was deleted using apigatewayv2 get-authorizer operation
  debug_msg "get-authorizer with api-id $__api_id and authorizer_id $__authorizer_id"
  daws apigatewayv2 get-authorizer --api-id="$__api_id" --authorizer-id="$__authorizer_id" >/dev/null 2>&1
  local __status=$?
  if [[ $__status -ne 255 && $__status -ne 254 ]]; then
    echo "FATAL: Expected not-found status code from 'apigatewayv2 get-authorizer --api-id=$__api_id --authorizer-id=$__authorizer_id' but got $__status"
    exit 1
  fi
}

# apigwv2_delete_authorizer_lambda deletes the lambda function created for http-api-authorizer.
# apigwv2_delete_authorizer_lambda accepts one required parameters. lambda_function_name
apigwv2_delete_authorizer_lambda() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to delete_authorizer_lambda"
    echo "Usage: apigwv2_delete_authorizer_lambda function_name"
    exit 1
  fi

  local __function_name="$1"
  daws lambda delete-function --function-name "$__function_name" >/dev/null
  assert_equal "0" "$?" "Expected success from lambda delete-function --function-name $__function_name but got $?" || exit 1
}

## apigwv2_clean_up_iam_resources_for_authorizer deletes the lambda execution role created for http-api-authorizer.
## apigwv2_clean_up_iam_resources_for_authorizer accepts one required parameters. iam_role_name
apigwv2_clean_up_iam_resources_for_authorizer() {
  local __role_name="$1"
  daws iam detach-role-policy --role-name "$__role_name" --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam detach-role-policy --role-name $__role_name but got $?" || exit 1

  daws iam delete-role --role-name "$__role_name" >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam delete-role --role-name $__role_name but got $?" || exit 1
}

###########################################
# INVOCATION
###########################################
# apigwv2_perform_invocation_and_validate invokes an http-api and validates the successful invocation
# apigwv2_perform_invocation_and_validate accepts three required parameters. api_name, stage_name and route_key
apigwv2_perform_invocation_and_validate() {
  if [[ $# -ne 3 ]]; then
    echo "FATAL: Wrong number of arguments passed to perform_invocation_and_validate"
    echo "Usage: apigwv2_perform_invocation_and_validate api_name stage_name route_key"
    exit 1
  fi

  local __api_name="$1"
  local __api_resource_name=api/"$1"
  local __stage_name="$2"
  local __route_key="$3"
  local __api_endpoint=$(get_field_from_status "$__api_resource_name" "apiEndpoint")
  local __invocation_endpoint="$__api_endpoint"/"$__stage_name"/"$__route_key"
  local __test_header="Authorization"
  local __input_value="SecretToken"
  local __output_value=$(curl -s -H "$__test_header: $__input_value" "$__invocation_endpoint" | jq -r .headers.$__test_header)
  if [[ $__input_value != $__output_value ]]; then
    echo "FATAL: expected invocation result: $__input_value but received; $__output_value"
    exit 1
  fi
}
