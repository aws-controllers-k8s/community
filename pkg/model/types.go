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
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

	ackgenconfig "github.com/aws/aws-controllers-k8s/pkg/generate/config"
	"github.com/aws/aws-controllers-k8s/pkg/names"
)

// cleanGoType returns a tuple of three strings representing the normalized Go
// types in "element", "normal" and "with package name" format for a particular
// Shape.
func cleanGoType(
	api *SDKAPI,
	cfg *ackgenconfig.Config,
	shape *awssdkmodel.Shape,
) (string, string, string) {
	// There are shapes that are called things like DBProxyStatus that are
	// fields in a DBProxy CRD... we need to ensure the type names don't
	// conflict. Also, the name of the Go type in the generated code is
	// Camel-cased and normalized, so we use that as the Go type
	gt := shape.GoType()
	gte := shape.GoTypeElem()
	gtwp := shape.GoTypeWithPkgName()
	// Normalize the type names for structs and list elements
	if shape.Type == "structure" {
		cleanNames := names.New(gte)
		gte = cleanNames.Camel
		if api.HasConflictingTypeName(gte, cfg) {
			gte += "_SDK"
		}
		gt = "*" + gte
	} else if shape.Type == "list" {
		// If it's a list type, where the element is a structure, we need to
		// set the GoType to the cleaned-up Camel-cased name
		mgte, mgt, _ := cleanGoType(api, cfg, shape.MemberRef.Shape)
		cleanNames := names.New(mgte)
		gte = cleanNames.Camel
		if api.HasConflictingTypeName(mgte, cfg) {
			gte += "_SDK"
		}

		gt = "[]" + mgt
	} else if shape.Type == "timestamp" {
		// time.Time needs to be converted to apimachinery/metav1.Time
		// otherwise there is no DeepCopy support
		gtwp = "*metav1.Time"
		gte = "metav1.Time"
		gt = "*metav1.Time"
	}

	// Replace the type part of the full type-with-package-name with the
	// cleaned up type name
	typeParts := strings.Split(gtwp, ".")
	if len(typeParts) == 2 {
		gtwp = typeParts[0] + "." + gte
	}
	return gte, gt, gtwp
}

// replacePkgName accepts a type string, as returned by
// Shape.GoTypeWithPkgName() and replaces the package name of the aws-sdk-go
// SDK API (e.g. "ecr" for the ECR API) with the string "svcsdkapi" which is
// the only alias we always use in our templated output.
func replacePkgName(
	subject string,
	apiPkgName string,
	replacePkgAlias string,
	keepPointer bool,
) string {
	memberType := subject
	sliceDepth := 0 // Depth of the slice type
	isSliceType := strings.HasPrefix(memberType, "[]")
	if isSliceType {
		sliceDepth = strings.LastIndex(subject, "[]")/2 + 1
		memberType = memberType[sliceDepth*2:]
	}
	mapDepth := 0 // Depth of the map type
	// Assuming the map keys are always of type string.
	isMapType := strings.HasPrefix(memberType, "map[string]")
	if isMapType {
		mapDepth = strings.LastIndex(subject, "map[string]")/11 + 1
		memberType = memberType[mapDepth*11:]
	}
	isPointerType := strings.HasPrefix(memberType, "*")
	if isPointerType {
		memberType = memberType[1:]
	}
	// We need to convert any package name that the aws-sdk-private
	// model uses "such as 'ecr.' to just 'svcapitypes' since we always
	// alias the Kubernetes API types for the service API with that
	if strings.Contains(memberType, ".") {
		pkgName := strings.Split(memberType, ".")[0]
		typeName := strings.Split(memberType, ".")[1]
		if pkgName == apiPkgName {
			memberType = replacePkgAlias + "." + typeName
		} else {
			memberType = pkgName + "." + typeName
		}
	}
	if isPointerType && keepPointer {
		memberType = "*" + memberType
	}
	if isMapType {
		memberType = strings.Repeat("map[string]", mapDepth) + memberType
	}
	if isSliceType {
		memberType = strings.Repeat("[]", sliceDepth) + memberType
	}
	return memberType
}
