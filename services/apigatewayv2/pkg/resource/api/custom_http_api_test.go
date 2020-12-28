// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package api

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-controllers-k8s/pkg/metrics"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/apigatewayv2/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
	"github.com/stretchr/testify/assert"
)

type mockApiGatewayV2 struct {
	apigatewayv2iface.ApiGatewayV2API
	importApiOutput   *apigatewayv2.ImportApiOutput
	reimportApiOutput *apigatewayv2.ReimportApiOutput
	updateApiOutput   *apigatewayv2.UpdateApiOutput
}

func (i mockApiGatewayV2) ImportApiWithContext(aws.Context, *apigatewayv2.ImportApiInput, ...request.Option) (*apigatewayv2.ImportApiOutput, error) {
	return i.importApiOutput, nil
}

func (i mockApiGatewayV2) ReimportApiWithContext(aws.Context, *apigatewayv2.ReimportApiInput, ...request.Option) (*apigatewayv2.ReimportApiOutput, error) {
	return i.reimportApiOutput, nil
}

func (i mockApiGatewayV2) UpdateApiWithContext(aws.Context, *apigatewayv2.UpdateApiInput, ...request.Option) (*apigatewayv2.UpdateApiOutput, error) {
	return i.updateApiOutput, nil
}

// Helper methods to setup tests
// provideResourceManager returns pointer to resourceManager
func provideResourceManager(importApiOutput *apigatewayv2.ImportApiOutput, reimportApiOutput *apigatewayv2.ReimportApiOutput,
	updateApiOutput *apigatewayv2.UpdateApiOutput) *resourceManager {
	return &resourceManager{
		rr:           nil,
		awsAccountID: "",
		awsRegion:    "",
		sess:         nil,
		sdkapi:       mockApiGatewayV2{importApiOutput: importApiOutput, reimportApiOutput: reimportApiOutput, updateApiOutput: updateApiOutput},
		metrics:      metrics.NewMetrics("apigatewayv2"),
	}
}

// provideResource returns pointer to resource
func provideResource() *resource {
	return &resource{
		ko: &svcapitypes.API{},
	}
}

func Test_ImportApi_IncompatibleFieldsPresent(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	body := "body"
	name := "name"
	desired.ko.Spec = svcapitypes.APISpec{Body: &body, Name: &name}

	var ctx context.Context

	res, err := rm.customCreateApi(ctx, desired)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "only 'FailOnWarnings' and 'Basepath' fields can be used with 'Body' field"))
}

func Test_ImportApi_BodyFieldMissing(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	basepath := "basepath"
	failOnWarning := false
	desired.ko.Spec = svcapitypes.APISpec{Basepath: &basepath, FailOnWarnings: &failOnWarning}

	var ctx context.Context

	res, err := rm.customCreateApi(ctx, desired)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'FailOnWarnings' and 'Basepath' field(s) can only be used with 'Body' field for import-api operation"))

	desired.ko.Spec = svcapitypes.APISpec{Basepath: &basepath}
	res, err = rm.customCreateApi(ctx, desired)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'Basepath' field(s) can only be used with 'Body' field for import-api operation"))

	desired.ko.Spec = svcapitypes.APISpec{FailOnWarnings: &failOnWarning}
	res, err = rm.customCreateApi(ctx, desired)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'FailOnWarnings' field(s) can only be used with 'Body' field for import-api operation"))
}

func Test_ImportApi_Successful(t *testing.T) {
	assert := assert.New(t)
	// Setup
	apiId := "apiId"
	importApiOutput := apigatewayv2.ImportApiOutput{ApiId: &apiId}
	rm := provideResourceManager(&importApiOutput, nil, nil)

	desired := provideResource()
	body := "body"
	desired.ko.Spec = svcapitypes.APISpec{Body: &body}

	var ctx context.Context

	res, err := rm.customCreateApi(ctx, desired)
	assert.NotNil(res)
	assert.Nil(err)
	assert.Equal(apiId, *res.ko.Status.APIID)
}

func Test_CreateApi_MissingRequiredFields(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	desired.ko.Spec = svcapitypes.APISpec{}

	var ctx context.Context

	res, err := rm.customCreateApi(ctx, desired)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'Name' and 'ProtocolType' are required properties if 'Body' field is not present"))
}

