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
	"sort"

	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
)

func attrCamelNames(fields map[string]*ackmodel.CRDField) []string {
	res := []string{}
	for _, attr := range fields {
		res = append(res, attr.Names.Camel)
	}
	sort.Strings(res)
	return res
}

func getCRDByName(name string, crds []*ackmodel.CRD) *ackmodel.CRD {
	for _, c := range crds {
		if c.Names.Original == name {
			return c
		}
	}
	return nil
}

func getTypeDefByName(name string, tdefs []*ackmodel.TypeDef) *ackmodel.TypeDef {
	for _, td := range tdefs {
		if td.Names.Original == name {
			return td
		}
	}
	return nil
}
