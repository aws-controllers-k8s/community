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
	"sort"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
	"github.com/gertd/go-pluralize"

	ackgenconfig "github.com/aws/aws-controllers-k8s/pkg/generate/config"
	"github.com/aws/aws-controllers-k8s/pkg/names"
	"github.com/aws/aws-controllers-k8s/pkg/util"
)

// Ops are the CRUD operations controlling a particular resource
type Ops struct {
	Create        *awssdkmodel.Operation
	ReadOne       *awssdkmodel.Operation
	ReadMany      *awssdkmodel.Operation
	Update        *awssdkmodel.Operation
	Delete        *awssdkmodel.Operation
	GetAttributes *awssdkmodel.Operation
	SetAttributes *awssdkmodel.Operation
}

// IterOps returns a slice of Operations for a resource
func (ops Ops) IterOps() []*awssdkmodel.Operation {
	res := []*awssdkmodel.Operation{}
	if ops.Create != nil {
		res = append(res, ops.Create)
	}
	if ops.ReadOne != nil {
		res = append(res, ops.ReadOne)
	}
	if ops.ReadMany != nil {
		res = append(res, ops.ReadMany)
	}
	if ops.Update != nil {
		res = append(res, ops.Update)
	}
	if ops.Delete != nil {
		res = append(res, ops.Delete)
	}
	return res
}

// PrinterColumn represents a single field in the CRD's Spec or Status objects
type PrinterColumn struct {
	CRD      *CRD
	Name     string
	Type     string
	JSONPath string
}

// CRD describes a single top-level resource in an AWS service API
type CRD struct {
	sdkAPI *SDKAPI
	cfg    *ackgenconfig.Config
	Names  names.Names
	Kind   string
	Plural string
	// Ops are the CRUD operations controlling this resource
	Ops Ops
	// AdditionalPrinterColumns is an array of PrinterColumn objects
	// representing the printer column settings for the CRD
	// AdditionalPrinterColumns field.
	AdditionalPrinterColumns []*PrinterColumn
	// SpecFields is a map, keyed by the **original SDK member name** of
	// Field objects representing those fields in the CRD's Spec struct
	// field.
	SpecFields map[string]*Field
	// StatusFields is a map, keyed by the **original SDK member name** of
	// Field objects representing those fields in the CRD's Status struct
	// field. Note that there are no fields in StatusFields that are also in
	// SpecFields.
	StatusFields map[string]*Field
	// TypeImports is a map, keyed by an import string, with the map value
	// being the import alias
	TypeImports map[string]string
	// ShortNames represent the CRD list of aliases. Short names allow shorter
	// strings to match a CR on the CLI.
	ShortNames []string
}

// Config returns a pointer to the generator config
func (r *CRD) Config() *ackgenconfig.Config {
	return r.cfg
}

// SDKAPIPackageName returns the aws-sdk-go package name used for this
// resource's API
func (r *CRD) SDKAPIPackageName() string {
	return r.sdkAPI.API.PackageName()
}

// TypeRenames returns a map of original type name to renamed name (some
// type definition names conflict with generated names)
func (r *CRD) TypeRenames() map[string]string {
	return r.sdkAPI.GetTypeRenames(r.cfg)
}

// HasShapeAsMember returns true if the supplied Shape name appears in *any*
// payload shape of *any* Operation for the resource. It recurses down through
// the resource's Operation Input and Output shapes and their member shapes
// looking for a shape with the supplied name
func (r *CRD) HasShapeAsMember(toFind string) bool {
	for _, op := range r.Ops.IterOps() {
		if op.InputRef.Shape != nil {
			inShape := op.InputRef.Shape
			for _, memberShapeRef := range inShape.MemberRefs {
				if shapeHasMember(memberShapeRef.Shape, toFind) {
					return true
				}
			}
		}
		if op.OutputRef.Shape != nil {
			outShape := op.OutputRef.Shape
			for _, memberShapeRef := range outShape.MemberRefs {
				if shapeHasMember(memberShapeRef.Shape, toFind) {
					return true
				}
			}
		}
	}
	return false
}

func shapeHasMember(shape *awssdkmodel.Shape, toFind string) bool {
	if shape.ShapeName == toFind {
		return true
	}
	switch shape.Type {
	case "structure":
		for _, memberShapeRef := range shape.MemberRefs {
			if shapeHasMember(memberShapeRef.Shape, toFind) {
				return true
			}
		}
	case "list":
		return shapeHasMember(shape.MemberRef.Shape, toFind)
	case "map":
		return shapeHasMember(shape.ValueRef.Shape, toFind)
	}
	return false
}

// InputFieldRename returns the renamed field for a supplied Operation ID and
// original field name and whether or not a renamed override field name was
// found
func (r *CRD) InputFieldRename(
	opID string,
	origFieldName string,
) (string, bool) {
	if r.cfg == nil {
		return origFieldName, false
	}
	return r.cfg.ResourceInputFieldRename(
		r.Names.Original, opID, origFieldName,
	)
}

