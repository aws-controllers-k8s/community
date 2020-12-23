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

package ack

import (
	"strings"
	ttpl "text/template"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/generate/templateset"
)

var (
	releaseTemplatePaths = []string{
		"helm/Chart.yaml.tpl",
		"helm/values.yaml.tpl",
		"helm/templates/role-reader.yaml.tpl",
		"helm/templates/role-writer.yaml.tpl",
	}
	releaseIncludePaths = []string{}
	releaseCopyPaths    = []string{
		"helm/templates/_helpers.tpl",
		"helm/templates/cluster-role-binding.yaml",
		"helm/templates/deployment.yaml",
		"helm/templates/service-account.yaml",
	}
	releaseFuncMap = ttpl.FuncMap{
		"ToLower": strings.ToLower,
		"Empty": func(subject string) bool {
			return strings.TrimSpace(subject) == ""
		},
	}
)

// Release returns a pointer to a TemplateSet containing all the templates for
// generating an ACK service controller release (Helm artifacts, etc)
func Release(
	g *generate.Generator,
	templateBasePath string,
	// releaseVersion is the SemVer string describing the release that the Helm
	// chart will install
	releaseVersion string,
	// imageRepository is the Docker image repository to use when generating
	// release files
	imageRepository string,
	// serviceAccountName is the name of the ServiceAccount and ClusterRole
	// used in the Helm chart
	serviceAccountName string,
) (*templateset.TemplateSet, error) {
	ts := templateset.New(
		templateBasePath,
		apisIncludePaths,
		apisCopyPaths,
		apisFuncMap,
	)

	metaVars := g.MetaVars()
	releaseVars := &templateReleaseVars{
		metaVars,
		releaseVersion,
		imageRepository,
		serviceAccountName,
	}
	for _, path := range releaseTemplatePaths {
		outPath := strings.TrimSuffix(path, ".tpl")
		if err := ts.Add(outPath, path, releaseVars); err != nil {
			return nil, err
		}
	}

	return ts, nil
}

// templateReleaseVars contains template variables for the template that
// outputs Go code for a release artifact
type templateReleaseVars struct {
	templateset.MetaVars
	// ReleaseVersion is the semver release tag (or Git SHA1 commit) that is
	// used for the binary image artifacts and Helm release version
	ReleaseVersion string
	// ImageRepository is the Docker image repository to inject into the Helm
	// values template
	ImageRepository string
	// ServiceAccountName is the name of the service account and cluster role
	// created by the Helm chart
	ServiceAccountName string
}
