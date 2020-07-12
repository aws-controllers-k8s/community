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

import awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

type SDKMapper struct {
	CRD *CRD
	// A map, keyed by a pointer to the aws-sdk-go private/model/api.Operation
	// object (from which we can determine an output shape) and the name of the
	// field in that Operation's output shape that contains the primary
	// resource's ARN field. This mapping allows the generated code to read
	// different fields from different output shapes and place this information
	// consistently into the CRD's Status.ACKResourceMetadata.ARN field.
	primaryResourceARNOutputFieldMap map[*awssdkmodel.Operation]string
}

func (m *SDKMapper) SetPrimaryResourceARNField(
	op *awssdkmodel.Operation,
	fieldName string,
) {
	m.primaryResourceARNOutputFieldMap[op] = fieldName
}

func NewSDKMapper(
	crd *CRD,
) *SDKMapper {
	return &SDKMapper{
		CRD:                              crd,
		primaryResourceARNOutputFieldMap: map[*awssdkmodel.Operation]string{},
	}
}
