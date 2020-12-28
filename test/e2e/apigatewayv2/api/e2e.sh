#!/usr/bin/env bash

set -u

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/apigatewayv2.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="apigatewayv2"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

api_name="ack-test-$service_name-api"
api_resource_name="api/$api_name"
import_api_name="import-$api_name"
import_api_resource_name="api/$import_api_name"
integration_name="ack-test-$service_name-integration"
integration_resource_name="integration/$integration_name"
route_name="ack-test-$service_name-route"
route_key="httpbin"
route_resource_name="route/$route_name"
stage_name="test"
stage_resource_name="stage/$stage_name"
authorizer_name="ack-test-$service_name-authorizer"
authorizer_resource_name="authorizer/$authorizer_name"
authorizer_role_name="ack-apigwv2-authorizer-role"
authorizer_function_name="ack-apigwv2-authorizer"

#PRE-CHECKS
if k8s_resource_exists "$stage_resource_name"; then
    echo "FATAL: expected $stage_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$route_resource_name"; then
    echo "FATAL: expected $route_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$integration_resource_name"; then
    echo "FATAL: expected $integration_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$authorizer_resource_name"; then
    echo "FATAL: expected $authorizer_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$import_api_resource_name"; then
    echo "FATAL: expected $import_api_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$api_resource_name"; then
    echo "FATAL: expected $api_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS
apigwv2_setup_iam_resources_for_authorizer "$authorizer_role_name"
sleep 5
apigwv2_create_lambda_authorizer "$authorizer_function_name" "$authorizer_role_name"

apigwv2_import_http_api_and_validate "$import_api_name"
apigwv2_reimport_http_api_and_validate "$import_api_name"
apigwv2_create_http_api_and_validate "$api_name"
apigwv2_update_http_api_and_validate "$api_name"
apigwv2_create_integration_and_validate "$api_name" "$integration_name"
apigwv2_create_authorizer_and_validate "$api_name" "$authorizer_name" "$authorizer_function_name"
apigwv2_create_route_and_validate "$api_name" "$route_name" "$route_key" "$integration_name" "$authorizer_name"
apigwv2_create_stage_and_validate "$api_name" "$stage_name"

# waiting 30 seconds for api configuration to deploy
sleep 30
apigwv2_perform_invocation_and_validate "$api_name" "$stage_name" "$route_key"

#cleanup
apigwv2_delete_stage_and_validate "$api_name" "$stage_name"
apigwv2_delete_route_and_validate "$api_name" "$route_name"
apigwv2_delete_integration_and_validate "$api_name" "$integration_name"
apigwv2_delete_authorizer_and_validate "$api_name" "$authorizer_name"
apigwv2_delete_http_api_and_validate "$api_name"
apigwv2_delete_http_api_and_validate "$import_api_name"

apigwv2_delete_authorizer_lambda "$authorizer_function_name"
apigwv2_clean_up_iam_resources_for_authorizer "$authorizer_role_name"

assert_pod_not_restarted "$ack_ctrl_pod_id"
