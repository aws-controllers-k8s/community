# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the API Gateway V2
"""

import logging
import time

import boto3
import pytest
import requests
import random
import string

from apigatewayv2 import SERVICE_NAME, service_marker
from apigatewayv2.bootstrap_resources import get_bootstrap_resources
from apigatewayv2.replacement_values import REPLACEMENT_VALUES
import apigatewayv2.tests.helper as helper
from apigatewayv2.tests.helper import ApiGatewayValidator
from common import k8s
from common.aws import get_aws_region, get_aws_account_id
from common.resources import load_resource_file

DELETE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
APIGW_DEPLOYMENT_WAIT_AFTER_SECONDS = 10

apigw_validator = ApiGatewayValidator(boto3.client('apigatewayv2'))
test_resource_values = REPLACEMENT_VALUES.copy()


@pytest.fixture(scope="module")
def api_resource():
    api_resource_name = test_resource_values['API_NAME']
    api_ref, api_data = helper.api_ref_and_data(api_resource_name=api_resource_name,
                                                replacement_values=test_resource_values)
    if k8s.get_resource_exists(api_ref):
        raise Exception(f"expected {api_resource_name} to not exist. Did previous test cleanup?")
    logging.debug(f"http api resource. name: {api_resource_name}, data: {api_data}")

    k8s.create_custom_resource(api_ref, api_data)
    cr = k8s.wait_resource_consumed_by_controller(api_ref)

    assert cr is not None
    assert k8s.get_resource_exists(api_ref)

    api_id = cr['status']['apiID']
    test_resource_values['API_ID'] = api_id

    yield api_ref, cr

    k8s.delete_custom_resource(api_ref)


@pytest.fixture(scope="module")
def integration_resource(api_resource):
    integration_resource_name = test_resource_values['INTEGRATION_NAME']
    integration_ref, integration_data = helper.integration_ref_and_data(
        integration_resource_name=integration_resource_name,
        replacement_values=test_resource_values)
    if k8s.get_resource_exists(integration_ref):
        raise Exception(f"expected {integration_resource_name} to not exist. Did previous test cleanup?")
    logging.debug(f"apigatewayv2 integration resource. name: {integration_resource_name}, data: {integration_data}")

    k8s.create_custom_resource(integration_ref, integration_data)
    cr = k8s.wait_resource_consumed_by_controller(integration_ref)

    assert cr is not None
    assert k8s.get_resource_exists(integration_ref)

    integration_id = cr['status']['integrationID']
    test_resource_values['INTEGRATION_ID'] = integration_id

    yield integration_ref, cr

    k8s.delete_custom_resource(integration_ref)


@pytest.fixture(scope="module")
def authorizer_resource(api_resource):
    authorizer_resource_name = test_resource_values['AUTHORIZER_NAME']
    authorizer_uri = f'arn:aws:apigateway:{get_aws_region()}:lambda:path/2015-03-31/functions/{get_bootstrap_resources().AuthorizerFunctionArn}/invocations'
    test_resource_values["AUTHORIZER_URI"] = authorizer_uri
    authorizer_ref, authorizer_data = helper.authorizer_ref_and_data(authorizer_resource_name=authorizer_resource_name,
                                                                     replacement_values=test_resource_values)
    if k8s.get_resource_exists(authorizer_ref):
        raise Exception(f"expected {authorizer_resource_name} to not exist. Did previous test cleanup?")
    logging.debug(f"apigatewayv2 authorizer resource. name: {authorizer_resource_name}, data: {authorizer_data}")

    k8s.create_custom_resource(authorizer_ref, authorizer_data)
    cr = k8s.wait_resource_consumed_by_controller(authorizer_ref)

    assert cr is not None
    assert k8s.get_resource_exists(authorizer_ref)

    authorizer_id = cr['status']['authorizerID']
    test_resource_values['AUTHORIZER_ID'] = authorizer_id

    # add permissions for apigateway to invoke authorizer lambda
    authorizer_arn = "arn:aws:execute-api:{region}:{account}:{api_id}/authorizers/{authorizer_id}".format(
        region=get_aws_region(),
        account=get_aws_account_id(),
        api_id=test_resource_values['API_ID'],
        authorizer_id=authorizer_id
    )
    lambda_client = boto3.client("lambda")
    lambda_client.add_permission(FunctionName=get_bootstrap_resources().AuthorizerFunctionName,
                                 StatementId='apigatewayv2-authorizer-invoke-permissions',
                                 Action='lambda:InvokeFunction',
                                 Principal='apigateway.amazonaws.com',
                                 SourceArn=authorizer_arn)

    yield authorizer_ref, cr

    k8s.delete_custom_resource(authorizer_ref)


@pytest.fixture(scope="module")
def route_resource(integration_resource, authorizer_resource):
    route_resource_name = test_resource_values['ROUTE_NAME']
    route_ref, route_data = helper.route_ref_and_data(route_resource_name=route_resource_name,
                                                      replacement_values=test_resource_values)
    if k8s.get_resource_exists(route_ref):
        raise Exception(f"expected {route_resource_name} to not exist. Did previous test cleanup?")
    logging.debug(f"apigatewayv2 route resource. name: {route_resource_name}, data: {route_data}")

    k8s.create_custom_resource(route_ref, route_data)
    cr = k8s.wait_resource_consumed_by_controller(route_ref)

    assert cr is not None
    assert k8s.get_resource_exists(route_ref)

    route_id = cr['status']['routeID']
    test_resource_values['ROUTE_ID'] = route_id

    yield route_ref, cr

    k8s.delete_custom_resource(route_ref)


@pytest.fixture(scope="module")
def stage_resource(route_resource):
    stage_resource_name = test_resource_values['STAGE_NAME']
    stage_ref, stage_data = helper.stage_ref_and_data(stage_resource_name=stage_resource_name,
                                                      replacement_values=test_resource_values)
    if k8s.get_resource_exists(stage_ref):
        raise Exception(f"expected {stage_resource_name} to not exist. Did previous test cleanup?")
    logging.debug(f"apigatewayv2 stage resource. name: {stage_resource_name}, data: {stage_data}")

    k8s.create_custom_resource(stage_ref, stage_data)
    cr = k8s.wait_resource_consumed_by_controller(stage_ref)

    assert cr is not None
    assert k8s.get_resource_exists(stage_ref)

    yield stage_ref, cr

    k8s.delete_custom_resource(stage_ref)


@service_marker
@pytest.mark.canary
class TestApiGatewayV2:

    def test_perform_invocation(self, api_resource, stage_resource):
        time.sleep(APIGW_DEPLOYMENT_WAIT_AFTER_SECONDS)
        api_ref, api_cr = api_resource
        api_endpoint = api_cr['status']['apiEndpoint']
        invoke_url = "{api_endpoint}/{stage_name}/{route_path}" \
            .format(api_endpoint=api_endpoint, stage_name=test_resource_values['STAGE_NAME'],
                    route_path=test_resource_values['ROUTE_PATH']
                    )
        response = requests.request(method='GET', url=invoke_url, headers={'Authorization': 'SecretToken'})
        assert 200 == response.status_code

    def test_crud_httpapi(self):
        test_data = REPLACEMENT_VALUES.copy()
        random_suffix = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
        api_name = "ack-test-httpapi-" + random_suffix
        test_data['API_NAME'] = api_name
        test_data['API_TITLE'] = api_name
        api_ref, api_data = helper.api_ref_and_data(api_resource_name=api_name,
                                                    replacement_values=test_data)
        logging.debug(f"http api resource. name: {api_name}, data: {api_data}")

        # test create
        k8s.create_custom_resource(api_ref, api_data)
        cr = k8s.wait_resource_consumed_by_controller(api_ref)

        assert cr is not None
        assert k8s.get_resource_exists(api_ref)

        api_id = cr['status']['apiID']

        # Let's check that the HTTP Api appears in Amazon API Gateway
        apigw_validator.assert_api_is_present(api_id=api_id)

        apigw_validator.assert_api_name(
            api_id=api_id,
            expected_api_name=api_name
        )

        # test update
        updated_api_title = 'updated-' + api_name
        test_data['API_TITLE'] = updated_api_title
        updated_api_resource_data = load_resource_file(
            SERVICE_NAME,
            "httpapi",
            additional_replacements=test_data,
        )
        logging.debug(f"updated http api resource: {updated_api_resource_data}")

        # Update the k8s resource
        k8s.patch_custom_resource(api_ref, updated_api_resource_data)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Let's check that the HTTP Api appears in Amazon API Gateway with updated title
        apigw_validator.assert_api_name(
            api_id=api_id,
            expected_api_name=updated_api_title
        )

        # test delete
        k8s.delete_custom_resource(api_ref)
        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        assert not k8s.get_resource_exists(api_ref)
        # HTTP Api should no longer appear in Amazon API Gateway
        apigw_validator.assert_api_is_deleted(api_id=api_id)

    def test_crud_httpapi_using_import(self):
        test_data = REPLACEMENT_VALUES.copy()
        random_suffix = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
        api_name = "ack-test-importapi-" + random_suffix
        test_data['API_NAME'] = api_name
        test_data['API_TITLE'] = api_name
        api_ref, api_data = helper.import_api_ref_and_data(api_resource_name=api_name,
                                                           replacement_values=test_data)
        logging.debug(f"imported http api resource. name: {api_name}, data: {api_data}")

        # test create
        k8s.create_custom_resource(api_ref, api_data)
        cr = k8s.wait_resource_consumed_by_controller(api_ref)

        assert cr is not None
        assert k8s.get_resource_exists(api_ref)

        api_id = cr['status']['apiID']

        # Let's check that the imported HTTP Api appears in Amazon API Gateway
        apigw_validator.assert_api_is_present(api_id=api_id)

        apigw_validator.assert_api_name(
            api_id=api_id,
            expected_api_name=api_name
        )

        # test update
        updated_api_title = 'updated-' + api_name
        test_data['API_TITLE'] = updated_api_title
        updated_api_resource_data = load_resource_file(
            SERVICE_NAME,
            "import_api",
            additional_replacements=test_data,
        )
        logging.debug(f"updated import http api resource: {updated_api_resource_data}")

        # Update the k8s resource
        k8s.patch_custom_resource(api_ref, updated_api_resource_data)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Let's check that the HTTP Api appears in Amazon API Gateway with updated title
        apigw_validator.assert_api_name(
            api_id=api_id,
            expected_api_name=updated_api_title
        )

        # test delete
        k8s.delete_custom_resource(api_ref)
        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        assert not k8s.get_resource_exists(api_ref)
        # HTTP Api should no longer appear in Amazon API Gateway
        apigw_validator.assert_api_is_deleted(api_id=api_id)

    def test_crud_integration(self, api_resource):
        api_ref, api_cr = api_resource
        api_id = api_cr['status']['apiID']
        test_data = REPLACEMENT_VALUES.copy()
        random_suffix = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
        integration_name = "ack-test-integration-" + random_suffix
        test_data['INTEGRATION_NAME'] = integration_name
        test_data['API_ID'] = api_id
        integration_ref, integration_data = helper.integration_ref_and_data(integration_resource_name=integration_name,
                                                                            replacement_values=test_data)
        logging.debug(f"http api integration resource. name: {integration_name}, data: {integration_data}")

        # test create
        k8s.create_custom_resource(integration_ref, integration_data)
        cr = k8s.wait_resource_consumed_by_controller(integration_ref)

        assert cr is not None
        assert k8s.get_resource_exists(integration_ref)

        integration_id = cr['status']['integrationID']

        # Let's check that the HTTP Api integration appears in Amazon API Gateway
        apigw_validator.assert_integration_is_present(api_id=api_id, integration_id=integration_id)

        apigw_validator.assert_integration_uri(
            api_id=api_id,
            integration_id=integration_id,
            expected_uri=test_data['INTEGRATION_URI']
        )

        # test update
        updated_uri = 'https://httpbin.org/post'
        test_data['INTEGRATION_URI'] = updated_uri
        updated_integration_resource_data = load_resource_file(
            SERVICE_NAME,
            "integration",
            additional_replacements=test_data,
        )
        logging.debug(f"updated http api integration resource: {updated_integration_resource_data}")

        # Update the k8s resource
        k8s.patch_custom_resource(integration_ref, updated_integration_resource_data)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Let's check that the HTTP Api integration appears in Amazon API Gateway with updated uri
        apigw_validator.assert_integration_uri(
            api_id=api_id,
            integration_id=integration_id,
            expected_uri=updated_uri
        )

        # test delete
        k8s.delete_custom_resource(integration_ref)
        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        assert not k8s.get_resource_exists(integration_ref)
        # HTTP Api integration should no longer appear in Amazon API Gateway
        apigw_validator.assert_integration_is_deleted(api_id=api_id, integration_id=integration_id)

    def test_crud_authorizer(self, api_resource):
        api_ref, api_cr = api_resource
        api_id = api_cr['status']['apiID']
        test_data = REPLACEMENT_VALUES.copy()
        random_suffix = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
        authorizer_name = "ack-test-authorizer-" + random_suffix
        test_data['AUTHORIZER_NAME'] = authorizer_name
        test_data['AUTHORIZER_TITLE'] = authorizer_name
        test_data['API_ID'] = api_id
        test_data['AUTHORIZER_URI'] = test_resource_values['AUTHORIZER_URI']
        authorizer_ref, authorizer_data = helper.authorizer_ref_and_data(authorizer_resource_name=authorizer_name,
                                                                         replacement_values=test_data)
        logging.debug(f"http api authorizer resource. name: {authorizer_name}, data: {authorizer_data}")

        # test create
        k8s.create_custom_resource(authorizer_ref, authorizer_data)
        cr = k8s.wait_resource_consumed_by_controller(authorizer_ref)

        assert cr is not None
        assert k8s.get_resource_exists(authorizer_ref)

        authorizer_id = cr['status']['authorizerID']

        # Let's check that the HTTP Api integration appears in Amazon API Gateway
        apigw_validator.assert_authorizer_is_present(api_id=api_id, authorizer_id=authorizer_id)

        apigw_validator.assert_authorizer_name(
            api_id=api_id,
            authorizer_id=authorizer_id,
            expected_authorizer_name=authorizer_name
        )

        # test update
        updated_authorizer_title = 'updated-' + authorizer_name
        test_data['AUTHORIZER_TITLE'] = updated_authorizer_title
        updated_authorizer_resource_data = load_resource_file(
            SERVICE_NAME,
            "authorizer",
            additional_replacements=test_data,
        )
        logging.debug(f"updated http api authorizer resource: {updated_authorizer_resource_data}")

        # Update the k8s resource
        k8s.patch_custom_resource(authorizer_ref, updated_authorizer_resource_data)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Let's check that the HTTP Api authorizer appears in Amazon API Gateway with updated title
        apigw_validator.assert_authorizer_name(
            api_id=api_id,
            authorizer_id=authorizer_id,
            expected_authorizer_name=updated_authorizer_title
        )

        # test delete
        k8s.delete_custom_resource(authorizer_ref)
        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        assert not k8s.get_resource_exists(authorizer_ref)
        # HTTP Api authorizer should no longer appear in Amazon API Gateway
        apigw_validator.assert_authorizer_is_deleted(api_id=api_id, authorizer_id=authorizer_id)

    def test_crud_route(self, api_resource, integration_resource, authorizer_resource):
        api_ref, api_cr = api_resource
        api_id = api_cr['status']['apiID']
        integration_ref, integration_cr = integration_resource
        integration_id = integration_cr['status']['integrationID']
        authorizer_ref, authorizer_cr = authorizer_resource
        authorizer_id = authorizer_cr['status']['authorizerID']
        test_data = REPLACEMENT_VALUES.copy()
        random_suffix = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
        route_name = "ack-test-route-" + random_suffix
        test_data['ROUTE_NAME'] = route_name
        test_data['AUTHORIZER_ID'] = authorizer_id
        test_data['INTEGRATION_ID'] = integration_id
        test_data['API_ID'] = api_id
        test_data['ROUTE_KEY'] = 'GET /httpbins'
        route_ref, route_data = helper.route_ref_and_data(route_resource_name=route_name,
                                                          replacement_values=test_data)
        logging.debug(f"http api route resource. name: {route_name}, data: {route_data}")

        # test create
        k8s.create_custom_resource(route_ref, route_data)
        cr = k8s.wait_resource_consumed_by_controller(route_ref)

        assert cr is not None
        assert k8s.get_resource_exists(route_ref)

        route_id = cr['status']['routeID']

        # Let's check that the HTTP Api route appears in Amazon API Gateway
        apigw_validator.assert_route_is_present(api_id=api_id, route_id=route_id)

        apigw_validator.assert_route_key(
            api_id=api_id,
            route_id=route_id,
            expected_route_key=test_data['ROUTE_KEY']
        )

        # test update
        updated_route_key = 'GET /uhttpbins'
        test_data['ROUTE_KEY'] = updated_route_key
        updated_route_resource_data = load_resource_file(
            SERVICE_NAME,
            "route",
            additional_replacements=test_data,
        )
        logging.debug(f"updated http api route resource: {updated_route_resource_data}")

        # Update the k8s resource
        k8s.patch_custom_resource(route_ref, updated_route_resource_data)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Let's check that the HTTP Api route appears in Amazon API Gateway with updated route key
        apigw_validator.assert_route_key(
            api_id=api_id,
            route_id=route_id,
            expected_route_key=updated_route_key
        )

        # test delete
        k8s.delete_custom_resource(route_ref)
        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        assert not k8s.get_resource_exists(route_ref)
        # HTTP Api route should no longer appear in Amazon API Gateway
        apigw_validator.assert_route_is_deleted(api_id=api_id, route_id=route_id)

    def test_crud_stage(self, api_resource):
        api_ref, api_cr = api_resource
        api_id = api_cr['status']['apiID']
        test_data = REPLACEMENT_VALUES.copy()
        random_suffix = (''.join(random.choice(string.ascii_lowercase) for _ in range(6)))
        stage_name = "ack-test-stage-" + random_suffix
        test_data['STAGE_NAME'] = stage_name
        test_data['API_ID'] = api_id
        stage_ref, stage_data = helper.stage_ref_and_data(stage_resource_name=stage_name,
                                                          replacement_values=test_data)
        logging.debug(f"http api stage resource. name: {stage_name}, data: {stage_data}")

        # test create
        k8s.create_custom_resource(stage_ref, stage_data)
        cr = k8s.wait_resource_consumed_by_controller(stage_ref)

        assert cr is not None
        assert k8s.get_resource_exists(stage_ref)

        # Let's check that the HTTP Api integration appears in Amazon API Gateway
        apigw_validator.assert_stage_is_present(api_id=api_id, stage_name=stage_name)

        stage_description = test_data['STAGE_DESCRIPTION']
        apigw_validator.assert_stage_description(
            api_id=api_id,
            stage_name=stage_name,
            expected_description=stage_description
        )

        # test update
        updated_description = 'updated' + stage_description
        test_data['STAGE_DESCRIPTION'] = updated_description
        updated_stage_resource_data = load_resource_file(
            SERVICE_NAME,
            "stage",
            additional_replacements=test_data,
        )
        logging.debug(f"updated http api stage resource: {updated_stage_resource_data}")

        # Update the k8s resource
        k8s.patch_custom_resource(stage_ref, updated_stage_resource_data)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Let's check that the HTTP Api stage appears in Amazon API Gateway with updated description
        apigw_validator.assert_stage_description(
            api_id=api_id,
            stage_name=stage_name,
            expected_description=updated_description
        )

        # test delete
        k8s.delete_custom_resource(stage_ref)
        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        assert not k8s.get_resource_exists(stage_ref)
        # HTTP Api stage should no longer appear in Amazon API Gateway
        apigw_validator.assert_stage_is_deleted(api_id=api_id, stage_name=stage_name)
