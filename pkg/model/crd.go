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
	"errors"
	"fmt"
	"sort"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
	"github.com/gertd/go-pluralize"

	"github.com/aws/aws-controllers-k8s/pkg/names"
)

type CRDOps struct {
	Create        *awssdkmodel.Operation
	ReadOne       *awssdkmodel.Operation
	ReadMany      *awssdkmodel.Operation
	Update        *awssdkmodel.Operation
	Delete        *awssdkmodel.Operation
	GetAttributes *awssdkmodel.Operation
	SetAttributes *awssdkmodel.Operation
}

// CRDField represents a single field in the CRD's Spec or Status objects
type CRDField struct {
	CRD                  *CRD
	Names                names.Names
	GoType               string
	GoTypeElem           string
	GoTypeWithPkgName    string
	ShapeRef             *awssdkmodel.ShapeRef
	FieldGeneratorConfig *FieldGeneratorConfig
}

// newCRDField returns a pointer to a new CRDField object
func newCRDField(
	crd *CRD,
	fieldNames names.Names,
	shapeRef *awssdkmodel.ShapeRef,
	generatorConfig *FieldGeneratorConfig,
) *CRDField {
	var gte, gt, gtwp string
	var shape *awssdkmodel.Shape
	if shapeRef != nil {
		shape = shapeRef.Shape
	}
	if shape != nil {
		gte, gt, gtwp = crd.cleanGoType(shape)
	} else {
		gte = "string"
		gt = "*string"
		gtwp = "*string"
	}
	return &CRDField{
		CRD:                  crd,
		Names:                fieldNames,
		ShapeRef:             shapeRef,
		GoType:               gt,
		GoTypeElem:           gte,
		GoTypeWithPkgName:    gtwp,
		FieldGeneratorConfig: generatorConfig,
	}
}

type CRD struct {
	helper *Helper
	Names  names.Names
	Kind   string
	Plural string
	Ops    CRDOps
	// SpecFields is a map, keyed by the **original SDK member name** of
	// CRDField objects representing those fields in the CRD's Spec struct
	// field.
	SpecFields map[string]*CRDField
	// StatusFields is a map, keyed by the **original SDK member name** of
	// CRDField objects representing those fields in the CRD's Status struct
	// field. Note that there are no fields in StatusFields that are also in
	// SpecFields.
	StatusFields map[string]*CRDField
	SDKMapper    *SDKMapper
	// TypeImports is a map, keyed by an import string, with the map value
	// being the import alias
	TypeImports map[string]string
}