// AddSpecField adds a new Field of a given name and shape into the Spec
// field of a CRD
func (r *CRD) AddSpecField(
	memberNames names.Names,
	shapeRef *awssdkmodel.ShapeRef,
) *Field {
	fieldConfigs := r.cfg.ResourceFields(r.Names.Original)
	f := newField(r, memberNames, shapeRef, fieldConfigs[memberNames.Original])
	r.SpecFields[memberNames.Original] = f
	return f
}

// AddStatusField adds a new Field of a given name and shape into the Status
// field of a CRD
func (r *CRD) AddStatusField(
	memberNames names.Names,
	shapeRef *awssdkmodel.ShapeRef,
) *Field {
	f := newField(r, memberNames, shapeRef, nil)
	r.StatusFields[memberNames.Original] = f
	return f
}

// AddTypeImport adds an entry in the CRD's TypeImports map for an import line
// and optional alias
func (r *CRD) AddTypeImport(
	packagePath string,
	alias string,
) {
	if r.TypeImports == nil {
		r.TypeImports = map[string]string{}
	}
	r.TypeImports[packagePath] = alias
}

// SpecFieldNames returns a sorted slice of field names for the Spec fields
func (r *CRD) SpecFieldNames() []string {
	res := make([]string, 0, len(r.SpecFields))
	for fieldName := range r.SpecFields {
		res = append(res, fieldName)
	}
	sort.Strings(res)
	return res
}

// AddPrintableColumn adds an entry to the list of additional printer columns
// using the given path and field types.
func (r *CRD) AddPrintableColumn(
	field *Field,
	jsonPath string,
) *PrinterColumn {
	fieldColumnType := field.GoTypeElem

	// Printable columns must be primitives supported by the OpenAPI list of data
	// types as defined by
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types
	// This maps Go type to OpenAPI type.
	acceptableColumnMaps := map[string]string{
		"string":  "string",
		"boolean": "boolean",
		"int":     "integer",
		"int8":    "integer",
		"int16":   "integer",
		"int32":   "integer",
		"int64":   "integer",
		"uint":    "integer",
		"uint8":   "integer",
		"uint16":  "integer",
		"uint32":  "integer",
		"uint64":  "integer",
		"uintptr": "integer",
		"float32": "number",
		"float64": "number",
	}
	printColumnType, exists := acceptableColumnMaps[fieldColumnType]

	if !exists {
		msg := fmt.Sprintf(
			"GENERATION FAILURE! Unable to generate a printer column for the field %s that has type %s.",
			field.Names.Camel, fieldColumnType,
		)
		panic(msg)
		return nil
	}

	column := &PrinterColumn{
		CRD:      r,
		Name:     field.Names.Camel,
		Type:     printColumnType,
		JSONPath: jsonPath,
	}
	r.AdditionalPrinterColumns = append(r.AdditionalPrinterColumns, column)
	return column
}

// AddSpecPrintableColumn adds an entry to the list of additional printer columns
// using the path of the given spec field.
func (r *CRD) AddSpecPrintableColumn(
	field *Field,
) *PrinterColumn {
	return r.AddPrintableColumn(
		field,
		//TODO(nithomso): Ideally we'd use `r.cfg.PrefixConfig.SpecField` but it uses uppercase
		fmt.Sprintf("%s.%s", ".spec", field.Names.CamelLower),
	)
}

// AddStatusPrintableColumn adds an entry to the list of additional printer columns
// using the path of the given status field.
func (r *CRD) AddStatusPrintableColumn(
	field *Field,
) *PrinterColumn {
	return r.AddPrintableColumn(
		field,
		//TODO(nithomso): Ideally we'd use `r.cfg.PrefixConfig.StatusField` but it uses uppercase
		fmt.Sprintf("%s.%s", ".status", field.Names.CamelLower),
	)
}

// UnpacksAttributesMap returns true if the underlying API has
// Get{Resource}Attributes/Set{Resource}Attributes API calls that map real,
// schema'd fields to a raw `map[string]*string` for this resource (see SNS and
// SQS APIs)
func (r *CRD) UnpacksAttributesMap() bool {
	return r.cfg.UnpacksAttributesMap(r.Names.Original)
}

// CompareIgnoredFields returns the list of fields compare logic should ignore
func (r *CRD) CompareIgnoredFields() []string {
	return r.cfg.GetCompareIgnoredFields(r.Names.Original)
}

// SetAttributesSingleAttribute returns true if the supplied resource name has
// a SetAttributes operation that only actually changes a single attribute at a
// time. See: SNS SetTopicAttributes API call, which is entirely different from
// the SNS SetPlatformApplicationAttributes API call, which sets multiple
// attributes at once. :shrug:
func (r *CRD) SetAttributesSingleAttribute() bool {
	return r.cfg.SetAttributesSingleAttribute(r.Names.Original)
}

