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

package crossplane

import (
	"path/filepath"
	"strings"
	ttpl "text/template"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/generate/code"
	"github.com/aws/aws-controllers-k8s/pkg/generate/templateset"
	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/iancoleman/strcase"
)

var (
	apisTemplatePaths = []string{
		"apis/doc.go.tpl",
		"apis/enums.go.tpl",
		"apis/groupversion_info.go.tpl",
		"apis/types.go.tpl",
	}
	includePaths = []string{
		"boilerplate.go.tpl",
		"apis/enum_def.go.tpl",
		"apis/type_def.go.tpl",
		"pkg/sdk_find_read_one.go.tpl",
		"pkg/sdk_find_read_many.go.tpl",
	}
	copyPaths = []string{}
	funcMap   = ttpl.FuncMap{
		"ToLower": strings.ToLower,
		"ResourceExceptionCode": func(r *ackmodel.CRD, httpStatusCode int) string {
			return r.ExceptionCode(httpStatusCode)
		},
		"GoCodeSetExceptionMessagePrefixCheck": func(r *ackmodel.CRD, httpStatusCode int) string {
			return code.CheckExceptionMessagePrefix(r.Config(), r, httpStatusCode)
		},
		"GoCodeSetReadOneOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return code.SetResource(r.Config(), r, ackmodel.OpTypeGet, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetReadOneInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return code.SetSDK(r.Config(), r, ackmodel.OpTypeGet, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadManyOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return code.SetResource(r.Config(), r, ackmodel.OpTypeList, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetReadManyInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return code.SetSDK(r.Config(), r, ackmodel.OpTypeList, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetCreateOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return code.SetResource(r.Config(), r, ackmodel.OpTypeCreate, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetCreateInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return code.SetSDK(r.Config(), r, ackmodel.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetUpdateInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return code.SetSDK(r.Config(), r, ackmodel.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetDeleteInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return code.SetSDK(r.Config(), r, ackmodel.OpTypeDelete, sourceVarName, targetVarName, indentLevel)
		},
		"Empty": func(subject string) bool {
			return strings.TrimSpace(subject) == ""
		},
	}
)

// templateAPIVars contains template variables for templates that output Go
// code in the /services/$SERVICE/apis/$API_VERSION directory
type templateAPIVars struct {
	templateset.MetaVars
	EnumDefs []*ackmodel.EnumDef
	TypeDefs []*ackmodel.TypeDef
	Imports  map[string]string
}

// templateCRDVars contains template variables for the template that outputs Go
// code for a single top-level resource's API definition
type templateCRDVars struct {
	templateset.MetaVars
	CRD *ackmodel.CRD
}

// Crossplane returns a pointer to a TemplateSet containing all the templates for
// generating Crossplane API types and controller code for an AWS service API
func Crossplane(
	g *generate.Generator,
	templateBasePath string,
) (*templateset.TemplateSet, error) {
	enumDefs, err := g.GetEnumDefs()
	if err != nil {
		return nil, err
	}
	typeDefs, typeImports, err := g.GetTypeDefs()
	if err != nil {
		return nil, err
	}
	crds, err := g.GetCRDs()
	if err != nil {
		return nil, err
	}

	ts := templateset.New(
		templateBasePath,
		includePaths,
		copyPaths,
		funcMap,
	)

	metaVars := g.MetaVars()

	// First add all the CRDs and API types
	apiVars := &templateAPIVars{
		metaVars,
		enumDefs,
		typeDefs,
		typeImports,
	}
	for _, path := range apisTemplatePaths {
		outPath := filepath.Join(
			"apis",
			metaVars.ServiceIDClean,
			metaVars.APIVersion,
			"zz_"+strings.TrimSuffix(filepath.Base(path), ".tpl"),
		)
		if err = ts.Add(outPath, path, apiVars); err != nil {
			return nil, err
		}
	}
	for _, crd := range crds {
		crdFileName := filepath.Join(
			"apis", metaVars.ServiceIDClean, metaVars.APIVersion,
			"zz_"+strcase.ToSnake(crd.Kind)+".go",
		)
		crdVars := &templateCRDVars{
			metaVars,
			crd,
		}
		if err = ts.Add(crdFileName, "apis/crd.go.tpl", crdVars); err != nil {
			return nil, err
		}
	}

	// Next add the controller package for each CRD
	targets := []string{
		"controller.go.tpl",
		"conversions.go.tpl",
	}
	for _, crd := range crds {
		for _, target := range targets {
			outPath := filepath.Join(
				"pkg", "controller", metaVars.ServiceIDClean, crd.Names.Lower,
				"zz_"+strings.TrimSuffix(filepath.Base(target), ".tpl"),
			)
			crdVars := &templateCRDVars{
				metaVars,
				crd,
			}
			if err = ts.Add(outPath, "pkg/"+target, crdVars); err != nil {
				return nil, err
			}
		}
	}

	return ts, nil
}
