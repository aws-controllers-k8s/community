#!/usr/bin/env bash

set -u

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws_apigwv2_testutil.sh"

check_is_installed jq

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="apigatewayv2"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

api_name="ack-test-$service_name-api"
api_resource_name="api/$api_name"
integration_name="ack-test-$service_name-integration"
integration_resource_name="integration/$integration_name"
route_name="ack-test-$service_name-route"
route_key="httpbin"
route_resource_name="route/$route_name"
stage_name="test"
stage_resource_name="stage/$stage_name"

#PRE-CHECKS
if k8s_resource_exists "$stage_resource_name"; then
    echo "FAIL: expected $stage_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$route_resource_name"; then
    echo "FAIL: expected $route_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$integration_resource_name"; then
    echo "FAIL: expected $integration_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$api_resource_name"; then
    echo "FAIL: expected $api_resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# TEST ACTIONS and ASSERTIONS
create_http_api_and_validate
create_integration_and_validate
create_route_and_validate
create_stage_and_validate

#wait for api configuration to deploy
sleep 30
perform_invocation_and_validate

delete_stage_and_validate
delete_route_and_validate
delete_integration_and_validate
delete_http_api_and_validate

assert_pod_not_restarted "$ack_ctrl_pod_id"

echo "Successful apigatewayv2 e2e test"