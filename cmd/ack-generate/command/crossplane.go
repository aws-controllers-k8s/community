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
	"github.com/aws/aws-controllers-k8s/pkg/model"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-controllers-k8s/pkg/crossplane"

	"github.com/spf13/cobra"
)

// crossplaneCmd is the command that generates Crossplane API types
var crossplaneCmd = &cobra.Command{
	Use:   "crossplane <service>",
	Short: "Generate Crossplane Provider",
	RunE:  generateCrossplane,
}

var providerDir string

func init() {
	crossplaneCmd.PersistentFlags().StringVar(
		&providerDir, "provider-dir", ".", "the directory of the Crossplane provider",
	)
	rootCmd.AddCommand(crossplaneCmd)
}

func generateCrossplane(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please specify the service alias for the AWS service API to generate")
	}
	if err := ensureSDKRepo(optCacheDir); err != nil {
		return err
	}
	optTemplatesDir = filepath.Join(optTemplatesDir, "crossplane")
	svcAlias := strings.ToLower(args[0])
	sdkHelper := model.NewSDKHelper(sdkDir)
	sdkHelper.APIGroupSuffix = "aws.crossplane.io"
	sdkAPI, err := sdkHelper.API(svcAlias)
	if err != nil {
		newSvcAlias, err := FallBackFindServiceID(sdkDir, svcAlias)
		if err != nil {
			return err
		}
		sdkAPI, err = sdkHelper.API(newSvcAlias) // retry with serviceID
		if err != nil {
			return fmt.Errorf("cannot get the API model for service %s", svcAlias)
		}
	}
	var opts []crossplane.GenerationOption
	cfgPath := filepath.Join(providerDir, "apis", svcAlias, optGenVersion, "generator-config.yaml")
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		opts = append(opts, crossplane.WithGeneratorConfigFilePath(cfgPath))
	}
	g := crossplane.NewGeneration(svcAlias, optGenVersion, providerDir, optTemplatesDir, sdkAPI, opts...)
	return g.Generate()
}
