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

package resource

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/aws-service-operator-k8s/pkg/types"

	svcapitypes "github.com/aws/aws-service-operator-k8s/services/example/apis/v1alpha1"
)

// bookResourceManagerFactory produces bookResourceManager objects. It
// implements the `types.AWSResourceManagerFactory` interface.
type bookResourceManagerFactory struct {
	sync.RWMutex
	// rmCache contains resource managers for a particular AWS account ID
	rmCache map[types.AWSAccountID]*bookResourceManager
}

// ObjectPrototype returns the runtime.Object that resource managers produced
// by this factory handle
func (f *bookResourceManagerFactory) ObjectPrototype() runtime.Object {
	return &svcapitypes.Book{}
}

// GroupKind returns a string representation of the CRs handled by this
// resource manager
func (f *bookResourceManagerFactory) GroupKind() string {
	return "example.services.k8s.aws:Book"
}

// For returns a resource manager object that can manage resources for a
// supplied AWS account
func (f *bookResourceManagerFactory) For(id types.AWSAccountID) (types.AWSResourceManager, error) {
	f.RLock()
	rm, found := f.rmCache[id]
	f.RUnlock()

	if found {
		return rm, nil
	}

	f.Lock()
	defer f.Unlock()

	rm, err := newBookResourceManager(id)
	if err != nil {
		return nil, err
	}
	f.rmCache[id] = rm
	return rm, nil
}

func init() {
	RegisterManagerFactory(&bookResourceManagerFactory{})
}