func Test_CreateApi_HappyCase(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	name := "apiname"
	protocoltype := "protocolType"
	desired.ko.Spec = svcapitypes.APISpec{Name: &name, ProtocolType: &protocoltype}

	var ctx context.Context

	res, err := rm.customCreateApi(ctx, desired)
	assert.Nil(res)
	assert.Nil(err)
}

func Test_ReImportApi_IncompatibleFieldsPresent(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	body := "body"
	name := "name"
	desired.ko.Spec = svcapitypes.APISpec{Body: &body, Name: &name}

	var ctx context.Context

	res, err := rm.customUpdateApi(ctx, desired, nil, nil)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "only 'FailOnWarnings' and 'Basepath' fields can be used with 'Body' field"))
}

func Test_ReImportApi_BodyFieldMissing(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	basepath := "basepath"
	failOnWarning := false
	desired.ko.Spec = svcapitypes.APISpec{Basepath: &basepath, FailOnWarnings: &failOnWarning}

	var ctx context.Context

	res, err := rm.customUpdateApi(ctx, desired, nil, nil)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'FailOnWarnings' and 'Basepath' field(s) can only be used with 'Body' field for import-api operation"))

	desired.ko.Spec = svcapitypes.APISpec{Basepath: &basepath}
	res, err = rm.customUpdateApi(ctx, desired, nil, nil)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'Basepath' field(s) can only be used with 'Body' field for import-api operation"))

	desired.ko.Spec = svcapitypes.APISpec{FailOnWarnings: &failOnWarning}
	res, err = rm.customUpdateApi(ctx, desired, nil, nil)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'FailOnWarnings' field(s) can only be used with 'Body' field for import-api operation"))
}

func Test_ReImportApi_ApiIdMissing(t *testing.T) {
	assert := assert.New(t)
	// Setup
	apiId := "apiId"
	reimportApiOutput := apigatewayv2.ReimportApiOutput{ApiId: &apiId}
	rm := provideResourceManager(nil, &reimportApiOutput, nil)

	desired := provideResource()
	body := "body"
	desired.ko.Spec = svcapitypes.APISpec{Body: &body}

	var ctx context.Context

	res, err := rm.customUpdateApi(ctx, desired, nil, nil)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'APIID' is required input parameter for 'ReimportApi' operation"))
}

func Test_ReImportApi_Successful(t *testing.T) {
	assert := assert.New(t)
	// Setup
	apiId := "apiId"
	reimportApiOutput := apigatewayv2.ReimportApiOutput{ApiId: &apiId}
	rm := provideResourceManager(nil, &reimportApiOutput, nil)

	desired := provideResource()
	body := "body"
	desired.ko.Spec = svcapitypes.APISpec{Body: &body}
	desired.ko.Status = svcapitypes.APIStatus{APIID: &apiId}

	var ctx context.Context

	res, err := rm.customUpdateApi(ctx, desired, nil, nil)
	assert.NotNil(res)
	assert.Nil(err)
	assert.Equal(apiId, *res.ko.Status.APIID)
}

func Test_UpdateApi_MissingRequiredFields(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager(nil, nil, nil)

	desired := provideResource()
	desired.ko.Spec = svcapitypes.APISpec{}

	var ctx context.Context

	res, err := rm.customUpdateApi(ctx, desired, nil, nil)
	assert.Nil(res)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "'Name' and 'ProtocolType' are required properties if 'Body' field is not present"))

}

func Test_UpdateApi_HappyCase(t *testing.T) {
	assert := assert.New(t)
	// Setup
	apiId := "apiId"
	updateApiOutput := apigatewayv2.UpdateApiOutput{ApiId: &apiId}
	rm := provideResourceManager(nil, nil, &updateApiOutput)

	desired := provideResource()
	name := "name"
	protocolType := "HTTP"
	desired.ko.Spec = svcapitypes.APISpec{Name: &name, ProtocolType: &protocolType}

	var ctx context.Context

	res, err := rm.customUpdateApi(ctx, desired, nil, nil)
	assert.NotNil(res)
	assert.Nil(err)
	assert.Equal(apiId, *res.ko.Status.APIID)
}
