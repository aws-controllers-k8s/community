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
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

// TypeDef is a Go type definition for a struct that is present in the
// definition of a Custom Resource Definition (CRD)
type TypeDef struct {
	Names names.Names
	Attrs map[string]*Attr
}
