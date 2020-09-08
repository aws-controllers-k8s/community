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
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	ttpl "text/template"

	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
)

var (
	goTemplatePaths = []string{
		"apis/crd",
		"apis/doc",
		"apis/enums",
		"apis/groupversion_info",
		"apis/types",
		"cmd/controller/main",
		"pkg/crd_descriptor",
		"pkg/crd_identifiers",
		"pkg/crd_manager",
		"pkg/crd_manager_factory",
		"pkg/crd_resource",
		"pkg/crd_sdk",
		"pkg/resource_registry",
	}
	yamlTemplatePaths = []string{
		"config/controller/deployment",
		"config/controller/kustomization",
		"config/default/kustomization",
		"config/rbac/cluster-role-binding",
		"config/rbac/kustomization",
	}
	goTemplateFuncMap = ttpl.FuncMap{
		"ToLower": strings.ToLower,
		"ResourceExceptionCode": func(r *ackmodel.CRD, httpStatusCode int) string {
			return r.ExceptionCode(httpStatusCode)
		},
		"GoCodeSetReadOneOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeGet, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadOneInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeGet, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadManyOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeList, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadManyInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeList, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeGetAttributesSetInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeGetAttributesSetInput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeGetAttributesSetOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeGetAttributesSetOutput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetCreateOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetCreateInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetUpdateOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetUpdateInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetDeleteInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeDelete, sourceVarName, targetVarName, indentLevel)
		},
		"Empty": func(subject string) bool {
			return strings.TrimSpace(subject) == ""
		},
		"GoCodeRequiredStatusFieldsMissingFromReadOneInput": func(r *ackmodel.CRD, koVarName string, indentLevel int) string {
			return r.GoCodeRequiredStatusFieldsMissingFromReadOneInput(koVarName, indentLevel)
		},
	}
)

func errUnknownResource(resource string) error {
	return errors.New("unknown resource: " + resource)
}

func errUnknownTemplate(target string) error {
	return errors.New("unknown template: " + target)
}

// templateMetaVars contains template variables that most templates need access to
// that describe the service alias, its package name, etc
type templateMetaVars struct {
	// ServiceAlias contains the exact string used to identify the AWS service
	// API in the aws-sdk-go's models/apis/ directory. Note that some APIs this
	// alias does not match the ServiceID. e.g. The AWS Step Functions API has
	// a ServiceID of "SFN" and a service alias of "states"...
	ServiceAlias string
	// ServiceID is the exact string that appears in the AWS service API's
	// api-2.json descriptor file under `metadata.serviceId`
	ServiceID string
	// ServiceIDClean is the ServiceID lowercased and stripped of any
	// non-alphanumeric characters
	ServiceIDClean string
	// APIVersion contains the version of the Kubernetes API resources, e.g.
	// "v1alpha1"
	APIVersion string
	// APIGroup contains the normalized name of the Kubernetes APIGroup used
	// for custom resources, e.g. "sns.services.k8s.aws" or
	// "sfn.services.k8s.aws"
	APIGroup string
	// SDKAPIInterfaceTypeName is the name of the interface type used by the
	// aws-sdk-go services/$SERVICE/api.go file
	SDKAPIInterfaceTypeName string
}

// TemplateAPIVars contains template variables for templates that output Go
// code in the /services/$SERVICE/apis/$API_VERSION directory
type templateAPIVars struct {
	templateMetaVars
	EnumDefs []*ackmodel.EnumDef
	TypeDefs []*ackmodel.TypeDef
	Imports  map[string]string
}

// TemplateCRDVars contains template variables for the template that outputs Go
// code for a single top-level resource's API definition
type templateCRDVars struct {
	templateMetaVars
	CRD *ackmodel.CRD
}

// TemplateCmdVars contains template variables for the template that outputs Go
// code for a single top-level resource's API definition
type templateCmdVars struct {
	templateMetaVars
	SnakeCasedCRDNames []string
}

// templateMetaVars returns a templateMetaVars struct populated with metadata
// about the AWS service API
func (g *Generator) templateMetaVars() templateMetaVars {
	return templateMetaVars{
		ServiceAlias:            g.serviceAlias,
		ServiceID:               g.SDKAPI.ServiceID(),
		ServiceIDClean:          g.SDKAPI.ServiceIDClean(),
		APIGroup:                g.SDKAPI.APIGroup(),
		APIVersion:              g.apiVersion,
		SDKAPIInterfaceTypeName: g.SDKAPI.SDKAPIInterfaceTypeName(),
	}
}

// templateAPIVars returns a templateAPIVars struct populated with information
// used to generate the Kubernetes API types for the AWS service API
func (g *Generator) templateAPIVars() (*templateAPIVars, error) {
	enumDefs, err := g.GetEnumDefs()
	if err != nil {
		return nil, err
	}
	typeDefs, typeImports, err := g.GetTypeDefs()
	if err != nil {
		return nil, err
	}
	return &templateAPIVars{
		g.templateMetaVars(),
		enumDefs,
		typeDefs,
		typeImports,
	}, nil
}

// templateCRDVars returns a templateCRDVars struct populated with information
// for a particular top-level resource
func (g *Generator) templateCRDVars(crdName string) (*templateCRDVars, error) {
	crds, err := g.GetCRDs()
	if err != nil {
		return nil, err
	}

	for _, crd := range crds {
		if crd.Names.Original == crdName {
			return &templateCRDVars{
				g.templateMetaVars(),
				crd,
			}, nil
		}
	}
	return nil, errUnknownResource(crdName)
}