func (r *CRD) cleanGoType(shape *awssdkmodel.Shape) (string, string, string) {
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
		if r.helper.HasConflictingTypeName(gte) {
			gte += "_SDK"
		}
		gt = "*" + gte
	} else if shape.Type == "list" {
		// If it's a list type, where the element is a structure, we need to
		// set the GoType to the cleaned-up Camel-cased name
		mgte, mgt, _ := r.cleanGoType(shape.MemberRef.Shape)
		cleanNames := names.New(mgte)
		gte = cleanNames.Camel
		if r.helper.HasConflictingTypeName(mgte) {
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

// AddSpecField adds a new CRDField of a given name and shape into the Spec
// field of a CRD
func (r *CRD) AddSpecField(
	memberNames names.Names,
	shapeRef *awssdkmodel.ShapeRef,
) {
	crdField := newCRDField(r, memberNames, shapeRef, nil)
	r.SpecFields[memberNames.Original] = crdField
}

// AddStatusField adds a new CRDField of a given name and shape into the Status
// field of a CRD
func (r *CRD) AddStatusField(
	memberNames names.Names,
	shapeRef *awssdkmodel.ShapeRef,
) {
	crdField := newCRDField(r, memberNames, shapeRef, nil)
	r.StatusFields[memberNames.Original] = crdField
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

// UnpacksAttributesMap returns true if the underlying API has
// Get{Resource}Attributes/Set{Resource}Attributes API calls that map real,
// schema'd fields to a raw `map[string]*string` for this resource (see SNS and
// SQS APIs)
func (r *CRD) UnpacksAttributesMap() bool {
	return r.helper.UnpacksAttributesMap(r.Names.Original)
}

// UnpackAttributes grabs instructions about fields that are represented in the
// AWS API as a `map[string]*string` but are actually real, schema'd fields and
// adds CRDField definitions for those fields.
func (r *CRD) UnpackAttributes() {
	if !r.helper.UnpacksAttributesMap(r.Names.Original) {
		return
	}
	attrMapConfig := r.helper.generatorConfig.Resources[r.Names.Original].UnpackAttributesMapConfig
	for fieldName, fieldConfig := range attrMapConfig.Fields {
		if r.IsPrimaryARNField(fieldName) {
			// ignore since this is handled by Status.ACKResourceMetadata.ARN
			continue
		}
		fieldNames := names.New(fieldName)
		crdField := newCRDField(r, fieldNames, nil, &fieldConfig)
		if !fieldConfig.IsReadOnly {
			r.SpecFields[fieldName] = crdField
		} else {
			r.StatusFields[fieldName] = crdField
		}
	}
}

// IsPrimaryARNField returns true if the supplied field name is likely the resource's
// ARN identifier field.
func (r *CRD) IsPrimaryARNField(fieldName string) bool {
	lowerName := strings.ToLower(fieldName)
	lowerResName := strings.ToLower(r.Names.Original)
	return lowerName == "arn" || lowerName == lowerResName+"arn"
}

// GoCodeSetInput returns the Go code that sets an input shape's member fields
// from a CRD's fields.
//
// Assume a CRD called Repository that looks like this pseudo-schema:
//
// .Status
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name (*string)
//
// And assume an SDK Shape CreateRepositoryInput that looks like this
// pseudo-schema:
//
// .Repository
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name
//
// This function is called from a template that generates the Go code that
// represents linkage between the Kubernetes objects (CRs) and the aws-sdk-go
// (SDK) objects. If we call this function with the following parameters:
//
//  opType:			OpTypeCreate
//  sourceVarName:	ko
//  targetVarName:	res
//  indentLevel:	1
//
// Then this function should output something like this:
//
//   field1 := []*string{}
//   for _, elem0 := range r.ko.Spec.Authors {
//       elem0 := &string{*elem0}
//       field0 = append(field0, elem0)
//   }
//   res.Authors = field1
//   field1 := &svcsdk.ImageData{}
//   field1.SetLocation(*r.ko.Spec.ImageData.Location)
//   field1.SetTag(*r.ko.Spec.ImageData.Tag)
//   res.ImageData = field1
//	 res.SetName(*r.ko.Spec.Name)
//
// Note that for scalar fields, we use the SetXXX methods that are on all
// aws-sdk-go SDK structs
func (r *CRD) GoCodeSetInput(
	// The type of operation to look for the Input shape
	opType OpType,
	// String representing the name of the variable that we will grab the Input
	// shape from. This will likely be "r.ko" since in the templates that call
	// this method, the "source variable" is the CRD struct which is used to
	// populate the target variable, which is the Input shape
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "res" since that is the name of the "target variable" that the
	// templates that call this method use for the Input shape.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	var op *awssdkmodel.Operation
	switch opType {
	case OpTypeCreate:
		op = r.Ops.Create
	case OpTypeGet:
		op = r.Ops.ReadOne
	case OpTypeList:
		op = r.Ops.ReadMany
	case OpTypeUpdate:
		op = r.Ops.Update
	case OpTypeDelete:
		op = r.Ops.Delete
	default:
		return ""
	}
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	// Some input shapes for APIs that use GetAttributes API calls don't have
	// an Attributes member (example: all the Delete shapes...)
	_, foundAttrs := inputShape.MemberRefs["Attributes"]
	if r.UnpacksAttributesMap() && foundAttrs {
		// For APIs that use a pattern of a parameter called "Attributes" that
		// is of type `map[string]*string` to represent real, schema'd fields,
		// we need to set the input shape's "Attributes" member field to the
		// re-constructed, packed set of fields.
		//
		// Therefore, we output here something like this (example from SNS
		// Topic's Attributes map):
		//
		// attrMap := map[string]*string{}
		// attrMap["DeliveryPolicy"] = r.ko.Spec.DeliveryPolicy
		// attrMap["DisplayName"} = r.ko.Spec.DisplayName
		// attrMap["KmsMasterKeyId"] = r.ko.Spec.KMSMasterKeyID
		// attrMap["Policy"] = r.ko.Spec.Policy
		// res.SetAttributes(attrMap)
		attrMapConfig := r.helper.generatorConfig.Resources[r.Names.Original].UnpackAttributesMapConfig
		out += fmt.Sprintf("%sattrMap := map[string]*string{}\n", indent)
		sortedAttrFieldNames := []string{}
		for fieldName := range attrMapConfig.Fields {
			sortedAttrFieldNames = append(sortedAttrFieldNames, fieldName)
		}
		sort.Strings(sortedAttrFieldNames)
		for _, fieldName := range sortedAttrFieldNames {
			fieldConfig := attrMapConfig.Fields[fieldName]
			fieldNames := names.New(fieldName)
			if !fieldConfig.IsReadOnly {
				out += fmt.Sprintf("%sattrMap[\"%s\"] = %s.%s\n", indent, fieldName, sourceVarName+".Spec", fieldNames.Camel)
			}
		}
		out += fmt.Sprintf("%s%s.SetAttributes(attrMap)\n", indent, targetVarName)
	}

	for memberIndex, memberName := range inputShape.MemberNames() {
		if r.UnpacksAttributesMap() && memberName == "Attributes" {
			continue
		}
		// Determine whether the input shape's field is in the Spec or the
		// Status struct and set the source variable appropriately.
		var crdField *CRDField
		var found bool
		sourceAdaptedVarName := sourceVarName
		crdField, found = r.SpecFields[memberName]
		if found {
			sourceAdaptedVarName += ".Spec"
		} else {
			crdField, found = r.StatusFields[memberName]
			if !found {
				// TODO(jaypipes): check generator config for exceptions?
				continue
			}
			sourceAdaptedVarName += ".Status"
		}
		memberShapeRef, _ := inputShape.MemberRefs[memberName]
		memberShape := memberShapeRef.Shape

		// we construct variables containing temporary storage for sub-elements
		// and sub-fields that are structs. Names of fields are "f" appended by
		// the 0-based index of the field within the set of the target struct's
		// set of fields. Nested structs simply append another "f" and the
		// field index to the variable name.
		//
		// This means you can tell what field a temporary fields variable
		// represents by the name.
		//
		// For example, the field variable name "f0f5f2", it contains the third
		// field of the sixth field of the first field of the input shape being
		// constructed.
		//
		// If we have two levels of nested struct fields, we will end
		// up with a targetVarName of "field0f0f0" and the generated code
		// might look something like this:
		//
		// res := &sdkapi.CreateBookInput{}
		// f0 := &sdkapi.BookData{}
		// f0f0 := &sdkapi.Author{}
		// f0f0f0 := &sdkapi.Address{}
		// f0f0f0.SetStreet(*ko.Spec.Author.Address.Street)
		// f0f0f0.SetCity(*ko.Spec.Author.Address.City)
		// f0f0f0.SetState(*ko.Spec.Author.Address.State)
		// f0f0.Address = f0f0f0
		// f0f0.SetName(*r.ko.Author.Name)
		// f0.Author = f0f0
		// res.Book = f0
		//
		// It's ugly but at least consistent and mostly readable...
		//
		// For populating list fields, we need an iterator and a temporary
		// element variable. We name these "{fieldName}iter" and
		// "{fieldName}elem" respectively. For nested levels, the names will be
		// progressively longer.
		//
		// For list fields, we want to end up with something like this:
		//
		// res := &sdkapi.CreateCustomAvailabilityZoneInput{}
		// f0 := []*sdkapi.VpnGroupMembership{}
		// for _, f0iter := ko.Spec.VPNGroupMemberships {
		//     f0elem := &sdkapi.VpnGroupMembership{}
		//     f0elem.SetVpnId(f0elem.VPNID)
		//     f0 := append(f0, f0elem)
		// }
		// res.VpnMemberships = f0

		switch memberShape.Type {
		case "list", "structure", "map":
			{
				memberVarName := fmt.Sprintf("f%d", memberIndex)
				out += r.goCodeVarEmptyConstructorSDKType(
					memberVarName,
					memberShape,
					indentLevel,
				)
				out += r.goCodeSetInputForContainer(
					memberName,
					memberVarName,
					sourceAdaptedVarName+"."+crdField.Names.Camel,
					memberShapeRef,
					indentLevel,
				)
				out += r.goCodeSetInputForScalar(
					memberName,
					targetVarName,
					inputShape.Type,
					memberVarName,
					memberShapeRef,
					indentLevel,
				)
			}
		default:
			out += r.goCodeSetInputForScalar(
				memberName,
				targetVarName,
				inputShape.Type,
				sourceAdaptedVarName+"."+crdField.Names.Camel,
				memberShapeRef,
				indentLevel,
			)
		}
	}
	return out
}

// GoCodeGetAttributesSetInput returns the Go code that sets the Input shape for a
// resource's GetAttributes operation.
//
// As an example, for the GetTopicAttributes SNS API call, the returned code
// looks like this:
//
// res.SetTopicArn(string(*r.ko.Status.ACKResourceMetadata.ARN))
//
// For the SQS API's GetQueueAttributes call, the returned code looks like this:
//
// res.SetQueueUrl(*r.ko.Status.QueueURL)
//
// You will note the difference due to the special handling of the ARN fields.
func (r *CRD) GoCodeGetAttributesSetInput(
	// String representing the name of the variable that we will grab the
	// Input shape from. This will likely be "r.ko.Spec" since in the templates
	// that call this method, the "source variable" is the CRD struct's Spec
	// field which is used to populate the target variable, which is the Input
	// shape
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "res" since that is the name of the "target variable" that the
	// templates that call this method use for the Input shape.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	op := r.Ops.GetAttributes
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	for _, memberName := range inputShape.MemberNames() {
		if r.IsPrimaryARNField(memberName) {
			out += fmt.Sprintf(
				"%s%s.Set%s(string(*%s.Status.ACKResourceMetadata.ARN))\n",
				indent, targetVarName, memberName, sourceVarName,
			)
			continue
		}

		cleanMemberNames := names.New(memberName)
		cleanMemberName := cleanMemberNames.Camel

		sourceVarPath := sourceVarName
		field, found := r.SpecFields[memberName]
		if found {
			sourceVarPath = sourceVarName + ".Spec." + cleanMemberName
		} else {
			field, found = r.StatusFields[memberName]
			if !found {
				// If it isn't in our spec/status fields, just ignore it
				continue
			}
			sourceVarPath = sourceVarPath + ".Status." + cleanMemberName
		}
		out += r.goCodeSetInputForScalar(
			memberName,
			targetVarName,
			inputShape.Type,
			sourceVarPath,
			field.ShapeRef,
			indentLevel,
		)
	}
	return out
}

func (r *CRD) goCodeSetInputForContainer(
	// The name of the SDK Input shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the struct field
	shapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	shape := shapeRef.Shape

	switch shape.Type {
	case "structure":
		{
			for memberIndex, memberName := range shape.MemberNames() {
				cleanMemberNames := names.New(memberName)
				cleanMemberName := cleanMemberNames.Camel
				memberVarName := fmt.Sprintf("%sf%d", targetVarName, memberIndex)
				memberShapeRef := shape.MemberRefs[memberName]
				memberShape := memberShapeRef.Shape
				switch memberShape.Type {
				case "list", "structure", "map":
					{
						out += r.goCodeVarEmptyConstructorSDKType(
							memberVarName,
							memberShape,
							indentLevel,
						)
						out += r.goCodeSetInputForContainer(
							memberName,
							memberVarName,
							sourceVarName+"."+cleanMemberName,
							memberShapeRef,
							indentLevel,
						)
						out += r.goCodeSetInputForScalar(
							memberName,
							targetVarName,
							shape.Type,
							memberVarName,
							memberShapeRef,
							indentLevel,
						)
					}
				default:
					out += r.goCodeSetInputForScalar(
						memberName,
						targetVarName,
						shape.Type,
						sourceVarName+"."+cleanMemberName,
						memberShapeRef,
						indentLevel,
					)
				}
			}
		}
	case "list":
		{
			iterVarName := fmt.Sprintf("%siter", targetVarName)
			elemVarName := fmt.Sprintf("%selem", targetVarName)
			// for _, f0iter := range r.ko.Spec.Tags {
			out += fmt.Sprintf("%sfor _, %s := range %s {\n", indent, iterVarName, sourceVarName)
			//		f0elem := string{}
			out += r.goCodeVarEmptyConstructorSDKType(
				elemVarName,
				shape.MemberRef.Shape,
				indentLevel+1,
			)
			//  f0elem = *f0iter
			//
			// or
			//
			//  f0elem.SetMyField(*f0iter)
			containerFieldName := ""
			if shape.MemberRef.Shape.Type == "structure" {
				containerFieldName = targetFieldName
			}
			out += r.goCodeSetInputForContainer(
				containerFieldName,
				elemVarName,
				iterVarName,
				&shape.MemberRef,
				indentLevel+1,
			)
			addressOfVar := ""
			switch shape.MemberRef.Shape.Type {
			case "structure", "list", "map":
				break
			default:
				addressOfVar = "&"
			}
			//  f0 = append(f0, elem0)
			out += fmt.Sprintf("%s\t%s = append(%s, %s%s)\n", indent, targetVarName, targetVarName, addressOfVar, elemVarName)
			out += fmt.Sprintf("%s}\n", indent)
		}
	case "map":
		{
			valIterVarName := fmt.Sprintf("%svaliter", targetVarName)
			keyVarName := fmt.Sprintf("%skey", targetVarName)
			valVarName := fmt.Sprintf("%sval", targetVarName)
			// for f0key, f0valiter := range r.ko.Spec.Tags {
			out += fmt.Sprintf("%sfor %s, %s := range %s {\n", indent, keyVarName, valIterVarName, sourceVarName)
			//		f0elem := string{}
			out += r.goCodeVarEmptyConstructorSDKType(
				valVarName,
				shape.ValueRef.Shape,
				indentLevel+1,
			)
			//  f0val = *f0valiter
			//
			// or
			//
			//  f0val.SetMyField(*f0valiter)
			containerFieldName := ""
			if shape.ValueRef.Shape.Type == "structure" {
				containerFieldName = targetFieldName
			}
			out += r.goCodeSetInputForContainer(
				containerFieldName,
				valVarName,
				valIterVarName,
				&shape.ValueRef,
				indentLevel+1,
			)
			addressOfVar := ""
			switch shape.ValueRef.Shape.Type {
			case "structure", "list", "map":
				break
			default:
				addressOfVar = "&"
			}
			// f0[f0key] = f0val
			out += fmt.Sprintf("%s\t%s[%s] = %s%s\n", indent, targetVarName, keyVarName, addressOfVar, valVarName)
			out += fmt.Sprintf("%s}\n", indent)
		}
	default:
		out += r.goCodeSetInputForScalar(
			targetFieldName,
			targetVarName,
			shape.Type,
			sourceVarName,
			shapeRef,
			indentLevel,
		)
	}
	return out
}

func (r *CRD) goCodeVarEmptyConstructorSDKType(
	varName string,
	// The shape we want to construct a new thing for
	shape *awssdkmodel.Shape,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	goType := shape.GoTypeWithPkgName()
	keepPointer := (shape.Type == "list" || shape.Type == "map")
	goType = r.replacePkgName(goType, "svcsdk", keepPointer)
	switch shape.Type {
	case "structure":
		// f0 := &svcsdk.BookData{}
		out += fmt.Sprintf("%s%s := &%s{}\n", indent, varName, goType)
	case "list", "map":
		// f0 := []*string{}
		out += fmt.Sprintf("%s%s := %s{}\n", indent, varName, goType)
	default:
		// var f0 string
		out += fmt.Sprintf("%svar %s %s\n", indent, varName, goType)
	}
	return out
}

func (r *CRD) goCodeVarEmptyConstructorK8sType(
	varName string,
	// The shape we want to construct a new thing for
	shape *awssdkmodel.Shape,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	goType := shape.GoTypeWithPkgName()
	keepPointer := (shape.Type == "list" || shape.Type == "map")
	goType = r.replacePkgName(goType, "svcapitypes", keepPointer)
	goTypeNoPkg := goType
	goPkg := ""
	hadPkg := false
	if strings.Contains(goType, ".") {
		parts := strings.Split(goType, ".")
		goTypeNoPkg = parts[1]
		goPkg = parts[0]
		hadPkg = true
	}
	renames := r.helper.GetTypeRenames()
	altTypeName, renamed := renames[goTypeNoPkg]
	if renamed {
		goTypeNoPkg = altTypeName
	}
	goType = goTypeNoPkg
	if hadPkg {
		goType = goPkg + "." + goType
	}

	switch shape.Type {
	case "structure":
		// f0 := &svcapitypes.BookData{}
		out += fmt.Sprintf("%s%s := &%s{}\n", indent, varName, goType)
	case "list", "map":
		// f0 := []*string{}
		out += fmt.Sprintf("%s%s := %s{}\n", indent, varName, goType)
	default:
		// var f0 string
		out += fmt.Sprintf("%svar %s %s\n", indent, varName, goType)
	}
	return out
}

// goCodeSetInputForScalar returns the Go code that sets the value of a target
// variable or field to a scalar value. For target variables that are structs,
// we output the aws-sdk-go's common SetXXX() method. For everything else, we
// output normal assignment operations.
func (r *CRD) goCodeSetInputForScalar(
	// The name of the Input SDK Shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The type of shape of the target variable
	targetVarType string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	shapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	setTo := sourceVarName
	shape := shapeRef.Shape
	if shape.Type == "timestamp" {
		setTo += ".Time"
	} else if shapeRef.UseIndirection() {
		setTo = "*" + setTo
	}
	if targetVarType == "structure" {
		out += fmt.Sprintf("%s%s.Set%s(%s)\n", indent, targetVarName, targetFieldName, setTo)
	} else {
		targetVarPath := targetVarName
		if targetFieldName != "" {
			targetVarPath += "." + targetFieldName
		}
		out += fmt.Sprintf("%s%s = %s\n", indent, targetVarPath, setTo)
	}
	return out
}

// GoCodeSetOutput returns the Go code that sets a CRD's Status field value to
// the value of an output shape's member fields.
//
// Assume a CRD called Repository that looks like this pseudo-schema:
//
// .Status
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name (*string)
//
// And assume an SDK Shape CreateRepositoryOutput that looks like this
// pseudo-schema:
//
// .Repository
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name
//
// This function is called from a template that generates the Go code that
// represents linkage between the Kubernetes objects (CRs) and the aws-sdk-go
// (SDK) objects. If we call this function with the following parameters:
//
//  opType:			OpTypeCreate
//  sourceVarName:	resp
//  targetVarName:	ko.Status
//  indentLevel:	1
//
// Then this function should output something like this:
//
//   field0 := []*string{}
//   for _, iter0 := range resp.Authors {
//       elem0 := &string{*iter0}
//       field0 = append(field0, elem0)
//   }
//   ko.Status.Authors = field0
//   field1 := &svcapitypes.ImageData{}
//   field1.Location = resp.ImageData.Location
//   field1.Tag = resp.ImageData.Tag
//   ko.Status.ImageData = field1
//   ko.Status.Name = resp.Name
func (r *CRD) GoCodeSetOutput(
	// The type of operation to look for the Output shape
	opType OpType,
	// String representing the name of the variable that we will grab the
	// Output shape from. This will likely be "resp" since in the templates
	// that call this method, the "source variable" is the response struct
	// returned by the aws-sdk-go's SDK API call corresponding to the Operation
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "ko.Status" since that is the name of the "target variable" that the
	// templates that call this method use.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	var op *awssdkmodel.Operation
	switch opType {
	case OpTypeCreate:
		op = r.Ops.Create
	case OpTypeGet:
		op = r.Ops.ReadOne
	case OpTypeList:
		op = r.Ops.ReadMany
	case OpTypeUpdate:
		op = r.Ops.Update
	case OpTypeDelete:
		op = r.Ops.Delete
	default:
		return ""
	}
	if op == nil {
		return ""
	}
	outputShape := op.OutputRef.Shape
	if outputShape == nil {
		return ""
	}

	// We might be in a "wrapper" shape. Unwrap it to find the real object
	// representation for the CRD's createOp. If there is a single member
	// shape and that member shape is a structure, unwrap it.
	if outputShape.UsedAsOutput && len(outputShape.MemberRefs) == 1 {
		for memberName, memberRef := range outputShape.MemberRefs {
			if memberRef.Shape.Type == "structure" {
				sourceVarName += "." + memberName
				outputShape = memberRef.Shape
			}
		}
	}
	out := "\n"

	// Recursively descend down through the set of fields on the Output shape,
	// creating temporary variables, populating those temporary variables'
	// fields with further-nested fields as needed
	for memberIndex, memberName := range outputShape.MemberNames() {
		statusField, found := r.StatusFields[memberName]
		if !found {
			// Note that not all fields in the output shape will be in the
			// Status fields collection of the CRD. If a same-named field is in
			// the Spec, then that's where it stays. This function is only here
			// to set the Status field values after getting a response via the
			// aws-sdk-go for an API call...
			continue
		}

		// TODO(jaypipes): Handle the special case of ARN for primary
		// resource identifier

		memberShapeRef := outputShape.MemberRefs[memberName]
		if memberShapeRef.Shape == nil {
			// Technically this should not happen, so let's bail here if it
			// does...
			msg := fmt.Sprintf(
				"expected .Shape to not be nil for ShapeRef of memberName %s",
				memberName,
			)
			panic(msg)
		}

		// fieldVarName is the name of the variable that is used for temporary
		// storage of complex member field values
		//
		// For struct fields, we want to output code sort of like this:
		//
		//   field0 := &svapitypes.ImageData{}
		//   field0.Location = resp.ImageData.Location
		//   field0.Tag = resp.ImageData.Tag
		//   r.ko.Status.ImageData = field0
		//   r.ko.Status.Name = resp.Name
		//
		// For list fields, we want to end up with something like this:
		//
		// field0 := []*svcapitypes.VpnGroupMembership{}
		// for _, iter0 := resp.CustomAvailabilityZone.VpnGroupMemberships {
		//     elem0 := &svcapitypes.VPNGroupMembership{}
		//     elem0.VPNID = iter0.VPNID
		//     field0 := append(field0, elem0)
		// }
		// ko.Status.VpnMemberships = field0

		memberShape := memberShapeRef.Shape
		switch memberShape.Type {
		case "list", "structure", "map":
			{
				memberVarName := fmt.Sprintf("f%d", memberIndex)
				out += r.goCodeVarEmptyConstructorK8sType(
					memberVarName,
					memberShape,
					indentLevel,
				)
				out += r.goCodeSetOutputForContainer(
					statusField.Names.Camel,
					memberVarName,
					sourceVarName+"."+memberName,
					memberShapeRef,
					indentLevel,
				)
				out += r.goCodeSetOutputForScalar(
					statusField.Names.Camel,
					targetVarName,
					memberVarName,
					memberShapeRef,
					indentLevel,
				)
			}
		default:
			out += r.goCodeSetOutputForScalar(
				statusField.Names.Camel,
				targetVarName,
				sourceVarName+"."+memberName,
				memberShapeRef,
				indentLevel,
			)
		}
	}
	return out
}

// GoCodeGetAttributesSetOutput returns the Go code that sets the Status fields
// from the Output shape returned from a resource's GetAttributes operation.
//
// As an example, for the GetTopicAttributes SNS API call, the returned code
// looks like this:
//
// ko.Status.EffectiveDeliveryPolicy = resp.Attributes["EffectiveDeliveryPolicy"]
// ko.Status.ACKResourceMetadata.OwnerAccountID = ackv1alpha1.AWSAccountID(resp.Attributes["Owner"])
// ko.Status.ACKResourceMetadata.ARN = ackv1alpha1.AWSResourceName(resp.Attributes["TopicArn"])
func (r *CRD) GoCodeGetAttributesSetOutput(
	// String representing the name of the variable that we will grab the
	// Output shape from. This will likely be "resp" since in the templates
	// that call this method, the "source variable" is the response struct
	// returned by the aws-sdk-go's SDK API call corresponding to the Operation
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "ko.Status" since that is the name of the "target variable" that the
	// templates that call this method use.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	if !r.UnpacksAttributesMap() {
		// This is a bug in the code generation if this occurs...
		msg := fmt.Sprintf("called GoCodeGetAttributesSetOutput for a resource '%s' that doesn't unpack attributes map", r.Ops.GetAttributes.Name)
		panic(msg)
	}
	op := r.Ops.GetAttributes
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	attrMapConfig := r.helper.generatorConfig.Resources[r.Names.Original].UnpackAttributesMapConfig
	sortedAttrFieldNames := []string{}
	for fieldName := range attrMapConfig.Fields {
		sortedAttrFieldNames = append(sortedAttrFieldNames, fieldName)
	}
	sort.Strings(sortedAttrFieldNames)
	for _, fieldName := range sortedAttrFieldNames {
		if r.IsPrimaryARNField(fieldName) {
			out += fmt.Sprintf(
				"%stmpARN := ackv1alpha1.AWSResourceName(*%s.Attributes[\"%s\"])\n",
				indent,
				sourceVarName,
				fieldName,
			)
			out += fmt.Sprintf(
				"%s%s.ACKResourceMetadata.ARN = &tmpARN\n",
				indent,
				targetVarName,
			)
			continue
		}

		fieldConfig := attrMapConfig.Fields[fieldName]
		if fieldConfig.ContainsOwnerAccountID {
			out += fmt.Sprintf(
				"%stmpOwnerID := ackv1alpha1.AWSAccountID(*%s.Attributes[\"%s\"])\n",
				indent,
				sourceVarName,
				fieldName,
			)
			out += fmt.Sprintf(
				"%s%s.ACKResourceMetadata.OwnerAccountID = &tmpOwnerID\n",
				indent,
				targetVarName,
			)
			continue
		}

		fieldNames := names.New(fieldName)
		if fieldConfig.IsReadOnly {
			out += fmt.Sprintf(
				"%s%s.%s = %s.Attributes[\"%s\"]\n",
				indent,
				targetVarName,
				fieldNames.Camel,
				sourceVarName,
				fieldName,
			)
		}
	}
	return out
}

func (r *CRD) goCodeSetOutputForContainer(
	// The name of the SDK Input shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the struct field
	shapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	shape := shapeRef.Shape

	switch shape.Type {
	case "structure":
		{
			for memberIndex, memberName := range shape.MemberNames() {
				memberVarName := fmt.Sprintf("%sf%d", targetVarName, memberIndex)
				memberShapeRef := shape.MemberRefs[memberName]
				memberShape := memberShapeRef.Shape
				cleanNames := names.New(memberName)
				switch memberShape.Type {
				case "list", "structure", "map":
					{
						out += r.goCodeVarEmptyConstructorK8sType(
							memberVarName,
							memberShape,
							indentLevel,
						)
						out += r.goCodeSetOutputForContainer(
							cleanNames.Camel,
							memberVarName,
							sourceVarName+"."+memberName,
							memberShapeRef,
							indentLevel,
						)
						out += r.goCodeSetOutputForScalar(
							cleanNames.Camel,
							targetVarName,
							memberVarName,
							memberShapeRef,
							indentLevel,
						)
					}
				default:
					out += r.goCodeSetOutputForScalar(
						cleanNames.Camel,
						targetVarName,
						sourceVarName+"."+memberName,
						memberShapeRef,
						indentLevel,
					)
				}
			}
		}
	case "list":
		{
			iterVarName := fmt.Sprintf("%siter", targetVarName)
			elemVarName := fmt.Sprintf("%selem", targetVarName)
			// for _, f0iter0 := range resp.TagSpecifications {
			out += fmt.Sprintf("%sfor _, %s := range %s {\n", indent, iterVarName, sourceVarName)
			//		f0elem0 := &string{}
			out += r.goCodeVarEmptyConstructorK8sType(
				elemVarName,
				shape.MemberRef.Shape,
				indentLevel+1,
			)
			//  f0elem0 = *f0iter0
			//
			// or
			//
			//  f0elem0.SetMyField(*f0iter0)
			containerFieldName := ""
			if shape.MemberRef.Shape.Type == "structure" {
				containerFieldName = targetFieldName
			}
			out += r.goCodeSetOutputForContainer(
				containerFieldName,
				elemVarName,
				iterVarName,
				&shape.MemberRef,
				indentLevel+1,
			)
			addressOfVar := ""
			switch shape.MemberRef.Shape.Type {
			case "structure", "list", "map":
				break
			default:
				addressOfVar = "&"
			}
			//  f0 = append(f0, elem0)
			out += fmt.Sprintf("%s\t%s = append(%s, %s%s)\n", indent, targetVarName, targetVarName, addressOfVar, elemVarName)
			out += fmt.Sprintf("%s}\n", indent)
		}
	case "map":
		{
			valIterVarName := fmt.Sprintf("%svaliter", targetVarName)
			keyVarName := fmt.Sprintf("%skey", targetVarName)
			valVarName := fmt.Sprintf("%sval", targetVarName)
			// for f0key, f0valiter := range resp.Tags {
			out += fmt.Sprintf("%sfor %s, %s := range %s {\n", indent, keyVarName, valIterVarName, sourceVarName)
			//		f0elem := string{}
			out += r.goCodeVarEmptyConstructorK8sType(
				valVarName,
				shape.ValueRef.Shape,
				indentLevel+1,
			)
			//  f0val = *f0valiter
			containerFieldName := ""
			if shape.ValueRef.Shape.Type == "structure" {
				containerFieldName = targetFieldName
			}
			out += r.goCodeSetOutputForContainer(
				containerFieldName,
				valVarName,
				valIterVarName,
				&shape.ValueRef,
				indentLevel+1,
			)
			addressOfVar := ""
			switch shape.ValueRef.Shape.Type {
			case "structure", "list", "map":
				break
			default:
				addressOfVar = "&"
			}
			// f0[f0key] = f0val
			out += fmt.Sprintf("%s\t%s[%s] = %s%s\n", indent, targetVarName, keyVarName, addressOfVar, valVarName)
			out += fmt.Sprintf("%s}\n", indent)
		}
	default:
		out += r.goCodeSetOutputForScalar(
			targetFieldName,
			targetVarName,
			sourceVarName,
			shapeRef,
			indentLevel,
		)
	}
	return out
}

func (r *CRD) goCodeSetOutputForScalar(
	// The name of the Input SDK Shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	shapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	setTo := sourceVarName
	shape := shapeRef.Shape
	if shape.Type == "timestamp" {
		setTo = "&metav1.Time{*" + sourceVarName + "}"
	}
	targetVarPath := targetVarName
	if targetFieldName != "" {
		targetVarPath += "." + targetFieldName
	} else {
		setTo = "*" + setTo
	}
	out += fmt.Sprintf("%s%s = %s\n", indent, targetVarPath, setTo)
	return out
}

// replacePkgName accepts a type string, as returned by
// Shape.GoTypeWithPkgName() and replaces the package name of the aws-sdk-go
// SDK API (e.g. "ecr" for the ECR API) with the string "svcsdkapi" which is
// the only alias we always use in our templated output.
func (r *CRD) replacePkgName(
	subject string,
	replacePkgAlias string,
	keepPointer bool,
) string {
	memberType := subject
	isSliceType := strings.HasPrefix(memberType, "[]")
	if isSliceType {
		memberType = memberType[2:]
	}
	isMapType := strings.HasPrefix(memberType, "map[string]")
	if isMapType {
		memberType = memberType[11:]
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
		apiPkgName := r.helper.sdkAPI.PackageName()
		if pkgName == apiPkgName {
			memberType = replacePkgAlias + "." + typeName
		} else {
			// Leave package prefixes like "time." alone...
			memberType = pkgName + "." + typeName
		}
	}
	if isPointerType && keepPointer {
		memberType = "*" + memberType
	}
	if isMapType {
		memberType = "map[string]" + memberType
	}
	if isSliceType {
		memberType = "[]" + memberType
	}
	return memberType
}

func newCRD(
	helper *Helper,
	crdNames names.Names,
	crdOps CRDOps,
) *CRD {
	pluralize := pluralize.NewClient()
	kind := crdNames.Camel
	plural := pluralize.Plural(kind)
	return &CRD{
		helper:       helper,
		Names:        crdNames,
		Kind:         kind,
		Plural:       plural,
		Ops:          crdOps,
		SpecFields:   map[string]*CRDField{},
		StatusFields: map[string]*CRDField{},
	}
}

var (
	ErrNilShapePointer = errors.New("found nil Shape pointer")
)

func (h *Helper) GetCRDs() ([]*CRD, error) {
	if h.crds != nil {
		return h.crds, nil
	}
	crds := []*CRD{}

	opMap := h.GetOperationMap()

	createOps := (*opMap)[OpTypeCreate]
	readOneOps := (*opMap)[OpTypeGet]
	readManyOps := (*opMap)[OpTypeList]
	updateOps := (*opMap)[OpTypeUpdate]
	deleteOps := (*opMap)[OpTypeDelete]
	getAttributesOps := (*opMap)[OpTypeGetAttributes]
	setAttributesOps := (*opMap)[OpTypeSetAttributes]

	for crdName, createOp := range createOps {
		if h.IsIgnoredResource(crdName) {
			continue
		}
		crdNames := names.New(crdName)
		crdOps := CRDOps{
			Create:        createOps[crdName],
			ReadOne:       readOneOps[crdName],
			ReadMany:      readManyOps[crdName],
			Update:        updateOps[crdName],
			Delete:        deleteOps[crdName],
			GetAttributes: getAttributesOps[crdName],
			SetAttributes: setAttributesOps[crdName],
		}
		h.RemoveIgnoredOperations(&crdOps)
		crd := newCRD(h, crdNames, crdOps)
		sdkMapper := NewSDKMapper(crd)
		crd.SDKMapper = sdkMapper

		// OK, begin to gather the CRDFields that will go into the Spec struct.
		// These fields are those members of the Create operation's Input
		// Shape.
		inputShape := createOp.InputRef.Shape
		if inputShape == nil {
			return nil, ErrNilShapePointer
		}
		for memberName, memberShapeRef := range inputShape.MemberRefs {
			memberNames := names.New(memberName)
			if memberShapeRef.Shape == nil {
				return nil, ErrNilShapePointer
			}
			if memberName == "Attributes" && h.UnpacksAttributesMap(crdName) {
				crd.UnpackAttributes()
				continue
			}
			crd.AddSpecField(memberNames, memberShapeRef)
		}

		// Now process the fields that will go into the Status struct. We want
		// fields that are in the Create operation's Output Shape but that are
		// not in the Input Shape.
		outputShape := createOp.OutputRef.Shape
		if outputShape.UsedAsOutput && len(outputShape.MemberRefs) == 1 {
			// We might be in a "wrapper" shape. Unwrap it to find the real object
			// representation for the CRD's createOp. If there is a single member
			// shape and that member shape is a structure, unwrap it.
			for _, memberRef := range outputShape.MemberRefs {
				if memberRef.Shape.Type == "structure" {
					outputShape = memberRef.Shape
				}
			}
		}
		for memberName, memberShapeRef := range outputShape.MemberRefs {
			memberNames := names.New(memberName)
			if memberShapeRef.Shape == nil {
				return nil, ErrNilShapePointer
			}
			if _, found := crd.SpecFields[memberName]; found {
				// We don't put fields that are already in the Spec struct into
				// the Status struct
				continue
			}
			if memberName == "Attributes" && h.UnpacksAttributesMap(crdName) {
				continue
			}
			if strings.EqualFold(memberName, "arn") ||
				strings.EqualFold(memberName, crdName+"arn") {
				// Normalize primary resource ARN field in the returned output
				// shape. We want to map this Shape into the
				// Status.ACKResourceMetadata.ARN field
				sdkMapper.SetPrimaryResourceARNField(createOp, memberName)
				continue
			}
			crd.AddStatusField(memberNames, memberShapeRef)
		}

		crds = append(crds, crd)
	}
	sort.Slice(crds, func(i, j int) bool {
		return crds[i].Names.Camel < crds[j].Names.Camel
	})
	h.crds = crds
	return crds, nil
}

// IsIgnoredResource returns true if Operation Name is configured to be ignored
// in generator config for the AWS service
func (h *Helper) IsIgnoredResource(resourceName string) bool {
	if resourceName == "" {
		return true
	}
	if h.generatorConfig != nil {
		for _, ignoredResourceName := range h.generatorConfig.Ignore.ResourceNames {
			if ignoredResourceName == resourceName {
				return true
			}
		}
	}
	return false
}

// RemoveIgnoredOperations updates CRDOps argument by setting those operations to nil
// that are configured to be ignored in generator config for the AWS service
func (h *Helper) RemoveIgnoredOperations(crdOps *CRDOps) {
	if h.IsIgnoredOperation(crdOps.Create) {
		crdOps.Create = nil
	}
	if h.IsIgnoredOperation(crdOps.ReadOne) {
		crdOps.ReadOne = nil
	}
	if h.IsIgnoredOperation(crdOps.ReadMany) {
		crdOps.ReadMany = nil
	}
	if h.IsIgnoredOperation(crdOps.Update) {
		crdOps.Update = nil
	}
	if h.IsIgnoredOperation(crdOps.Delete) {
		crdOps.Delete = nil
	}
	if h.IsIgnoredOperation(crdOps.GetAttributes) {
		crdOps.GetAttributes = nil
	}
	if h.IsIgnoredOperation(crdOps.SetAttributes) {
		crdOps.SetAttributes = nil
	}
}

// IsIgnoredOperation returns true if Operation Name is configured to be ignored
// in generator config for the AWS service
func (h *Helper) IsIgnoredOperation(operation *awssdkmodel.Operation) bool {
	if operation == nil {
		return true
	}
	if h.generatorConfig != nil {
		for _, ignoredOperation := range h.generatorConfig.Ignore.Operations {
			if ignoredOperation == operation.Name {
				return true
			}
		}
	}
	return false
}

// UnpacksAttributeMap returns true if the underlying API has
// Get{Resource}Attributes/Set{Resource}Attributes API calls that map real,
// schema'd fields to a raw `map[string]*string` for this resource (see SNS and
// SQS APIs)
func (h *Helper) UnpacksAttributesMap(resourceName string) bool {
	if h.generatorConfig != nil {
		resGenConfig, found := h.generatorConfig.Resources[resourceName]
		if found {
			return resGenConfig.UnpackAttributesMapConfig != nil
		}
	}
	return false
}

// GetOperationMap returns a map, keyed by the operation type and operation
// ID/name, of aws-sdk-go private/model/api.Operation struct pointers
func (h *Helper) GetOperationMap() *OperationMap {
	if h.opMap != nil {
		return h.opMap
	}
	// create an index of Operations by operation types and resource name
	opMap := OperationMap{}
	for opID, op := range h.sdkAPI.Operations {
		opType, resName := GetOpTypeAndResourceNameFromOpID(opID)
		if _, found := opMap[opType]; !found {
			opMap[opType] = map[string]*awssdkmodel.Operation{}
		}
		opMap[opType][resName] = op
	}
	h.opMap = &opMap
	return &opMap
}
