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

package model

import (
	"fmt"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
)

type Helper struct {
	sdkAPI      *awssdkmodel.API
	crds        []*CRD
	typeDefs    []*TypeDef
	typeImports map[string]string
	// A map of operation type and resource name to
	// aws-sdk-go/private/model/api.Operation structs
	opMap *OperationMap
}

func (h *Helper) GetServiceAlias() string {
	if h.sdkAPI == nil {
		return ""
	}
	return awssdkmodel.ServiceID(h.sdkAPI)
}

func (h *Helper) GetServiceFullName() string {
	if h.sdkAPI == nil {
		return ""
	}
	return h.sdkAPI.Metadata.ServiceFullName
}

func (h *Helper) GetAPIGroup() string {
	serviceAlias := strings.ToLower(h.GetServiceAlias())
	return fmt.Sprintf("%s.services.k8s.aws", serviceAlias)
}

func NewHelper(sdkAPI *awssdkmodel.API) *Helper {
	// If we don't do this, we can end up with panic()'s like this:
	// panic: assignment to entry in nil map
	// when trying to execute Shape.GoType().
	//
	// Calling API.ServicePackageDoc() ends up resetting the API.imports
	// unexported map variable...
	_ = sdkAPI.ServicePackageDoc()

	return &Helper{sdkAPI, nil, nil, nil, nil}
}

// GetSDKAPIInterfaceTypeName returns the name of the aws-sdk-go primary API
// interface type name.
func (h *Helper) GetSDKAPIInterfaceTypeName() string {
	if h.sdkAPI == nil {
		return ""
	}
	return h.sdkAPI.StructName()
}