// templateCmdVars returns a templateCmdVars struct populated with information
// for files in a service controller's cmd/ directory
func (g *Generator) templateCmdVars() (*templateCmdVars, error) {
	crds, err := g.GetCRDs()
	if err != nil {
		return nil, err
	}
	// convert CRD names into snake_case to use for package import
	snakeCasedCRDNames := make([]string, 0)
	for _, crd := range crds {
		snakeCasedCRDNames = append(snakeCasedCRDNames, crd.Names.Snake)
	}
	return &templateCmdVars{
		g.templateMetaVars(),
		snakeCasedCRDNames,
	}, nil
}

// initTemplates initializes the templates for generating Kubernetes API
// type files and the service controller Go code files
func (g *Generator) initTemplates() error {
	if g.templates != nil {
		return nil
	}
	tpls := map[string]*ttpl.Template{}
	for _, path := range goTemplatePaths {
		tplPath := filepath.Join(g.templateBasePath, path+".go.tpl")
		tplContents, err := ioutil.ReadFile(tplPath)
		if err != nil {
			return err
		}
		t := ttpl.New(path)
		t = t.Funcs(goTemplateFuncMap)
		t, err = t.Parse(string(tplContents))
		if err != nil {
			return err
		}
		includes := []string{
			"boilerplate",
			"apis/enum_def",
			"apis/type_def",
		}
		for _, include := range includes {
			if t, err = IncludeTemplate(t, g.templateBasePath, include); err != nil {
				return err
			}
		}
		tpls[path] = t
	}
	for _, path := range yamlTemplatePaths {
		tplPath := filepath.Join(g.templateBasePath, path+".yaml.tpl")
		tplContents, err := ioutil.ReadFile(tplPath)
		if err != nil {
			return err
		}
		t := ttpl.New(path)
		t, err = t.Parse(string(tplContents))
		if err != nil {
			return err
		}
		tpls[path] = t
	}
	g.templates = tpls
	return nil
}

// GenerateAPIFile returns a byte buffer containing the output of an executed
// template for the Kubernetes API type definitions for a service API
func (g *Generator) GenerateAPIFile(
	// target is the thing to generate, e.g. "doc" or "groupversion_info"
	target string,
) (*bytes.Buffer, error) {
	if err := g.initTemplates(); err != nil {
		return nil, err
	}
	targetPath := "apis/" + target
	t, found := g.templates[targetPath]
	if !found {
		return nil, errUnknownTemplate(targetPath)
	}
	vars, err := g.templateAPIVars()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err = t.Execute(&b, vars); err != nil {
		return nil, err
	}
	return &b, nil
}

// GenerateCRDFile returns a byte buffer containing the output of an executed
// template for a particular top-level resource/CRD
func (g *Generator) GenerateCRDFile(
	crdName string,
) (*bytes.Buffer, error) {
	if err := g.initTemplates(); err != nil {
		return nil, err
	}
	t := g.templates["apis/crd"]
	vars, err := g.templateCRDVars(crdName)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, vars); err != nil {
		return nil, err
	}
	return &b, nil
}

// GenerateCmdControllerMainFile returns a byte buffer containing the output of
// an executed template for a service controller's main.go file
func (g *Generator) GenerateCmdControllerMainFile() (*bytes.Buffer, error) {
	if err := g.initTemplates(); err != nil {
		return nil, err
	}
	t := g.templates["cmd/controller/main"]
	vars, err := g.templateCmdVars()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, vars); err != nil {
		return nil, err
	}
	return &b, nil
}

// GenerateResourceRegistryFile returns a byte buffer containing the output of
// an executed template containing the resource registry for the service
// controller
func (g *Generator) GenerateResourceRegistryFile() (*bytes.Buffer, error) {
	if err := g.initTemplates(); err != nil {
		return nil, err
	}
	t := g.templates["pkg/resource_registry"]
	vars := g.templateMetaVars()
	var b bytes.Buffer
	if err := t.Execute(&b, vars); err != nil {
		return nil, err
	}
	return &b, nil
}

// GenerateCRDResourcePackageFile returns a byte buffer containing the output of
// an executed template containing a file in a specific CRD's resource package
func (g *Generator) GenerateCRDResourcePackageFile(
	crdName string,
	target string,
) (*bytes.Buffer, error) {
	if err := g.initTemplates(); err != nil {
		return nil, err
	}
	targetPath := "pkg/crd_" + target
	t, found := g.templates[targetPath]
	if !found {
		return nil, errUnknownTemplate(targetPath)
	}
	vars, err := g.templateCRDVars(crdName)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err = t.Execute(&b, vars); err != nil {
		return nil, err
	}
	return &b, nil
}

// GenerateConfigYAMLFile returns a byte buffer containing the output of an
// executed template for the Kubernetes YAML manifest/configuration file
func (g *Generator) GenerateConfigYAMLFile(
	// target is the thing to generate, e.g. "controller/deployment" or
	// "default/kustomization"
	target string,
) (*bytes.Buffer, error) {
	if err := g.initTemplates(); err != nil {
		return nil, err
	}
	targetPath := "config/" + target
	t, found := g.templates[targetPath]
	if !found {
		return nil, errUnknownTemplate(targetPath)
	}
	vars := g.templateMetaVars()
	var b bytes.Buffer
	if err := t.Execute(&b, vars); err != nil {
		return nil, err
	}
	return &b, nil
}

// IncludeTemplate includes a template into a supplied Template struct
func IncludeTemplate(t *ttpl.Template, tplDir string, tplName string) (*ttpl.Template, error) {
	tplPath := filepath.Join(tplDir, tplName+".go.tpl")
	tplContents, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return nil, err
	}
	if t, err = t.Parse(string(tplContents)); err != nil {
		return nil, err
	}
	return t, nil
}
