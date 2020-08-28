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
	"github.com/aws/aws-controllers-k8s/pkg/names"
)

const (
	// ConflictingNameSuffix is appended to type names when they overlap with
	// well-known common struct names for things like a CRD itself, or its
	// Spec/Status subfield struct type name.
	ConflictingNameSuffix = "_SDK"
)

// TypeDef is a Go type definition for structs that are member fields of the
// Spec or Status structs in Custom Resource Definitions (CRDs).
type TypeDef struct {
	Names names.Names
	Attrs map[string]*Attr
}
