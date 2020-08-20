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

	"github.com/aws/aws-controllers-k8s/pkg/names"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
)

type Helper struct {
	sdkAPI      *awssdkmodel.API
	crds        []*CRD
	typeDefs    []*TypeDef
	typeImports map[string]string
	typeRenames map[string]string
	// A map of operation type and resource name to
	// aws-sdk-go/private/model/api.Operation structs
	opMap *OperationMap
	// Instructions to the code generator how to handle the API and its
	// resources
	generatorConfig *GeneratorConfig
}

func (h *Helper) GetServiceAlias() string {
	if h.sdkAPI == nil {
		return ""
	}
	return awssdkmodel.ServiceID(h.sdkAPI)
}

func (h *Helper) GetCleanServiceAlias() string {
	serviceAlias := strings.ToLower(h.GetServiceAlias())
	return strings.Replace(serviceAlias, " ", "", -1)
}

func (h *Helper) GetServiceFullName() string {
	if h.sdkAPI == nil {
		return ""
	}
	return h.sdkAPI.Metadata.ServiceFullName
}

func (h *Helper) GetAPIGroup() string {
	cleanServiceAlias := h.GetCleanServiceAlias()
	return fmt.Sprintf("%s.services.k8s.aws", cleanServiceAlias)
}

func (h *Helper) GetCRDNames() []names.Names {
	opMap := h.GetOperationMap()
	createOps := (*opMap)[OpTypeCreate]
	crdNames := []names.Names{}
	for crdName := range createOps {
		if h.IsIgnoredResource(crdName) {
			continue
		}
		crdNames = append(crdNames, names.New(crdName))
	}
	return crdNames
}

// GetTypeRenames returns a map of original type name to renamed name (some
// type definition names conflict with generated names)
func (h *Helper) GetTypeRenames() map[string]string {
	_, _, _ = h.GetTypeDefs()
	return h.typeRenames
}

// NewHelper returns a new Helper struct for a supplied API model. Optionally,
// pass a file path to a generator config file that can be used to instruct the
// code generator how to handle the API properly
func NewHelper(
	sdkAPI *awssdkmodel.API,
	configPath string,
) (*Helper, error) {
	// If we don't do this, we can end up with panic()'s like this:
	// panic: assignment to entry in nil map
	// when trying to execute Shape.GoType().
	//
	// Calling API.ServicePackageDoc() ends up resetting the API.imports
	// unexported map variable...
	_ = sdkAPI.ServicePackageDoc()

	var gc *GeneratorConfig
	var err error
	if configPath != "" {
		gc, err = NewGeneratorConfig(configPath)
		if err != nil {
			return nil, err
		}
	}

	return &Helper{sdkAPI, nil, nil, nil, nil, nil, gc}, nil
}

// GetSDKAPIInterfaceTypeName returns the name of the aws-sdk-go primary API
// interface type name.
func (h *Helper) GetSDKAPIInterfaceTypeName() string {
	if h.sdkAPI == nil {
		return ""
	}
	return h.sdkAPI.StructName()
}
