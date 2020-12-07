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

package crossplane

import (
	"bytes"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/generate/templateset"
)

type Initializer func(g *generate.Generator, ts *templateset.TemplateSet, meta templateset.MetaVars) error

type GenerationOption func(*Generation)

func WithInitializer(i Initializer) GenerationOption {
	return func(g *Generation) {
		g.Initializers = append(g.Initializers, i)
	}
}

func NewGeneration(g *generate.Generator, ts *templateset.TemplateSet, opts ...GenerationOption) *Generation {
	generation := &Generation{Generator: g, TemplateSet: ts}
	for _, f := range opts {
		f(generation)
	}
	return generation
}

type Generation struct {
	TemplateSet *templateset.TemplateSet
	Generator   *generate.Generator

	Initializers []Initializer
}

// Run starts the generation flow.
func (g *Generation) Run() (map[string]*bytes.Buffer, error) {
	meta := g.Generator.MetaVars()
	for _, init := range g.Initializers {
		if err := init(g.Generator, g.TemplateSet, meta); err != nil {
			return nil, err
		}
	}
	if err := g.TemplateSet.Execute(); err != nil {
		return nil, err
	}
	return g.TemplateSet.Executed(), nil
}
