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

const (
	ConflictingNameSuffix = "_SDK"
)

// TypeDef is a Go type definition for structs that are member fields of the
// Spec or Status structs in Custom Resource Definitions (CRDs).
type TypeDef struct {
	Names names.Names
	Attrs map[string]*Attr
}

// HasConflictingTypeName returns true if the supplied type name will conflict
// with any generated type in the service's API package
func (h *Helper) HasConflictingTypeName(typeName string) bool {
	// First grab the set of CRD struct names and the names of their Spec and
	// Status structs
	cleanTypeName := names.New(typeName).Camel
	crdNames := h.GetCRDNames()
	crdResourceNames := []string{}
	crdSpecNames := []string{}
	crdStatusNames := []string{}

	for _, crdName := range crdNames {
		cleanResourceName := crdName.Camel
		crdResourceNames = append(crdResourceNames, cleanResourceName)
		crdSpecNames = append(crdSpecNames, cleanResourceName+"Spec")
		crdStatusNames = append(crdStatusNames, cleanResourceName+"Status")
	}
	return (inStrings(cleanTypeName, crdResourceNames) ||
		inStrings(cleanTypeName, crdSpecNames) ||
		inStrings(cleanTypeName, crdStatusNames))
}

// GetTypeDefs returns a slice of TypeDef pointers and a map of package import
// information
func (h *Helper) GetTypeDefs() ([]*TypeDef, map[string]string, error) {
	if h.typeDefs != nil {
		return h.typeDefs, h.typeImports, nil
	}

	tdefs := []*TypeDef{}
	// Map, keyed by package import path, with the values being an alias to use
	// for the package
	timports := map[string]string{}
	// Map, keyed by original Shape GoTypeElem(), with the values being a
	// renamed type name (due to conflicting names)
	trenames := map[string]string{}

	payloads := h.getPayloads()

	for shapeName, shape := range h.sdkAPI.Shapes {
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
		tdefNames := names.New(shapeName)
		if h.HasConflictingTypeName(shapeName) {
			tdefNames.Camel += ConflictingNameSuffix
			trenames[shapeName] = tdefNames.Camel
		}

		attrs := map[string]*Attr{}
		for memberName, memberRef := range shape.MemberRefs {
			memberNames := names.New(memberName)
			memberShape := memberRef.Shape
			goPkgType := memberRef.Shape.GoTypeWithPkgNameElem()
			if strings.Contains(goPkgType, ".") {
				if strings.HasPrefix(goPkgType, "[]") {
					// For slice types, we just want the element type...
					goPkgType = goPkgType[2:]
				}
				if strings.HasPrefix(goPkgType, "map[") {
					goPkgType = strings.Split(goPkgType, "]")[1]
				}
				if strings.HasPrefix(goPkgType, "*") {
					// For slice and map types, the element type might be a
					// pointer to a struct...
					goPkgType = goPkgType[1:]
				}
				pkg := strings.Split(goPkgType, ".")[0]
				if pkg != h.sdkAPI.PackageName() {
					// time.Time needs to be converted to apimachinery/metav1.Time otherwise there is no DeepCopy support
					if pkg == "time" {
						timports["k8s.io/apimachinery/pkg/apis/meta/v1"] = "metav1"
					} else if pkg == "aws" {
						// The "aws.JSONValue" type needs to be handled
						// specially.
						timports["github.com/aws/aws-sdk-go/aws"] = ""
					} else {
						// Shape.GoPTypeWithPkgNameElem() always returns the type
						// as a full package dot-notation name. We only want to add
						// imports for "normal" packages
						timports[pkg] = ""
					}
				}
			}
			// There are shapes that are called things like DBProxyStatus that are
			// fields in a DBProxy CRD... we need to ensure the type names don't
			// conflict. Also, the name of the Go type in the generated code is
			// Camel-cased and normalized, so we use that as the Go type
			gt := memberShape.GoType()
			if memberShape.Type == "structure" {
				typeNames := names.New(memberShape.ShapeName)
				if h.HasConflictingTypeName(memberShape.ShapeName) {
					typeNames.Camel += ConflictingNameSuffix
				}
				gt = "*" + typeNames.Camel
			} else if memberShape.Type == "list" {
				// If it's a list type, where the element is a structure, we need to
				// set the GoType to the cleaned-up Camel-cased name
				if memberShape.MemberRef.Shape.Type == "structure" {
					elemType := memberShape.MemberRef.Shape.GoTypeElem()
					typeNames := names.New(elemType)
					if h.HasConflictingTypeName(elemType) {
						typeNames.Camel += ConflictingNameSuffix
					}
					gt = "[]*" + typeNames.Camel
				}
			} else if memberShape.Type == "map" {
				// If it's a map type, where the value element is a structure,
				// we need to set the GoType to the cleaned-up Camel-cased name
				if memberShape.ValueRef.Shape.Type == "structure" {
					valType := memberShape.ValueRef.Shape.GoTypeElem()
					typeNames := names.New(valType)
					if h.HasConflictingTypeName(valType) {
						typeNames.Camel += ConflictingNameSuffix
					}
					gt = "[]map[string]*" + typeNames.Camel
				}
			} else if memberShape.Type == "timestamp" {
				// time.Time needs to be converted to apimachinery/metav1.Time
				// otherwise there is no DeepCopy support
				gt = "*metav1.Time"
			}
			attrs[memberName] = NewAttr(memberNames, gt, memberShape)
		}
		if len(attrs) == 0 {
			// Just ignore these...
			continue
		}
		tdefs = append(tdefs, &TypeDef{
			Names: tdefNames,
			Attrs: attrs,
		})
	}
	sort.Slice(tdefs, func(i, j int) bool {
		return tdefs[i].Names.Camel < tdefs[j].Names.Camel
	})
	h.typeDefs = tdefs
	h.typeImports = timports
	h.typeRenames = trenames
	return tdefs, timports, nil
}
