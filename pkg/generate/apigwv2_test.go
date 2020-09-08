// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	 http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package generate_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func TestAPIGatewayV2_GetTypeDefs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "apigatewayv2")

	tdefs, timports, err := g.GetTypeDefs()
	require.Nil(err)

	// APIGatewayV2 shapes have time.Time ("timestamp") types and so
	// GetTypeDefs() should return the special-cased with apimachinery/metav1
	// import, aliased as "metav1"
	expImports := map[string]string{"k8s.io/apimachinery/pkg/apis/meta/v1": "metav1"}
	assert.Equal(expImports, timports)

	// There is an "Api" Shape that is a struct that is an element of the
	// GetApis Operation. Its name conflicts with the CRD called API and thus
	// we need to check that its cleaned name is set to API_SDK (the _SDK
	// suffix is appended to the type name to avoid the conflict with
	// CRD-specific structs.
	tdef := getTypeDefByName("Api", tdefs)
	require.NotNil(tdef)

	assert.Equal("API_SDK", tdef.Names.Camel)
}

func TestAPIGatewayV2_Route(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "apigatewayv2")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Route", crds)
	require.NotNil(crd)

	assert.Equal("Route", crd.Names.Camel)
	assert.Equal("route", crd.Names.CamelLower)
	assert.Equal("route", crd.Names.Snake)

	// The GetRoute operation has the following definition:
	//
	//    "GetRoute" : {
	//      "name" : "GetRoute",
	//      "http" : {
	//        "method" : "GET",
	//        "requestUri" : "/v2/apis/{apiId}/routes/{routeId}",
	//        "responseCode" : 200
	//      },
	//      "input" : {
	//        "shape" : "GetRouteRequest"
	//      },
	//      "output" : {
	//        "shape" : "GetRouteResult"
	//      },
	//      "errors" : [ {
	//        "shape" : "NotFoundException"
	//      }, {
	//        "shape" : "TooManyRequestsException"
	//      } ]
	//    },
	//
	// Where the NotFoundException shape looks like this:
	//
	//    "NotFoundException" : {
	//      "type" : "structure",
	//      "members" : {
	//        "Message" : {
	//          "shape" : "__string",
	//          "locationName" : "message"
	//        },
	//        "ResourceType" : {
	//          "shape" : "__string",
	//          "locationName" : "resourceType"
	//        }
	//      },
	//      "exception" : true,
	//      "error" : {
	//        "httpStatusCode" : 404
	//      }
	//    },
	//
	// Which indicates that the error is a 404 and is our NotFoundException
	// code but the "code" attribute of the ErrorInfo struct is empty so
	// instead of returning a blank string, we need to use the name of the
	// shape itself...
	assert.Equal("NotFoundException", crd.ExceptionCode(404))

	// The APIGatewayV2 Route API has CRUD+L operations:
	//
	// * CreateRoute
	// * DeleteRoute
	// * UpdateRoute
	// * GetRoute
	// * GetRoutes
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.Update)
	assert.NotNil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)

	// And no separate get/set attributes calls.
	assert.Nil(crd.Ops.GetAttributes)
	assert.Nil(crd.Ops.SetAttributes)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"APIID",
		"APIKeyRequired",
		"AuthorizationScopes",
		"AuthorizationType",
		"AuthorizerID",
		"ModelSelectionExpression",
		"OperationName",
		"RequestModels",
		"RequestParameters",
		"RouteKey",
		"RouteResponseSelectionExpression",
		"Target",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		"APIGatewayManaged",
		"RouteID",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	expCreateInput := `
	if r.ko.Spec.APIID != nil {
		res.SetApiId(*r.ko.Spec.APIID)
	}
	if r.ko.Spec.APIKeyRequired != nil {
		res.SetApiKeyRequired(*r.ko.Spec.APIKeyRequired)
	}
	if r.ko.Spec.AuthorizationScopes != nil {
		f2 := []*string{}
		for _, f2iter := range r.ko.Spec.AuthorizationScopes {
			var f2elem string
			f2elem = *f2iter
			f2 = append(f2, &f2elem)
		}
		res.SetAuthorizationScopes(f2)
	}
	if r.ko.Spec.AuthorizationType != nil {
		res.SetAuthorizationType(*r.ko.Spec.AuthorizationType)
	}
	if r.ko.Spec.AuthorizerID != nil {
		res.SetAuthorizerId(*r.ko.Spec.AuthorizerID)
	}
	if r.ko.Spec.ModelSelectionExpression != nil {
		res.SetModelSelectionExpression(*r.ko.Spec.ModelSelectionExpression)
	}
	if r.ko.Spec.OperationName != nil {
		res.SetOperationName(*r.ko.Spec.OperationName)
	}
	if r.ko.Spec.RequestModels != nil {
		f7 := map[string]*string{}
		for f7key, f7valiter := range r.ko.Spec.RequestModels {
			var f7val string
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		res.SetRequestModels(f7)
	}
	if r.ko.Spec.RequestParameters != nil {
		f8 := map[string]*svcsdk.ParameterConstraints{}
		for f8key, f8valiter := range r.ko.Spec.RequestParameters {
			f8val := &svcsdk.ParameterConstraints{}
			if f8valiter.Required != nil {
				f8val.SetRequired(*f8valiter.Required)
			}
			f8[f8key] = f8val
		}
		res.SetRequestParameters(f8)
	}
	if r.ko.Spec.RouteKey != nil {
		res.SetRouteKey(*r.ko.Spec.RouteKey)
	}
	if r.ko.Spec.RouteResponseSelectionExpression != nil {
		res.SetRouteResponseSelectionExpression(*r.ko.Spec.RouteResponseSelectionExpression)
	}
	if r.ko.Spec.Target != nil {
		res.SetTarget(*r.ko.Spec.Target)
	}
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko", "res", 1))

	expCreateOutput := `
	if resp.ApiGatewayManaged != nil {
		ko.Status.APIGatewayManaged = resp.ApiGatewayManaged
	}
	if resp.RouteId != nil {
		ko.Status.RouteID = resp.RouteId
	}
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	expRequiredStatusFieldsMissingFromReadOneInput := `
	if r.ko.Status.RouteID == nil  {
		return true
	} else {
		return false
	}
`
	assert.Equal(expRequiredStatusFieldsMissingFromReadOneInput, fmt.Sprintf("\n%s\n", crd.GoCodeRequiredStatusFieldsMissingFromReadOneInput("r.ko", 1)))
}
