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

package ack

import (
	"path/filepath"
	"strings"
	ttpl "text/template"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
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
	apisIncludePaths = []string{
		"boilerplate.go.tpl",
		"apis/enum_def.go.tpl",
		"apis/type_def.go.tpl",
	}
	apisCopyPaths = []string{}
	apisFuncMap   = ttpl.FuncMap{
		"Join": strings.Join,
	}
)

// APIs returns a pointer to a TemplateSet containing all the templates for
// generating ACK service controller's apis/ contents
func APIs(
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
		apisIncludePaths,
		apisCopyPaths,
		apisFuncMap,
	)

	metaVars := g.MetaVars()
	apiVars := &templateAPIVars{
		metaVars,
		enumDefs,
		typeDefs,
		typeImports,
	}
	for _, path := range apisTemplatePaths {
		outPath := strings.TrimSuffix(filepath.Base(path), ".tpl")
		if err = ts.Add(outPath, path, apiVars); err != nil {
			return nil, err
		}
	}

	for _, crd := range crds {
		crdFileName := strcase.ToSnake(crd.Kind) + ".go"
		crdVars := &templateCRDVars{
			metaVars,
			crd,
		}
		if err = ts.Add(crdFileName, "apis/crd.go.tpl", crdVars); err != nil {
			return nil, err
		}
	}
	return ts, nil
}

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
