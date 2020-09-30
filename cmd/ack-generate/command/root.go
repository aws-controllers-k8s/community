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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const (
	appName      = "ack-generate"
	appShortDesc = "ack-generate - generate AWS service controller code"
	appLongDesc  = `ack-generate

A tool to generate AWS service controller code`
)

var (
	version                string
	buildHash              string
	buildDate              string
	defaultCacheDir        string
	optCacheDir            string
	optRefreshCache        bool
	defaultTemplatesDir    string
	optTemplatesDir        string
	defaultServicesDir     string
	optServicesDir         string
	optDryRun              bool
	sdkDir                 string
	optGeneratorConfigPath string
)

var rootCmd = &cobra.Command{
	Use:          appName,
	Short:        appShortDesc,
	Long:         appLongDesc,
	SilenceUsage: true,
}

func init() {
	cd, err := os.Getwd()
	if err != nil {
		fmt.Printf("unable to determine current working directory: %s\n", err)
		os.Exit(1)
	}

	hd, err := homedir.Dir()
	if err != nil {
		fmt.Printf("unable to determine $HOME: %s\n", err)
		os.Exit(1)
	}
	defaultCacheDir = filepath.Join(hd, ".cache", appName)

	// try to determine a default template and services directory. If the call
	// is executing `ack-generate` via a checked-out ACK source repository,
	// then the templates and services directory in the source repo can serve
	// as sensible defaults
	tryPaths := []string{
		filepath.Join(cd, "templates"),
		filepath.Join(cd, "..", "templates"),
	}
	for _, tryPath := range tryPaths {
		if fi, err := os.Stat(tryPath); err == nil {
			if fi.IsDir() {
				defaultTemplatesDir = tryPath
				break
			}
		}
	}
	tryPaths = []string{
		filepath.Join(cd, "services"),
		filepath.Join(cd, "..", "services"),
	}
	for _, tryPath := range tryPaths {
		if fi, err := os.Stat(tryPath); err == nil {
			if fi.IsDir() {
				defaultServicesDir = tryPath
				break
			}
		}
	}
	rootCmd.PersistentFlags().BoolVar(
		&optDryRun, "dry-run", false, "If true, outputs all files to stdout",
	)
	rootCmd.PersistentFlags().StringVar(
		&optTemplatesDir, "templates-dir", defaultTemplatesDir, "Path to directory with templates to use in code generation",
	)
	rootCmd.PersistentFlags().StringVar(
		&optServicesDir, "services-dir", defaultServicesDir, "Path to directory to output service-specific code",
	)
	rootCmd.PersistentFlags().StringVar(
		&optCacheDir, "cache-dir", defaultCacheDir, "Path to directory to store cached files (including clone'd aws-sdk-go repo)",
	)
	rootCmd.PersistentFlags().BoolVar(
		&optRefreshCache, "refresh-cache", true, "If true, and aws-sdk-go repo is already cloned, will git pull the latest aws-sdk-go commit",
	)
	rootCmd.PersistentFlags().StringVar(
		&optGeneratorConfigPath, "generator-config-path", "", "Path to file containing instructions for code generation to use",
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
		os.Exit(1)
	}
}