// UnpackAttributes grabs instructions about fields that are represented in the
// AWS API as a `map[string]*string` but are actually real, schema'd fields and
// adds Field definitions for those fields.
func (r *CRD) UnpackAttributes() {
	if !r.cfg.UnpacksAttributesMap(r.Names.Original) {
		return
	}
	fieldConfigs := r.cfg.ResourceFields(r.Names.Original)
	for fieldName, fieldConfig := range fieldConfigs {
		if !fieldConfig.IsAttribute {
			continue
		}
		if r.IsPrimaryARNField(fieldName) {
			// ignore since this is handled by Status.ACKResourceMetadata.ARN
			continue
		}
		fieldNames := names.New(fieldName)
		f := newField(r, fieldNames, nil, fieldConfig)
		if !fieldConfig.IsReadOnly {
			r.SpecFields[fieldName] = f
		} else {
			r.StatusFields[fieldName] = f
		}
	}
}

// IsPrimaryARNField returns true if the supplied field name is likely the resource's
// ARN identifier field.
func (r *CRD) IsPrimaryARNField(fieldName string) bool {
	if r.cfg != nil && !r.cfg.IncludeACKMetadata {
		return false
	}
	return strings.EqualFold(fieldName, "arn") ||
		strings.EqualFold(fieldName, r.Names.Original+"arn")
}

// SetOutputCustomMethodName returns custom set output operation as *string for
// given operation on custom resource, if specified in generator config
func (r *CRD) SetOutputCustomMethodName(
	// The operation to look for the Output shape
	op *awssdkmodel.Operation,
) *string {
	if op == nil {
		return nil
	}
	if r.cfg == nil {
		return nil
	}
	resGenConfig, found := r.cfg.Operations[op.Name]
	if !found {
		return nil
	}

	if resGenConfig.SetOutputCustomMethodName == "" {
		return nil
	}
	return &resGenConfig.SetOutputCustomMethodName
}

// GetCustomImplementation returns custom implementation method name for the
// supplied operation as specified in generator config
func (r *CRD) GetCustomImplementation(
	// The type of operation
	op *awssdkmodel.Operation,
) string {
	if op == nil || r.cfg == nil {
		return ""
	}

	operationConfig, found := r.cfg.Operations[op.Name]
	if !found {
		return ""
	}

	return operationConfig.CustomImplementation
}

// UpdateConditionsCustomMethodName returns custom update conditions operation
// as *string for custom resource, if specified in generator config
func (r *CRD) UpdateConditionsCustomMethodName() string {
	if r.cfg == nil {
		return ""
	}
	resGenConfig, found := r.cfg.Resources[r.Names.Original]
	if !found {
		return ""
	}
	return resGenConfig.UpdateConditionsCustomMethodName
}

// NameField returns the name of the "Name" or string identifier field in the Spec
func (r *CRD) NameField() string {
	if r.cfg != nil {
		rConfig, found := r.cfg.Resources[r.Names.Original]
		if found {
			for fName, fConfig := range rConfig.Fields {
				if fConfig.IsName {
					return fName
				}
			}
		}
	}
	lookup := []string{
		"Name",
		r.Names.Original + "Name",
		r.Names.Original + "Id",
	}
	for memberName := range r.SpecFields {
		if util.InStrings(memberName, lookup) {
			return memberName
		}
	}
	return "???"
}

// CustomUpdateMethodName returns the name of the custom resourceManager method
// for updating the resource state, if any has been specified in the generator
// config
func (r *CRD) CustomUpdateMethodName() string {
	if r.cfg == nil {
		return ""
	}
	rConfig, found := r.cfg.Resources[r.Names.Original]
	if found {
		if rConfig.UpdateOperation != nil {
			return rConfig.UpdateOperation.CustomMethodName
		}
	}
	return ""
}

// ListOpMatchFieldNames returns a slice of strings representing the field
// names in the List operation's Output shape's element Shape that we should
// check a corresponding value in the target Spec exists.
func (r *CRD) ListOpMatchFieldNames() []string {
	return r.cfg.ListOpMatchFieldNames(r.Names.Original)
}

// NewCRD returns a pointer to a new `ackmodel.CRD` struct that describes a
// single top-level resource in an AWS service API
func NewCRD(
	sdkAPI *SDKAPI,
	cfg *ackgenconfig.Config,
	crdNames names.Names,
	ops Ops,
) *CRD {
	pluralize := pluralize.NewClient()
	kind := crdNames.Camel
	plural := pluralize.Plural(kind)
	return &CRD{
		sdkAPI:                   sdkAPI,
		cfg:                      cfg,
		Names:                    crdNames,
		Kind:                     kind,
		Plural:                   plural,
		Ops:                      ops,
		AdditionalPrinterColumns: make([]*PrinterColumn, 0),
		SpecFields:               map[string]*Field{},
		StatusFields:             map[string]*Field{},
		ShortNames:               cfg.ResourceShortNames(kind),
	}
}
