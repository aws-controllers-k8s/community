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

package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/aws/aws-service-operator-k8s/pkg/resource"
	"github.com/aws/aws-service-operator-k8s/pkg/template"
)

type contentType int

const (
	ctUnknown contentType = iota
	ctJSON
	ctYAML
)

var (
	optGenVersion string
	optResource   string
)

// typesCmd is the command that generates service API types
var typesCmd = &cobra.Command{
	Use:   "types <file>",
	Short: "Generate a new AWS service API type collection from an OpenAPI3 descriptor document",
	RunE:  generateTypes,
}

func init() {
	typesCmd.PersistentFlags().StringVarP(
		&optGenVersion, "version", "v", "v1alpha1", "the resource API Version to use when generating types",
	)
	typesCmd.PersistentFlags().StringVarP(
		&optResource, "resource", "r", "", "only generate type for the specified resource",
	)
	rootCmd.AddCommand(typesCmd)
}

// generateTypes generates the Go files for each resource in the AWS service
// API.
func generateTypes(cmd *cobra.Command, args []string) error {
	api, err := getAPI(args)
	if err != nil {
		return err
	}
	resources, err := resource.ResourcesFromAPI(api)
	if err != nil {
		return err
	}
	structDefs, err := resource.StructDefsFromAPI(api, resources)
	if err != nil {
		return err
	}
	filtered := []*resource.Resource{}
	for _, res := range resources {
		if optResource != "" {
			if strings.ToLower(optResource) != strings.ToLower(res.Kind) {
				continue
			}
		}
		filtered = append(filtered, res)
	}
	vars := &template.TypesTemplateVars{
		Version:    optGenVersion,
		Resources:  filtered,
		StructDefs: structDefs,
	}
	var b bytes.Buffer
	tpl, err := template.NewTypesTemplate(templateDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	fmt.Println(strings.TrimSpace(b.String()))
	return nil
}

// getAPI returns an OpenAPI3 Swagger object representing the API from
// either STDIN or an input file
func getAPI(args []string) (*openapi3.Swagger, error) {
	var b []byte
	var err error
	contentType := ctUnknown
	switch len(args) {
	case 0:
		if b, err = ioutil.ReadAll(os.Stdin); err != nil {
			return nil, fmt.Errorf("expected OpenAPI3 descriptor document either via STDIN or path argument.")
		}
	case 1:
		fp := filepath.Clean(args[0])
		ext := filepath.Ext(fp)
		switch ext {
		case "json":
			contentType = ctJSON
		case "yaml", "yml":
			contentType = ctYAML
		}
		if b, err = ioutil.ReadFile(fp); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("expected OpenAPI3 descriptor document either via STDIN or path argument.")
	}

	if len(b) < 2 {
		return nil, fmt.Errorf("expected OpenAPI3 descriptor document but got '%s'.", string(b))
	}

	var jsonb []byte = b

	// First get our supplied document into JSON format
	if contentType == ctYAML || (contentType == ctUnknown && b[0] != '{' && b[0] != '[') {
		// It's probably YAML, so try decoding to YAML first and fall back to
		// JSON below
		if jsonb, err = yaml.YAMLToJSON(b); err != nil {
			jsonb = b
		}
	}

	api, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(jsonb)
	if err != nil {
		return nil, err
	}
	return api, nil
}
