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

package generate

import (
	"sort"
	"strings"
	ttpl "text/template"

	ackgenconfig "github.com/aws/aws-controllers-k8s/pkg/generate/config"
	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/names"
	"github.com/aws/aws-controllers-k8s/pkg/util"
)

// Generator creates the ACK service controller Kubernetes API types (CRDs) and
// the service controller implementation/SDK linkage
type Generator struct {
	SDKAPI           *ackmodel.SDKAPI
	serviceAlias     string
	apiVersion       string
	templateBasePath string
	templates        map[string]*ttpl.Template
	crds             []*ackmodel.CRD
	typeDefs         []*ackmodel.TypeDef
	typeImports      map[string]string
	typeRenames      map[string]string
	// Instructions to the code generator how to handle the API and its
	// resources
	cfg *ackgenconfig.Config
}

// GetCRDs returns a slice of `ackmodel.CRD` structs that describe the
// top-level resources discovered by the code generator for an AWS service API
func (g *Generator) GetCRDs() ([]*ackmodel.CRD, error) {
	if g.crds != nil {
		return g.crds, nil
	}
	crds := []*ackmodel.CRD{}

	opMap := g.SDKAPI.GetOperationMap()

	createOps := (*opMap)[ackmodel.OpTypeCreate]
	readOneOps := (*opMap)[ackmodel.OpTypeGet]
	readManyOps := (*opMap)[ackmodel.OpTypeList]
	updateOps := (*opMap)[ackmodel.OpTypeUpdate]
	deleteOps := (*opMap)[ackmodel.OpTypeDelete]
	getAttributesOps := (*opMap)[ackmodel.OpTypeGetAttributes]
	setAttributesOps := (*opMap)[ackmodel.OpTypeSetAttributes]

	for crdName, createOp := range createOps {
		if g.cfg.IsIgnoredResource(crdName) {
			continue
		}
		crdNames := names.New(crdName)
		crdOps := ackmodel.CRDOps{
			Create:        createOps[crdName],
			ReadOne:       readOneOps[crdName],
			ReadMany:      readManyOps[crdName],
			Update:        updateOps[crdName],
			Delete:        deleteOps[crdName],
			GetAttributes: getAttributesOps[crdName],
			SetAttributes: setAttributesOps[crdName],
		}
		g.RemoveIgnoredOperations(&crdOps)
		crd := ackmodel.NewCRD(g.SDKAPI, g.cfg, crdNames, crdOps)

		// OK, begin to gather the CRDFields that will go into the Spec struct.
		// These fields are those members of the Create operation's Input
		// Shape.
		inputShape := createOp.InputRef.Shape
		if inputShape == nil {
			return nil, ackmodel.ErrNilShapePointer
		}
		for memberName, memberShapeRef := range inputShape.MemberRefs {
			if memberShapeRef.Shape == nil {
				return nil, ackmodel.ErrNilShapePointer
			}
			if g.cfg.IsIgnoredShape(memberShapeRef.Shape.ShapeName) {
				continue
			}
			renamedName, _ := crd.InputFieldRename(
				createOp.Name, memberName,
			)
			memberNames := names.New(renamedName)
			memberNames.ModelOrginal = memberName
			if memberName == "Attributes" && g.cfg.UnpacksAttributesMap(crdName) {
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
			if g.cfg.IsIgnoredShape(memberShapeRef.Shape.ShapeName) {
				continue
			}
			if memberShapeRef.Shape == nil {
				return nil, ackmodel.ErrNilShapePointer
			}
			memberNames := names.New(memberName)
			if _, found := crd.SpecFields[memberName]; found {
				// We don't put fields that are already in the Spec struct into
				// the Status struct
				continue
			}
			if memberName == "Attributes" && g.cfg.UnpacksAttributesMap(crdName) {
				continue
			}
			if crd.IsPrimaryARNField(memberName) {
				// We automatically place the primary resource ARN value into
				// the Status.ACKResourceMetadata.ARN field
				continue
			}
			crd.AddStatusField(memberNames, memberShapeRef)
		}

		crds = append(crds, crd)
	}
	sort.Slice(crds, func(i, j int) bool {
		return crds[i].Names.Camel < crds[j].Names.Camel
	})
	g.crds = crds
	return crds, nil
}

// RemoveIgnoredOperations updates CRDOps argument by setting those operations to nil
// that are configured to be ignored in generator config for the AWS service
func (g *Generator) RemoveIgnoredOperations(crdOps *ackmodel.CRDOps) {
	if g.cfg.IsIgnoredOperation(crdOps.Create) {
		crdOps.Create = nil
	}
	if g.cfg.IsIgnoredOperation(crdOps.ReadOne) {
		crdOps.ReadOne = nil
	}
	if g.cfg.IsIgnoredOperation(crdOps.ReadMany) {
		crdOps.ReadMany = nil
	}
	if g.cfg.IsIgnoredOperation(crdOps.Update) {
		crdOps.Update = nil
	}
	if g.cfg.IsIgnoredOperation(crdOps.Delete) {
		crdOps.Delete = nil
	}
	if g.cfg.IsIgnoredOperation(crdOps.GetAttributes) {
		crdOps.GetAttributes = nil
	}
	if g.cfg.IsIgnoredOperation(crdOps.SetAttributes) {
		crdOps.SetAttributes = nil
	}
}

// IsShapeUsedInCRDs returns true if the supplied shape name is a member of amy
// CRD's payloads or those payloads sub-member shapes
func (g *Generator) IsShapeUsedInCRDs(shapeName string) bool {
	crds, _ := g.GetCRDs()
	for _, crd := range crds {
		if crd.HasShapeAsMember(shapeName) {
			return true
		}
	}
	return false
}

// GetTypeDefs returns a slice of `ackmodel.TypeDef` pointers and a map of
// package import information
func (g *Generator) GetTypeDefs() ([]*ackmodel.TypeDef, map[string]string, error) {
	if g.typeDefs != nil {
		return g.typeDefs, g.typeImports, nil
	}

	tdefs := []*ackmodel.TypeDef{}
	// Map, keyed by package import path, with the values being an alias to use
	// for the package
	timports := map[string]string{}
	// Map, keyed by original Shape GoTypeElem(), with the values being a
	// renamed type name (due to conflicting names)
	trenames := map[string]string{}

	payloads := g.SDKAPI.GetPayloads()

	for shapeName, shape := range g.SDKAPI.API.Shapes {
		if util.InStrings(shapeName, payloads) {
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
		if g.cfg.IsIgnoredShape(shapeName) {
			continue
		}
		tdefNames := names.New(shapeName)
		if g.SDKAPI.HasConflictingTypeName(shapeName, g.cfg) {
			tdefNames.Camel += ackmodel.ConflictingNameSuffix
			trenames[shapeName] = tdefNames.Camel
		}

		attrs := map[string]*ackmodel.Attr{}
		for memberName, memberRef := range shape.MemberRefs {
			memberNames := names.New(memberName)
			memberShape := memberRef.Shape
			if g.cfg.IsIgnoredShape(memberShape.ShapeName) {
				continue
			}
			if !g.IsShapeUsedInCRDs(memberShape.ShapeName) {
				continue
			}
			goPkgType := memberRef.Shape.GoTypeWithPkgNameElem()
			if strings.Contains(goPkgType, ".") {
				if strings.HasPrefix(goPkgType, "[]") {
					// For slice types, we just want the element type...
					goPkgType = strings.TrimLeft(goPkgType, "[]")
				}
				if strings.HasPrefix(goPkgType, "map[") {
					// Assuming the map keys are always of type string.
					goPkgType = strings.TrimLeft(goPkgType, "map[string]")
				}
				if strings.HasPrefix(goPkgType, "*") {
					// For slice and map types, the element type might be a
					// pointer to a struct...
					goPkgType = goPkgType[1:]
				}
				pkg := strings.Split(goPkgType, ".")[0]
				if pkg != g.SDKAPI.API.PackageName() {
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
				if g.SDKAPI.HasConflictingTypeName(memberShape.ShapeName, g.cfg) {
					typeNames.Camel += ackmodel.ConflictingNameSuffix
				}
				gt = "*" + typeNames.Camel
			} else if memberShape.Type == "list" {
				// If it's a list type, where the element is a structure, we need to
				// set the GoType to the cleaned-up Camel-cased name
				if memberShape.MemberRef.Shape.Type == "structure" {
					elemType := memberShape.MemberRef.Shape.GoTypeElem()
					typeNames := names.New(elemType)
					if g.SDKAPI.HasConflictingTypeName(elemType, g.cfg) {
						typeNames.Camel += ackmodel.ConflictingNameSuffix
					}
					gt = "[]*" + typeNames.Camel
				}
			} else if memberShape.Type == "map" {
				// If it's a map type, where the value element is a structure,
				// we need to set the GoType to the cleaned-up Camel-cased name
				if memberShape.ValueRef.Shape.Type == "structure" {
					valType := memberShape.ValueRef.Shape.GoTypeElem()
					typeNames := names.New(valType)
					if g.SDKAPI.HasConflictingTypeName(valType, g.cfg) {
						typeNames.Camel += ackmodel.ConflictingNameSuffix
					}
					gt = "[]map[string]*" + typeNames.Camel
				}
			} else if memberShape.Type == "timestamp" {
				// time.Time needs to be converted to apimachinery/metav1.Time
				// otherwise there is no DeepCopy support
				gt = "*metav1.Time"
			}
			attrs[memberName] = ackmodel.NewAttr(memberNames, gt, memberShape)
		}
		if len(attrs) == 0 {
			// Just ignore these...
			continue
		}
		tdefs = append(tdefs, &ackmodel.TypeDef{
			Names: tdefNames,
			Attrs: attrs,
		})
	}
	sort.Slice(tdefs, func(i, j int) bool {
		return tdefs[i].Names.Camel < tdefs[j].Names.Camel
	})
	g.typeDefs = tdefs
	g.typeImports = timports
	g.typeRenames = trenames
	return tdefs, timports, nil
}

// GetEnumDefs returns a slice of pointers to `ackmodel.EnumDef` structs which
// represent string fields whose value is constrained to one or more specific
// string values.
func (g *Generator) GetEnumDefs() ([]*ackmodel.EnumDef, error) {
	edefs := []*ackmodel.EnumDef{}

	for shapeName, shape := range g.SDKAPI.API.Shapes {
		if !shape.IsEnum() {
			continue
		}
		enumNames := names.New(shapeName)
		// Handle name conflicts with top-level CRD.Spec or CRD.Status
		// types
		if g.SDKAPI.HasConflictingTypeName(shapeName, g.cfg) {
			enumNames.Camel += ackmodel.ConflictingNameSuffix
		}
		edef, err := ackmodel.NewEnumDef(enumNames, shape.Enum)
		if err != nil {
			return nil, err
		}
		edefs = append(edefs, edef)
	}
	sort.Slice(edefs, func(i, j int) bool {
		return edefs[i].Names.Camel < edefs[j].Names.Camel
	})
	return edefs, nil
}

// New returns a new Generator struct for a supplied API model.
// Optionally, pass a file path to a generator config file that can be used to
// instruct the code generator how to handle the API properly
func New(
	SDKAPI *ackmodel.SDKAPI,
	apiVersion string,
	configPath string,
	templateBasePath string,
) (*Generator, error) {
	var gc *ackgenconfig.Config
	var err error
	if configPath != "" {
		gc, err = ackgenconfig.New(configPath)
		if err != nil {
			return nil, err
		}
	}

	return &Generator{
		SDKAPI: SDKAPI,
		// TODO(jaypipes): Handle cases where service alias and service ID
		// don't match (Step Functions)
		serviceAlias:     SDKAPI.ServiceID(),
		apiVersion:       apiVersion,
		templateBasePath: templateBasePath,
		cfg:              gc,
	}, nil
}
