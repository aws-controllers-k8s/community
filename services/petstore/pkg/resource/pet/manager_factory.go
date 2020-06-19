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

package pet

import (
	"sync"

	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"
	svcresource "github.com/aws/aws-service-operator-k8s/services/petstore/pkg/resource"
)

// petResourceManagerFactory produces petResourceManager objects. It
// implements the `types.AWSResourceManagerFactory` interface.
type petResourceManagerFactory struct {
	sync.RWMutex
	// rmCache contains resource managers for a particular AWS account ID
	rmCache map[ackv1alpha1.AWSAccountID]*petResourceManager
}

// ResourcePrototype returns an AWSResource that resource managers produced by
// this factory will handle
func (f *petResourceManagerFactory) ResourceDescriptor() acktypes.AWSResourceDescriptor {
	return &petResourceDescriptor{}
}

// GroupKind returns a string representation of the CRs handled by this
// resource manager
func (f *petResourceManagerFactory) GroupKind() string {
	return "example.services.k8s.aws:Pet"
}

// ManagerFor returns a resource manager object that can manage resources for a
// supplied AWS account
func (f *petResourceManagerFactory) ManagerFor(
	id ackv1alpha1.AWSAccountID,
) (acktypes.AWSResourceManager, error) {
	f.RLock()
	rm, found := f.rmCache[id]
	f.RUnlock()

	if found {
		return rm, nil
	}

	f.Lock()
	defer f.Unlock()

	rm, err := newPetResourceManager(id)
	if err != nil {
		return nil, err
	}
	f.rmCache[id] = rm
	return rm, nil
}

func newPetResourceManagerFactory() *petResourceManagerFactory {
	return &petResourceManagerFactory{
		rmCache: map[ackv1alpha1.AWSAccountID]*petResourceManager{},
	}
}

func init() {
	svcresource.RegisterManagerFactory(newPetResourceManagerFactory())
}
