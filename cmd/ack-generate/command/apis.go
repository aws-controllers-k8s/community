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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/model"
)

type contentType int

const (
	ctUnknown contentType = iota
	ctJSON
	ctYAML
)

var (
	optGenVersion     string
	optAPIsInputPath  string
	optAPIsOutputPath string
	apisVersionPath   string
)

// apiCmd is the command that generates service API types
var apisCmd = &cobra.Command{
	Use:   "apis <service>",
	Short: "Generate Kubernetes API type definitions for an AWS service API",
	RunE:  generateAPIs,
}

func init() {
	apisCmd.PersistentFlags().StringVar(
		&optGenVersion, "version", "v1alpha1", "the resource API Version to use when generating API infrastructure and type definitions",
	)
	apisCmd.PersistentFlags().StringVarP(
		&optAPIsOutputPath, "output", "o", "", "path to directory for service controller to create generated files. Defaults to "+optServicesDir+"/$service",
	)
	rootCmd.AddCommand(apisCmd)
}

// generateAPIs generates the Go files for each resource in the AWS service
// API.
func generateAPIs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please specify the service alias for the AWS service API to generate")
	}
	svcAlias := strings.ToLower(args[0])
	if optAPIsOutputPath == "" {
		optAPIsOutputPath = filepath.Join(optServicesDir)
	}
	if !optDryRun {
		apisVersionPath = filepath.Join(optAPIsOutputPath, svcAlias, "apis", optGenVersion)
		if _, err := ensureDir(apisVersionPath); err != nil {
			return err
		}
	}
	if err := ensureSDKRepo(optCacheDir); err != nil {
		return err
	}
	sdkHelper := model.NewSDKHelper(sdkDir)
	sdkAPI, err := sdkHelper.API(svcAlias)
	if err != nil {
		newSvcAlias, err := FallBackFindServiceID(sdkDir, svcAlias)
		if err != nil {
			return err
		}
		sdkAPI, err = sdkHelper.API(newSvcAlias) // retry with serviceID
		if err != nil {
			return err
		}
	}
	g, err := generate.New(
		sdkAPI, optGenVersion, optGeneratorConfigPath, optTemplatesDir,
	)
	if err != nil {
		return err
	}

	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	typeDefs, _, err := g.GetTypeDefs()
	if err != nil {
		return err
	}
	enumDefs, err := g.GetEnumDefs()
	if err != nil {
		return err
	}

	if err = writeDocGo(g); err != nil {
		return err
	}

	if err = writeGroupVersionInfoGo(g); err != nil {
		return err
	}

	if err = writeEnumsGo(g, enumDefs); err != nil {
		return err
	}

	if err = writeTypesGo(g, typeDefs); err != nil {
		return err
	}

	for _, crd := range crds {
		if err = writeCRDGo(g, crd); err != nil {
			return err
		}
	}
	return nil
}

func writeDocGo(g *generate.Generator) error {
	b, err := g.GenerateAPIFile("doc")
	if err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= doc.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(apisVersionPath, "doc.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeGroupVersionInfoGo(g *generate.Generator) error {
	b, err := g.GenerateAPIFile("groupversion_info")
	if err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= groupversion_info.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(apisVersionPath, "groupversion_info.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeEnumsGo(
	g *generate.Generator,
	enumDefs []*model.EnumDef,
) error {
	if len(enumDefs) == 0 {
		return nil
	}
	b, err := g.GenerateAPIFile("enums")
	if err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= enums.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(apisVersionPath, "enums.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeTypesGo(
	g *generate.Generator,
	typeDefs []*model.TypeDef,
) error {
	if len(typeDefs) == 0 {
		return nil
	}
	b, err := g.GenerateAPIFile("types")
	if err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= types.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(apisVersionPath, "types.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeCRDGo(
	g *generate.Generator,
	crd *model.CRD,
) error {
	b, err := g.GenerateCRDFile(crd.Names.Original)
	if err != nil {
		return err
	}
	crdFileName := strcase.ToSnake(crd.Kind) + ".go"
	if optDryRun {
		fmt.Printf("============================= %s ======================================\n", crdFileName)
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(apisVersionPath, crdFileName)
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}
