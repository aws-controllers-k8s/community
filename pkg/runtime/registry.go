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

package runtime

import (
	"sync"

	"github.com/aws/aws-service-operator-k8s/pkg/types"
)

type Registry struct {
	sync.RWMutex
	// resourceManagerFactories is a map of resource manager factories, keyed
	// by the GroupKind of the resource managed by the resource manager
	// produced by that factory
	resourceManagerFactories map[string]types.AWSResourceManagerFactory
}

// GetResourceManagerFactories returns AWSResourceManagerFactories that are
// registered with the RegistryA
func (r *Registry) GetResourceManagerFactories() []types.AWSResourceManagerFactory {
	r.Lock()
	defer r.Unlock()
	res := make([]types.AWSResourceManagerFactory, 0, len(r.resourceManagerFactories))
	for _, rmf := range r.resourceManagerFactories {
		res = append(res, rmf)
	}
	return res
}

// RegisterManagerFactory registers a resource manager factory with the
// package's registry
func (r *Registry) RegisterResourceManagerFactory(f types.AWSResourceManagerFactory) {
	r.Lock()
	defer r.Unlock()
	r.resourceManagerFactories[f.ResourceDescriptor().GroupKind().String()] = f
}

// NewRegistry retuns a thread-safe Registry object
func NewRegistry() *Registry {
	return &Registry{
		resourceManagerFactories: map[string]types.AWSResourceManagerFactory{},
	}
}
