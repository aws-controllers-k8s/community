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
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

const (
	sdkRepoURL = "https://github.com/aws/aws-sdk-go"
)

// ensureDir makes sure that a supplied directory exists and
// returns whether the directory already existed.
func ensureDir(fp string) (bool, error) {
	fi, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, os.MkdirAll(fp, os.ModePerm)
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("expected %s to be a directory", fp)
	}
	if !isDirWriteable(fp) {
		return true, fmt.Errorf("%s is not a writeable directory", fp)
	}

	return true, nil
}

// isDirWriteable returns true if the supplied directory path is writeable,
// false otherwise
func isDirWriteable(fp string) bool {
	testPath := filepath.Join(fp, "test")
	f, err := os.Create(testPath)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testPath)
	return true
}

// ensureSDKRepo ensures that we have a git clone'd copy of the aws-sdk-go
// repository, which we use model JSON files from. Upon successful return of
// this function, the sdkDir global variable will be set to the directory where
// the aws-sdk-go is found
func ensureSDKRepo(cacheDir string) error {
	var err error
	srcPath := filepath.Join(cacheDir, "src")
	if err = os.MkdirAll(srcPath, os.ModePerm); err != nil {
		return err
	}
	// clone the aws-sdk-go repository locally so we can query for API
	// information in the models/apis/ directories
	sdkDir, err = cloneSDKRepo(srcPath)
	return err
}

// cloneSDKRepo git clone's the aws-sdk-go source repo into the cache and
// returns the filepath to the clone'd repo. If the aws-sdk-go repository
// already exists in the cache, it will checkout the current sdk-go version
// mentionned in 'go.mod' file.
func cloneSDKRepo(srcPath string) (string, error) {
	clonePath := filepath.Join(srcPath, "aws-sdk-go")

	// since early versions of ack-generate used to clone aws-sdk-go repo using
	// 'git clone --depth 1', right now 'git checkout' doesn't run successfully
	// on those repositories. To solve this issue, we temporarilly add this code
	// to delete old cache versions.
	// TODO(hilalymh) remove this block Q1 2021
	if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
		err := removeOldRepoCache(clonePath)
		if err != nil {
			return "", err
		}
	}

	sdkVersion, err := getSDKVersion()
	if err != nil {
		return "", err
	}
	if optRefreshCache {
		if _, err := os.Stat(filepath.Join(clonePath, ".git")); !os.IsNotExist(err) {
			cmd := exec.Command("git", "-C", clonePath, "checkout", "tags/"+sdkVersion)
			return clonePath, cmd.Run()
		}
	}
	if _, err := os.Stat(clonePath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "-b", sdkVersion, sdkRepoURL, clonePath)
		return clonePath, cmd.Run()
	}
	return clonePath, nil
}

// removeOldRepoCache deletes old aws-sdk-go repositories. It deletes a
// repository if the number of commits is equal to 0
func removeOldRepoCache(repoDir string) error {
	b, err := exec.Command("git", "-C", repoDir, "rev-list", "--count", "HEAD").Output()
	if err != nil {
		return err
	}
	count, err := strconv.Atoi(strings.TrimSuffix(string(b), "\n"))
	if err != nil {
		return err
	}
	if count == 1 {
		err := os.RemoveAll(repoDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// getSDKVersion parses the go.mod file and returns aws-sdk-go version
func getSDKVersion() (string, error) {
	b, err := ioutil.ReadFile("./go.mod")
	if err != nil {
		return "", err
	}
	goMod, err := modfile.Parse("", b, nil)
	if err != nil {
		return "", err
	}
	sdkModule := strings.TrimPrefix(sdkRepoURL, "https://")
	for _, require := range goMod.Require {
		if require.Mod.Path == sdkModule {
			return require.Mod.Version, nil
		}
	}
	return "", fmt.Errorf("couldn't find %s in the go.mod require block", sdkModule)
}
