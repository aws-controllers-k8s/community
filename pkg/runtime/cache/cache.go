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

package cache

import (
	"os"
	"time"

	"github.com/go-logr/logr"
	kubernetes "k8s.io/client-go/kubernetes"
)

const (
	// defaultNamespace is the default namespace to use if the environment
	// variable NAMESPACE is not found. The NAMESPACE variable is injected
	// using the kubernetes downward api.
	defaultNamespace = "ack-system"

	// informerDefaultResyncPeriod is the period at which ShouldResync
	// is considered.
	informerResyncPeriod = 0 * time.Second
)

// currentNamespace is the namespace in which the current service
// controller Pod is running
var currentNamespace string

func init() {
	currentNamespace = os.Getenv("K8S_NAMESPACE")
	if currentNamespace == "" {
		currentNamespace = defaultNamespace
	}
}

// Caches is used to interact with the different caches
type Caches struct {
	// stopCh is a channel use to stop all the
	// owned caches
	stopCh chan struct{}

	// Accounts cache
	Accounts *AccountCache

	// Namespaces cache
	Namespaces *NamespaceCache
}

// New creates a new Caches object from a kubernetes.Interface and
// a logr.Logger
func New(clientset kubernetes.Interface, log logr.Logger) Caches {
	return Caches{
		Accounts:   NewAccountCache(clientset, log),
		Namespaces: NewNamespaceCache(clientset, log),
	}
}

// Run runs all the owned caches
func (c Caches) Run() {
	stopCh := make(chan struct{})
	if c.Accounts != nil {
		c.Accounts.Run(stopCh)
	}
	if c.Namespaces != nil {
		c.Namespaces.Run(stopCh)
	}
	c.stopCh = stopCh
}

// Stop closes the stop channel and cause all the SharedInformers
// by caches to stop running
func (c Caches) Stop() {
	close(c.stopCh)
}
