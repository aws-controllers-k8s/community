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
	"github.com/gertd/go-pluralize"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

type CRDOps struct {
	Create  *openapi3.Operation
	ReadOne *openapi3.Operation
	Update  *openapi3.Operation
	Delete  *openapi3.Operation
}

type CRD struct {
	Names  names.Names
	Kind   string
	Plural string
	// SDKObjectType is the string name of the struct type used to return a
	// Describe operation for this resource from the aws-sdk-go. This is
	// typically called "{Resource}Data", where "{Resource}" is the name of the
	// resource. For example, the AppMesh API calls this struct MeshData.
	SDKObjectType string
	Ops           CRDOps
	SpecAttrs     map[string]*Attr
	StatusAttrs   map[string]*Attr
}

func NewCRD(
	names names.Names,
	sdkObjType string,
	crdOps CRDOps,
) *CRD {
	pluralize := pluralize.NewClient()
	kind := names.Camel
	plural := pluralize.Plural(kind)
	return &CRD{
		Names:         names,
		Kind:          kind,
		Plural:        plural,
		SDKObjectType: sdkObjType,
		Ops:           crdOps,
	}
}
