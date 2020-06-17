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

package apis

import (
	"io/ioutil"
	"path/filepath"
	ttpl "text/template"
)

type DocTemplateVars struct {
	APIVersion string
	APIGroup   string
}

func NewDocTemplate(tplDir string) (*ttpl.Template, error) {
	tplPath := filepath.Join(tplDir, "apis", "doc.go.tpl")
	tplContents, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return nil, err
	}
	t := ttpl.New("doc")
	return t.Parse(string(tplContents))
}
