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
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/schema"
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
	optOutputPath string
)

// apiCmd is the command that generates service API types
var typesCmd = &cobra.Command{
	Use:   "types <file>",
	Short: "Generates Go files containing type definitions and API machinery base initialization",
	RunE:  generateTypes,
}

func init() {
	typesCmd.PersistentFlags().StringVarP(
		&optGenVersion, "version", "v", "v1alpha1", "the resource API Version to use when generating types",
	)
	typesCmd.PersistentFlags().StringVarP(
		&optOutputPath, "output", "o", "", "path to output directory to send generated files. If empty, outputs all files to stdout",
	)
	rootCmd.AddCommand(typesCmd)
}

// ensureOutputDir makes sure that the target output directory exists and
// returns whether the directory already existed. If the output path has been
// set to stdout, this is a noop and returns false.
func ensureOutputDir() (bool, error) {
	if optOutputPath == "" {
		return false, nil
	}
	fi, err := os.Stat(optOutputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, os.MkdirAll(optOutputPath, os.ModePerm)
		} else {
			return false, err
		}
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("expected %s to be a directory.", optOutputPath)
	}
	// Make sure the directory is writeable by the calling user
	testPath := filepath.Join(optOutputPath, "test")
	f, err := os.Create(testPath)
	if err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("%s is not a writeable directory.", optOutputPath)
		}
		return true, err
	}
	f.Close()
	os.Remove(testPath)
	return true, nil
}

// generateTypes generates the Go files for each resource in the AWS service
// API.
func generateTypes(cmd *cobra.Command, args []string) error {
	sh, err := getSchemaHelper(args)
	if err != nil {
		return err
	}
	crds, err := sh.GetCRDs()
	if err != nil {
		return err
	}
	typeDefs, err := sh.GetTypeDefs()
	if err != nil {
		return err
	}
	enumDefs, err := sh.GetEnumDefs()
	if err != nil {
		return err
	}

	if _, err := ensureOutputDir(); err != nil {
		return err
	}

	if err = writeDocGo(sh); err != nil {
		return err
	}

	if err = writeGroupVersionInfoGo(sh); err != nil {
		return err
	}

	if err = writeEnumsGo(enumDefs); err != nil {
		return err
	}

	if err = writeTypesGo(typeDefs); err != nil {
		return err
	}

	for _, crd := range crds {
		if err = writeCRDGo(crd); err != nil {
			return err
		}
	}
	return nil
}

func writeDocGo(sh *schema.Helper) error {
	var b bytes.Buffer
	apiGroup := sh.GetAPIGroup()
	vars := &template.DocTemplateVars{
		APIVersion: optGenVersion,
		APIGroup:   apiGroup,
	}
	tpl, err := template.NewDocTemplate(templatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optOutputPath == "" {
		fmt.Println("============================= doc.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	} else {
		path := filepath.Join(optOutputPath, "doc.go")
		return ioutil.WriteFile(path, b.Bytes(), 0666)
	}
}

func writeGroupVersionInfoGo(sh *schema.Helper) error {
	var b bytes.Buffer
	apiGroup := sh.GetAPIGroup()
	vars := &template.GroupVersionInfoTemplateVars{
		APIVersion: optGenVersion,
		APIGroup:   apiGroup,
	}
	tpl, err := template.NewGroupVersionInfoTemplate(templatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optOutputPath == "" {
		fmt.Println("============================= groupversion_info.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	} else {
		path := filepath.Join(optOutputPath, "groupversion_info.go")
		return ioutil.WriteFile(path, b.Bytes(), 0666)
	}
}

func writeEnumsGo(
	enumDefs []*model.EnumDef,
) error {
	if len(enumDefs) == 0 {
		return nil
	}
	vars := &template.EnumsTemplateVars{
		APIVersion: optGenVersion,
		EnumDefs:   enumDefs,
	}
	var b bytes.Buffer
	tpl, err := template.NewEnumsTemplate(templatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optOutputPath == "" {
		fmt.Println("============================= enums.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	} else {
		path := filepath.Join(optOutputPath, "enums.go")
		return ioutil.WriteFile(path, b.Bytes(), 0666)
	}
}

func writeTypesGo(
	typeDefs []*model.TypeDef,
) error {
	vars := &template.TypesTemplateVars{
		APIVersion: optGenVersion,
		TypeDefs:   typeDefs,
	}
	var b bytes.Buffer
	tpl, err := template.NewTypesTemplate(templatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optOutputPath == "" {
		fmt.Println("============================= types.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	} else {
		path := filepath.Join(optOutputPath, "types.go")
		return ioutil.WriteFile(path, b.Bytes(), 0666)
	}
}

func writeCRDGo(crd *model.CRD) error {
	vars := &template.CRDTemplateVars{
		APIVersion: optGenVersion,
		CRD:        crd,
	}
	var b bytes.Buffer
	tpl, err := template.NewCRDTemplate(templatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	crdFileName := strcase.ToSnake(crd.Kind) + ".go"
	if optOutputPath == "" {
		fmt.Printf("============================= %s ======================================\n", crdFileName)
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	} else {
		path := filepath.Join(optOutputPath, crdFileName)
		return ioutil.WriteFile(path, b.Bytes(), 0666)
	}
}

// getAPI returns a schema.Helper object representing the API from
// either STDIN or an input file
func getSchemaHelper(args []string) (*schema.Helper, error) {
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
	return schema.NewHelper(api), nil
}
