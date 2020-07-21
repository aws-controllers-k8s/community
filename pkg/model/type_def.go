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
	"sort"
	"strings"

	"github.com/aws/aws-controllers-k8s/pkg/names"
)

// TypeDef is a Go type definition for a struct that is present in the
// definition of a Custom Resource Definition (CRD)
type TypeDef struct {
	Names names.Names
	Attrs map[string]*Attr
}

// GetTypeDefs returns a slice of TypeDef pointers and a map of package import
// information
func (h *Helper) GetTypeDefs() ([]*TypeDef, map[string]string, error) {
	crds, err := h.GetCRDs()
	if err != nil {
		return nil, nil, err
	}
	tdefs := []*TypeDef{}
	// Map, keyed by package import path, with the values being an alias to use
	// for the package
	timports := map[string]string{}

	payloads := h.getPayloads()

	crdNames := []string{}
	for _, crd := range crds {
		crdNames = append(crdNames, crd.Kind)
	}

	for shapeName, shape := range h.sdkAPI.Shapes {
		if inStrings(shapeName, crdNames) {
			// CRDs are already top-level structs
			continue
		}
		if inStrings(shapeName, payloads) {
			// Payloads are not type defs
			continue
		}
		if shape.Type != "structure" {
			continue
		}
		if shape.Exception {
			// Neither are exceptions
			continue
		}
		attrs := map[string]*Attr{}
		for propName, memberRef := range shape.MemberRefs {
			propNames := names.New(propName)
			propShape := memberRef.Shape
			goPkgType := memberRef.Shape.GoTypeWithPkgNameElem()
			if strings.Contains(goPkgType, ".") {
				if strings.HasPrefix(goPkgType, "[]") {
					// For slice types, we just want the element type...
					goPkgType = goPkgType[2:]
				} else if strings.HasPrefix(goPkgType, "map[") {
					goPkgType = strings.Split(goPkgType, "]")[1]
				}
				if strings.HasPrefix(goPkgType, "*") {
					// For slice types, the element type might be a pointer to
					// a struct...
					goPkgType = goPkgType[1:]
				}
				pkg := strings.Split(goPkgType, ".")[0]
				if pkg != h.sdkAPI.PackageName() {
					// Shape.GoPTypeWithPkgNameElem() always returns the type
					// as a full package dot-notation name. We only want to add
					// imports for "normal" package types like "time.Time", not
					// "ecr.ImageScanningConfiguration"
					timports[pkg] = ""
				}
			}
			attrs[propName] = NewAttr(propNames, propShape.GoType(), propShape)
		}
		if len(attrs) == 0 {
			// Just ignore these...
			continue
		}
		tdefs = append(tdefs, &TypeDef{
			Names: names.New(shapeName),
			Attrs: attrs,
		})
	}
	sort.Slice(tdefs, func(i, j int) bool {
		return tdefs[i].Names.Camel < tdefs[j].Names.Camel
	})
	return tdefs, timports, nil
}
