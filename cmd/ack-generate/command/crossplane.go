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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/model"
)

// crossplaneCmd is the command that generates Crossplane API types
var crossplaneCmd = &cobra.Command{
	Use:   "crossplane <service>",
	Short: "Generate Crossplane Provider",
}

var crossplaneAPIsCmd = &cobra.Command{
	Use:   "apis <service>",
	Short: "Generate Crossplane-compatible Kubernetes API type definitions for a service",
	RunE:  generateCrossplaneAPIs,
}

var providerDir string

func init() {
	crossplaneCmd.PersistentFlags().StringVar(
		&providerDir, "provider-dir", ".", "the directory of the Crossplane provider",
	)
	crossplaneCmd.AddCommand(crossplaneAPIsCmd)
	rootCmd.AddCommand(crossplaneCmd)
}

// generateCrossplaneAPIs generates the Go files for Crossplane-compatible
// resources in the AWS service API.
func generateCrossplaneAPIs(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please specify the service alias for the AWS service API to generate")
	}
	optTemplatesDir = filepath.Join(optTemplatesDir, "crossplane")
	svcAlias := strings.ToLower(args[0])
	if !optDryRun {
		apisVersionPath = filepath.Join(providerDir, "apis", svcAlias, optGenVersion)
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
		return err
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
