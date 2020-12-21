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

package config

// SourceFieldConfig instructs the code generator how to handle a field in the
// Resource's SpecFields/StatusFields collection that takes its value from an
// abnormal source -- in other words, not the Create operation's Input or
// Output shape.
//
// This additional field can source its value from a shape in a different API
// Operation entirely.
type SourceFieldConfig struct {
	// Operation refers to the ID of the API Operation where we will
	// determine the field's Go type.
	Operation string `json:"operation"`
	// Path refers to the field path of the member of the Input or Output
	// shape in the Operation identified by OperationID that we will take as
	// our additional spec/status field's value.
	Path string `json:"path"`
}

// FieldConfig contains instructions to the code generator about how
// to interpret the value of an Attribute and how to map it to a CRD's Spec or
// Status field
type FieldConfig struct {
	// IsReadOnly indicates the field's value can not be set by a Kubernetes
	// user; in other words, the field should go in the CR's Status struct
	IsReadOnly bool `json:"is_read_only"`
	// ContainsOwnerAccountID indicates the field contains the AWS Account ID
	// that owns the resource. This is a special field that we direct to
	// storage in the common `Status.ACKResourceMetadata.OwnerAccountID` field.
	ContainsOwnerAccountID bool `json:"contains_owner_account_id"`
	// From instructs the code generator that the value of the field should
	// be retrieved from the specified operation and member path
	From *SourceFieldConfig `json:"from,omitempty"`
}
