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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	appName      = "ack-generate"
	appShortDesc = "ack-generate - generate AWS service controller code"
	appLongDesc  = `ack-generate

A tool to generate AWS service controller code`
)

var (
	version            string
	buildHash          string
	buildDate          string
	defaultTemplateDir string
	templateDir        string
)

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: appShortDesc,
	Long:  appLongDesc,
}

func init() {
	cd, err := os.Getwd()
	if err != nil {
		fmt.Printf("unable to determine current working directory: %s\n", err)
		os.Exit(1)
	}
	tryPaths := []string{
		filepath.Join(cd, "templates"),
		filepath.Join(cd, "..", "templates"),
	}
	for _, tryPath := range tryPaths {
		if fi, err := os.Stat(tryPath); err == nil {
			if fi.IsDir() {
				defaultTemplateDir = tryPath
				break
			}
		}
	}
	rootCmd.PersistentFlags().StringVar(
		&templateDir, "template-dir", defaultTemplateDir, "Path to template directory",
	)
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute(v string, bh string, bd string) {
	version = v
	buildHash = bh
	buildDate = bd

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
