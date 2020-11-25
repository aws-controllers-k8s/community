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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/pkg/errors"

	"github.com/aws/aws-controllers-k8s/pkg/model"
)

func WithGeneratorConfigFilePath(path string) GenerationOption {
	return func(g *Generation) {
		g.GeneratorConfigFilePath = path
	}
}

type GenerationOption func(*Generation)

type APIFileGenerator interface {
	Generate(g *generate.Generator, apiPath string) error
}

type ControllerGenerator interface {
	Generate(g *generate.Generator, controllerPath string) error
}

func NewGeneration(serviceAlias, apiVersion, providerDirectory, templatePath string, sdkAPI *model.SDKAPI, opts ...GenerationOption) *Generation {
	g := &Generation{
		ServiceAlias:      serviceAlias,
		APIVersion:        apiVersion,
		ProviderDirectory: providerDirectory,
		TemplateBasePath:  templatePath,
		SDKAPI:            sdkAPI,
		apis: APIFileGeneratorChain{
			GenerateCRDFiles,
			GenerateTypesFile,
			GenerateEnumsFile,
			GenerateGroupVersionInfoFile,
			GenerateDocFile,
		},
		controller: ControllerGeneratorChain{
			GenerateController,
			GenerateConversions,
			GenerateHooksBoilerplate,
		},
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

type Generation struct {
	ServiceAlias            string
	APIVersion              string
	ProviderDirectory       string
	TemplateBasePath        string
	GeneratorConfigFilePath string
	SDKAPI                  *model.SDKAPI

	apis       APIFileGenerator
	controller ControllerGenerator
}

func (g *Generation) Generate() error {
	apiPath := filepath.Join(g.ProviderDirectory, "apis", g.ServiceAlias, g.APIVersion)
	controllerPath := filepath.Join(g.ProviderDirectory, "pkg", "controller", g.ServiceAlias)
	o, err := generate.New(g.SDKAPI, g.APIVersion, g.GeneratorConfigFilePath, g.TemplateBasePath, DefaultConfig)
	if err != nil {
		return errors.Wrap(err, "cannot create a new ACK Generator")
	}

	// TODO(muvaf): Controllers are aware what CRD is used but APIs are not, so,
	// we have to include them all in the same folder.
	if err := os.MkdirAll(apiPath, os.ModePerm); err != nil {
		return errors.Wrap(err, "cannot create api folder")
	}
	// TODO(muvaf): ACK generator requires all template files to be present during
	// initTemplates even though we don't use them.
	if err := g.apis.Generate(o, apiPath); err != nil {
		return errors.Wrap(err, "cannot generate API files")
	}
	if err := g.controller.Generate(o, controllerPath); err != nil {
		return errors.Wrap(err, "cannot generate controller files")
	}
	// TODO(muvaf): goimports don't allow to be included as a library. Make sure
	// goimports binary exists.
	if err := exec.Command("goimports", "-w", apiPath, controllerPath).Run(); err != nil {
		return errors.Wrap(err, "cannot run goimports")
	}
	return nil
}
