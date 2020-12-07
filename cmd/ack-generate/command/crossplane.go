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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/generate/crossplane"
	"github.com/aws/aws-controllers-k8s/pkg/generate/templateset"
	"github.com/aws/aws-controllers-k8s/pkg/model"
)

// crossplaneCmd is the command that generates Initialize API types
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
	cfgPath := ""
	gcPath := filepath.Join(providerDir, "apis", svcAlias, optGenVersion, "generator-config.yaml")
	if _, err := os.Stat(gcPath); !os.IsNotExist(err) {
		cfgPath = gcPath
	}
	g, err := generate.New(
		sdkAPI, optGenVersion, cfgPath, crossplane.DefaultConfig,
	)
	if err != nil {
		return err
	}
	ts := templateset.New(optTemplatesDir, crossplane.IncludePaths, crossplane.CopyPaths, crossplane.TemplateFuncs)
	generation := crossplane.NewGeneration(g, ts,
		crossplane.WithInitializer(crossplane.AddAPIFiles),
		crossplane.WithInitializer(crossplane.AddCRDFiles),
		crossplane.WithInitializer(crossplane.AddControllerFiles),
	)
	output, err := generation.Run()
	if err != nil {
		return err
	}

	for path, contents := range output {
		if optDryRun {
			fmt.Printf("============================= %s ======================================\n", path)
			fmt.Println(strings.TrimSpace(contents.String()))
			continue
		}
		outPath := filepath.Join(providerDir, path)
		outDir := filepath.Dir(outPath)
		if _, err := ensureDir(outDir); err != nil {
			return err
		}
		// TODO(muvaf): Hooks file should be generated only once as boilerplate.
		// We will revisit to make it so lean that we don't need to generate it.
		if strings.Contains(outPath, "hooks.go") {
			if _, err := os.Stat(outPath); !os.IsNotExist(err) {
				continue
			}
		}
		if err = ioutil.WriteFile(outPath, contents.Bytes(), 0666); err != nil {
			return err
		}
	}
	apiPath := filepath.Join(providerDir, "apis", svcAlias, optGenVersion)
	controllerPath := filepath.Join(providerDir, "pkg", "controller", svcAlias)
	// TODO(muvaf): goimports don't allow to be included as a library. Make sure
	// goimports binary exists.
	if err := exec.Command("goimports", "-w", apiPath, controllerPath).Run(); err != nil {
		return errors.Wrap(err, "cannot run goimports")
	}
	return nil
}
